package record

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type PredictOptions struct {
	Repo string

	Runner string

	Debug bool

	Stdin  io.Reader
	Stdout io.Writer
}

func Predict(l *slog.Logger, env map[string]string, o PredictOptions) error {
	// Copy stdin to stdout for now.
	scanner := bufio.NewScanner(o.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// Omit any lines with spaces (containing go build output).
		if strings.Contains(line, " ") {
			continue
		}
		fmt.Fprintln(o.Stdout, line)
	}

	return scanner.Err()
}
