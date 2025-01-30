package cmd

import (
	"github.com/spf13/cobra"
	"github.com/testlabtools/record"
)

// predictCmd represents the predict command
var predictCmd = &cobra.Command{
	Use:   "predict",
	Short: "Predict CI and test results using TestLab",
	RunE: func(cmd *cobra.Command, args []string) error {
		setup := setupCommand(cmd, args)

		o := record.PredictOptions{
			Repo: cmd.Flag("repo").Value.String(),

			PredictedTestsFile: cmd.Flag("predicted-tests-file").Value.String(),

			Debug: setup.debug,
		}

		return record.Predict(setup.log, setup.env, o)
	},
}

func init() {
	Root.AddCommand(predictCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// predictCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// predictCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	predictCmd.Flags().String("predicted-tests-file", "/tmp/predicted-tests.json", "path to the predicted tests output file")
}
