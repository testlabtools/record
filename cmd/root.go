package cmd

import (
	"log/slog"
	"os"
	"time"

	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Root represents the base command when called without any subcommands
var Root = &cobra.Command{
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

	Root.PersistentFlags().String("repo", "", "path to git repo")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
