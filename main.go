package main

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/lmittmann/tint"
)

func main() {
	l := slog.New(
		Fanout(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: time.Kitchen,
			}),
			sentryslog.Option{
				Level:     slog.LevelDebug,
				AddSource: true,
			}.NewSentryHandler(),
		),
	)

	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://3c2b29267e8ad655b3f95b02cfb05627@o4508645503926272.ingest.de.sentry.io/4508652922536016",
	})
	if err != nil {
		l.Error("failed to initialize sentry", "err", err)
		os.Exit(1)
	}

	// Flush buffered events before the program terminates.
	defer func() {
		err := recover()

		if err != nil {
			sentry.CurrentHub().Recover(err)
			sentry.Flush(5 * time.Second)
			l.Error("failed to run", "err", err)
		}
	}()

	// Extract environment variables from OS.
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) > 1 {
			env[pair[0]] = pair[1]
		}
	}

	if err := upload(l, env); err != nil {
		sentry.CaptureException(err)
		sentry.Flush(5 * time.Second)
		l.Error("failed to run", "err", err)
	}
}
