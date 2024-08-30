package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/worker"
)

func createWorkerCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "worker FILE [FILE...]",
		Short: "Start the worker",
		Long:  "Start the worker.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := worker.Run(cfgFile, args)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	return cmd
}
