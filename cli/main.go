package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/testlabtools/record/cmd"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func main() {
	l := slog.Default()

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
			os.Exit(1)
		}
	}()

	if err := cmd.Root.Execute(); err != nil {
		sentry.CaptureException(err)
		sentry.Flush(5 * time.Second)
		l.Error("failed to run", "err", err)
		os.Exit(1)
	}
}
