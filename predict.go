package record

import (
	"fmt"
	"io"
	"log/slog"

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

	return run.Format(o.Stdout)
}
