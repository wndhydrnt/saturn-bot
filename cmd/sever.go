package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/server"
)

func createServerCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "server FILE [FILE...]",
		Short: "Start the server",
		Long:  "Start the server.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := server.Run(cfgFile, args)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	return cmd
}
