package record

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/testlabtools/record/client"
)

type Collector struct {
	log *slog.Logger
}

func NewCollector(l *slog.Logger) *Collector {
	return &Collector{
		log: l,
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

func (c *Collector) Bundle(initial bool, w io.Writer) error {
	files := map[string][]byte{
		"file1.txt": []byte("This is the content of file1."),
		"file2.txt": []byte("This is the content of file2."),
	}

	// TODO if initial run: add codeowners, git data, etc.

	for name, content := range files {
		c.log.Debug("add tar file", "name", name, "size", len(content))
	}

	var raw bytes.Buffer
	if err := createTarball(files, &raw); err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	c.log.Info("tarball created",
		"files", len(files),
		"rawSize", raw.Len(),
	)

	if err := compressZstd(&raw, w); err != nil {
		return fmt.Errorf("failed to compress tarball: %w", err)
	}

	return nil
}
