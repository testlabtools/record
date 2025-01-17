package record

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/tar"
	"github.com/testlabtools/record/zstd"
)

type Collector struct {
	log     *slog.Logger
	options UploadOptions
}

func NewCollector(l *slog.Logger, o UploadOptions) *Collector {
	return &Collector{
		log:     l,
		options: o,
	}
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

func (c *Collector) Env(env map[string]string) (RunEnv, error) {
	group := env["TESTLAB_GROUP"]
	if group == "" {
		return RunEnv{}, fmt.Errorf("env var TESTLAB_GROUP is required")
	}

	if env["GITHUB_ACTIONS"] != "" {
		ciEnv := make(map[string]interface{})
		extra := []string{
			"GITHUB_BASE_REF",
			"GITHUB_HEAD_REF",
			"GITHUB_JOB",
			"GITHUB_REF_TYPE",
		}

		for _, key := range extra {
			val := env[key]
			if val == "" {
				continue
			}
			ciEnv[key] = val
		}

		re := RunEnv{
			ActorName:      env["GITHUB_ACTOR"],
			CIProviderName: client.Github,
			GitRef:         env["GITHUB_REF"],
			GitRefName:     env["GITHUB_REF_NAME"],
			GitRepo:        env["GITHUB_REPOSITORY"],
			GitSha:         env["GITHUB_SHA"],
			Group:          group,
			CIEnv:          &ciEnv,
		}

		numeric := map[string]*int{
			"GITHUB_RUN_ATTEMPT": &re.RunAttempt,
			"GITHUB_RUN_ID":      &re.RunId,
			"GITHUB_RUN_NUMBER":  &re.RunNumber,
		}

		err := parseInts(env, numeric)

		return re, err
	}

	return RunEnv{}, fmt.Errorf("unknown CI provider")
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

func readDir(dir string) (map[string][]byte, error) {
	if dir == "" || !dirExists(dir) {
		return nil, nil
	}

	files := make(map[string][]byte)

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

		// Store the file content using the full path
		files[path] = content
		return nil
	})

	return files, err
}

func (c *Collector) Bundle(initial bool, w io.Writer) error {
	dir := c.options.Reports
	files, err := readDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read reports (%q): %w", dir, err)
	}

	// TODO if initial run: add codeowners, git data, etc.

	if len(files) == 0 {
		c.log.Warn("no files found for bundle", "reports", dir)
		return nil
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
