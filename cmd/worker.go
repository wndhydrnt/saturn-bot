package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/worker"
)

var (
	workerCommandHelp = `Starts the worker component.

"worker" queries the server component for tasks to execute,
executes them and reports the results back to the server.

Examples:

# Start the worker
saturn-bot worker --config config.yaml ./tasks/**/*.yaml
`
)

func createWorkerCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "worker FILE [FILE...]",
		Short: "Starts the worker component",
		Long:  workerCommandHelp,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := worker.Run(cfgFile, args)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	return cmd
}
