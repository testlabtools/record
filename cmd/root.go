package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "record",
	Short: "Manage CI and test runs in Test Lab",
	// Long: `A longer description that spans multiple lines and likely contains
	// examples and usage of using your application. For example:
	//
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },

	// Dont show CLI usage on error.
	SilenceUsage:  true,
	SilenceErrors: true,
}

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

	if err := rootCmd.Execute(); err != nil {
		sentry.CaptureException(err)
		sentry.Flush(5 * time.Second)
		l.Error("failed to run", "err", err)
		os.Exit(1)
	}
}

func setLogLevel(level slog.Level) {
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
	slog.SetDefault(l)
}

func init() {
	setLogLevel(slog.LevelInfo)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().String("repo", "", "path to git repo")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
