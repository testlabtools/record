package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
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

	if err := record(l); err != nil {
		sentry.CaptureException(err)
		sentry.Flush(5 * time.Second)
		l.Error("failed to run", "err", err)
	}
}

func record(l *slog.Logger) error {
	files := map[string][]byte{
		"file1.txt": []byte("This is the content of file1."),
		"file2.txt": []byte("This is the content of file2."),
	}

	var raw bytes.Buffer
	var buf bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for name, content := range files {
		l.Debug("add tar file", "name", name, "size", len(content))
	}

	if err := createTarball(files, &raw); err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	if err := compressZstd(&raw, &buf); err != nil {
		return fmt.Errorf("failed to compress tarball: %w", err)
	}

	l.Info("tarball compressed",
		"files", len(files),
		"rawSize", raw.Len(),
		"compressedSize", buf.Len(),
	)

	// TODO
	url := "https://example.com/upload"
	if err := uploadFile(ctx, url, &buf); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	l.Info("upload successful", "url", url)
	return nil
}
