package cmd

import (
	"github.com/spf13/cobra"
)

func createExperimentalCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "experimental",
		Short: "Commands that are not ready for use in production",
		Long:  "Commands that are not ready for use in production.",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(
		createServerCommand(),
		createWorkerCommand(),
	)
	return cmd
}
