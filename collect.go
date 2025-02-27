package record

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/git"
	"github.com/testlabtools/record/tar"
	"github.com/testlabtools/record/zstd"
)

type Collector struct {
	log  *slog.Logger
	repo *git.Repo

	env   RunEnv
	osEnv map[string]string
}

func NewCollector(l *slog.Logger, repo string, osEnv map[string]string) (*Collector, error) {
	r := git.NewRepo(repo)

	c := &Collector{
		log:  l,
		repo: r,

		osEnv: osEnv,
	}

	env, err := c.initEnv()
	c.env = env

	return c, err
}

func parseInts(env map[string]string, numeric map[string]*int) error {
	for key, ref := range numeric {
		val, err := strconv.Atoi(env[key])
		if err != nil {
			return fmt.Errorf("failed to parse %q: %s", key, err)
		}
		*ref = val
	}

	return nil
}

type RunEnv struct {
	ActorName      string
	CIProviderName client.CIProviderName
	GitRef         string
	GitRefName     string
	GitRepo        string
	GitSha         string
	Group          string
	RunAttempt     int
	RunId          int
	RunNumber      int
	CIEnv          *map[string]interface{}
}

func (env RunEnv) RunRequest() client.CIRunRequest {
	return client.CIRunRequest{
		ActorName:      env.ActorName,
		CiProviderName: env.CIProviderName,
		GitRef:         env.GitRef,
		GitRefName:     env.GitRefName,
		GitRepo:        env.GitRepo,
		GitSha:         env.GitSha,
		Group:          env.Group,
		RunAttempt:     env.RunAttempt,
		RunId:          env.RunId,
		RunNumber:      env.RunNumber,
		CiEnv:          env.CIEnv,
	}
}

func (c *Collector) collectGitEnv() (map[string]interface{}, error) {
	env := make(map[string]interface{})

	if !c.repo.Exists() {
		c.log.Warn("cannot get git env since git repo does not exist", "dir", c.repo.Dir)
		return env, nil
	}

	ref := "HEAD"
	tags, err := c.repo.TagsPointedAt(ref)
	if err != nil {
		return env, fmt.Errorf("failed to get tags pointed at ref %q: %w", ref, err)
	}

	if len(tags) > 0 {
		env["GIT_TAGS_POINTED_AT"] = strings.Join(tags, ";")
	}

	info, err := c.repo.CommitInfo(ref)
	if err != nil {
		return env, fmt.Errorf("failed to get commit info at ref %q: %w", ref, err)
	}
	env["GIT_COMMIT_AUTHOR_EMAIL"] = info.AuthorEmail
	env["GIT_COMMIT_SUBJECT"] = info.Subject

	return env, nil
}

func (c *Collector) initEnv() (RunEnv, error) {
	group := c.osEnv["TESTLAB_GROUP"]
	if group == "" {
		return RunEnv{}, fmt.Errorf("env var TESTLAB_GROUP is required")
	}

	ciEnv, err := c.collectGitEnv()
	if err != nil {
		return RunEnv{}, fmt.Errorf("failed to collect git env: %w", err)
	}

	if c.osEnv["GITHUB_ACTIONS"] != "" {
		extra := []string{
			"GITHUB_BASE_REF",
			"GITHUB_HEAD_REF",
			"GITHUB_JOB",
			"GITHUB_REF_TYPE",
		}

		for _, key := range extra {
			val := c.osEnv[key]
			if val == "" {
				continue
			}
			ciEnv[key] = val
		}

		re := RunEnv{
			ActorName:      c.osEnv["GITHUB_ACTOR"],
			CIProviderName: client.Github,
			GitRef:         c.osEnv["GITHUB_REF"],
			GitRefName:     c.osEnv["GITHUB_REF_NAME"],
			GitRepo:        c.osEnv["GITHUB_REPOSITORY"],
			GitSha:         c.osEnv["GITHUB_SHA"],
			Group:          group,
			CIEnv:          &ciEnv,
		}

		numeric := map[string]*int{
			"GITHUB_RUN_ATTEMPT": &re.RunAttempt,
			"GITHUB_RUN_ID":      &re.RunId,
			"GITHUB_RUN_NUMBER":  &re.RunNumber,
		}

		err := parseInts(c.osEnv, numeric)

		return re, err
	}

	return RunEnv{}, fmt.Errorf("unknown CI provider")
}

