package record

import (
	"io"
	"log/slog"
	"os"
)

type PredictOptions struct {
	Repo string

	Runner string

	Debug bool
}

func Predict(l *slog.Logger, env map[string]string, o PredictOptions) error {
	// Copy stdin to stdout for now.
	_, err := io.Copy(os.Stdout, os.Stdin)
	return err
}
