package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/worker"
)

func createWorkerCommand() *cobra.Command {
	var taskFiles []string

	var cmd = &cobra.Command{
		Use:   "worker",
		Short: "Start the worker",
		Long:  "Start the worker.",
		Run: func(cmd *cobra.Command, args []string) {
			err := worker.Run()
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	cmd.Flags().StringArrayVar(&taskFiles, "task", []string{}, "")
	return cmd
}