func (c *Collector) Env() RunEnv {
	return c.env
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func readReports(dir string, limit int) (map[string][]byte, error) {
	if dir == "" || !dirExists(dir) {
		return nil, nil
	}

	files := make(map[string][]byte)
	i := 0

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path %q: %w", path, err)
		}

		if d.IsDir() {
			// Skip directories
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %q: %w", path, err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", path, err)
		}

		// Store the file content using short file names
		i++
		name := fmt.Sprintf("reports/%d%s", i, filepath.Ext(path))
		files[name] = content

		if len(files) > limit {
			// Avoid bundling a whole repo.
			return fmt.Errorf("too many files (%d > %d) found", len(files), limit)
		}

		return nil
	})

	return files, err
}

func (c *Collector) findCodeOwners(dir string) string {
	// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners#codeowners-file-location
	names := []string{
		".github/CODEOWNERS",
		"CODEOWNERS",
		"docs/CODEOWNERS",
	}

	for _, name := range names {
		file := path.Join(dir, name)
		if fileExists(file) {
			return file
		}
	}
	return ""
}

func (c *Collector) addCodeOwners(files *map[string][]byte) error {
	repo := c.repo.Dir
	file := c.findCodeOwners(repo)

	if file == "" {
		c.log.Warn("missing CODEOWNERS", "file", file, "repo", repo)
		return nil
	}

	c.log.Info("found CODEOWNERS", "file", file, "repo", repo)

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", file, err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", file, err)
	}

	(*files)["CODEOWNERS"] = buf

	return nil
}

type GitSummary struct {
	DiffStat    *git.DiffStat    `json:"diffStat"`
	CommitFiles []git.CommitFile `json:"commitFiles"`
}

const GitSummaryFileName = "git.json"

func (c *Collector) addGitSummary(files *map[string][]byte) error {
	if !c.repo.Exists() {
		c.log.Warn("cannot get git summary since repo does not exist", "dir", c.repo.Dir)
		return nil
	}

	ds, err := c.repo.DiffStat("HEAD")
	if err != nil {
		return err
	}

	main, err := c.repo.MainBranch()
	if err != nil {
		return err
	}

	summary := GitSummary{
		DiffStat: ds,
	}

	c.log.Debug("compare git ref name with main branch",
		"refName", c.env.GitRefName,
		"main", main,
	)
	if c.env.GitRefName == main {
		cf, err := c.repo.CommitFiles()
		if err != nil {
			return err
		}
		summary.CommitFiles = cf
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(summary); err != nil {
		return err
	}
	(*files)[GitSummaryFileName] = buf.Bytes()

	return nil
}

type BundleOptions struct {
	InitialRun bool

	ReportsDir string
	MaxReports int
}

func (c *Collector) Bundle(o BundleOptions, w io.Writer) error {
	dir := o.ReportsDir

	maxReports := o.MaxReports
	if maxReports == 0 {
		maxReports = DefaulMaxReports
	}

	c.log.Debug("read file reports", "dir", dir, "max", maxReports)
	files, err := readReports(dir, maxReports)
	if err != nil {
		return fmt.Errorf("failed to read reports (%q): %w", dir, err)
	}

	if len(files) == 0 {
		c.log.Warn("no file reports found for bundle", "reports", dir)
		return nil
	}

	if o.InitialRun {
		// Add CODEOWNERS file to the initial run only. This avoids storing the
		// same information in each run bundle file.
		if err := c.addCodeOwners(&files); err != nil {
			return fmt.Errorf("failed to add CODEOWNERS: %w", err)
		}

		if err := c.addGitSummary(&files); err != nil {
			return fmt.Errorf("failed to add git summary: %w", err)
		}
	}

	for name, content := range files {
		c.log.Debug("add tar file", "name", name, "size", len(content))
	}

	var raw bytes.Buffer
	if err := tar.Create(files, &raw); err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	c.log.Info("tarball created",
		"files", len(files),
		"rawSize", raw.Len(),
	)

	if err := zstd.Compress(&raw, w); err != nil {
		return fmt.Errorf("failed to compress tarball: %w", err)
	}

	return nil
}
