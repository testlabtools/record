package main

import (
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/testlabtools/record"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload CI and test run results to Test Lab",
	// Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:
	//
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := slog.Default()
		env := getEnv()
		if val := cmd.Context().Value("env"); val != nil {
			env = val.(map[string]string)
		}

		o := record.UploadOptions{
			Repo:    cmd.Flag("repo").Value.String(),
			Reports: cmd.Flag("reports").Value.String(),
		}

		started := cmd.Flag("started").Value.String()
		if started != "" {
			val, err := parseStarted(started)
			if err != nil {
				return err
			}
			o.Started = &val
		}

		return record.Upload(l, env, o)
	},
}

func parseStarted(s string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05-0700", s)
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	uploadCmd.Flags().String("started", "", "set run's start time (ISO 8601 format)")

	uploadCmd.Flags().String("reports", "", "path to the JUnit reports directory")
}
