package record

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

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

	predicted, err := predict(l, env, run)
	if err != nil {
		l.Error("failed to predict", "err", err)

		// Fallback to original input.
		l.Warn("fallback to original test input", "files", len(run.Files()))
		predicted = run.Files()
	}

	return run.Format(predicted, o.Stdout)
}

func predict(l *slog.Logger, osEnv map[string]string, input runner.Parser) ([]string, error) {
	_, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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
	l.Info("upload predict", "server", server, "apiKey", mask(apiKey), "files", len(files))

	if len(files) == 0 {
		l.Warn("no test files read from stdin")
		return files, nil
	}

	return files, nil
}
