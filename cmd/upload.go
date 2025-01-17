package main

import (
	"log/slog"

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
	Run: func(cmd *cobra.Command, args []string) {
		l := slog.Default()
		env := getEnv()
		record.Upload(l, env, record.UploadOptions{
			Repo: repo,
		})
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
