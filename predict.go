package record

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/runner"
)

type PredictOptions struct {
	Repo string

	WorkDir string

	Runner string

	Debug bool

	Stdin  io.Reader
	Stdout io.Writer
}

func Predict(l *slog.Logger, env map[string]string, o PredictOptions) error {
	po := runner.ParserOptions{
		WorkDir: o.WorkDir,
	}

	run, err := runner.New(o.Runner, po)
	if err != nil {
		return err
	}

	if err := run.Parse(o.Stdin); err != nil {
		return fmt.Errorf("failed to parse stdin for format %q: %w", o.Runner, err)
	}

	predicted, err := predict(l, env, o, run)
	if err != nil {
		l.Error("failed to predict", "err", err)

		// Fallback to original input.
		l.Warn("fallback to original test input", "files", len(run.Files()))
		predicted = run.Files()
	}

	return run.Format(predicted, o.Stdout)
}

func predict(l *slog.Logger, osEnv map[string]string, o PredictOptions, input runner.Parser) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	server := osEnv["TESTLAB_HOST"]
	if server == "" {
		server = "https://eu.testlab.tools"
	}

	apiKey := osEnv["TESTLAB_KEY"]
	if apiKey == "" {
		return nil, fmt.Errorf("env var TESTLAB_KEY is required")
	}

	files := input.Files()

	if len(files) == 0 {
		l.Warn("no test files read from stdin")
		return files, nil
	}

	collector, err := NewCollector(l, o.Repo, osEnv)
	if err != nil {
		return nil, err
	}

	env := collector.Env()
	l.Debug("collected env vars", "env", env)

	l.Info("upload predict request", "server", server, "apiKey", mask(apiKey), "files", len(files))

	api, err := newApi(l, server, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize api: %w", err)
	}

	summary, err := collector.gitSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get git summary: %w", err)
	}

	var ds *client.GitDiffStat
	if summary != nil {
		var changes []client.GitFileChange
		for _, c := range summary.DiffStat.Changes {
			changes = append(changes, client.GitFileChange{
				Name:       c.Name,
				Insertions: c.Insertions,
				Deletions:  c.Deletions,
			})
		}

		ds = &client.GitDiffStat{
			Hash:       summary.DiffStat.Hash,
			Changes:    changes,
			Files:      summary.DiffStat.Files,
			Insertions: summary.DiffStat.Insertions,
			Deletions:  summary.DiffStat.Deletions,
		}
	}

	var testFiles []client.PredictTestFile
	for _, f := range files {
		testFiles = append(testFiles, client.PredictTestFile{
			Path: f,
		})
	}

	req := client.PredictRequest{
		CiRun: env.RunRequest(),
		GitSummary: client.GitSummary{
			DiffStat: ds,
		},
		TestFiles: testFiles,
	}
	predicted, err := api.predictTests(ctx, req)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, file := range predicted.TestFiles {
		out = append(out, file.Path)
	}

	return out, nil
}

// predictTests predicts what tests to run for a CI run.
func (u *api) predictTests(ctx context.Context, body client.PredictRequest) (*client.PredictResponse, error) {
	params := &client.PredictTestsParams{
		// TODO add zstd compression.
	}
	predict, err := u.api.PredictTestsWithResponse(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("failed to predict tests: %w", err)
	}

	if status := predict.StatusCode(); status != http.StatusOK {
		return nil, fmt.Errorf("predict tests returned invalid status code: %d", status)
	}

	return predict.JSON200, nil
}
