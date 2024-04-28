package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg"
)

func createRunCommand() *cobra.Command {
	var taskFiles []string

	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Run locally",
		Long:  "Run locally",
		Run: func(cmd *cobra.Command, args []string) {
			err := pkg.ExecuteRun(cfgFile, taskFiles)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	cmd.Flags().StringArrayVar(&taskFiles, "task", []string{}, "")
	return cmd
}
