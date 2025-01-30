package record

import "log/slog"

type PredictOptions struct {
	Repo string

	Runner string

	Debug bool
}

func Predict(l *slog.Logger, env map[string]string, o PredictOptions) error {
	return nil
}
