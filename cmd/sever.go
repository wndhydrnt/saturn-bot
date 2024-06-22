package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/server"
)

func createServerCommand() *cobra.Command {
	var taskFiles []string

	var cmd = &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		Long:  "Start the server.",
		Run: func(cmd *cobra.Command, args []string) {
			err := server.Run()
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	cmd.Flags().StringArrayVar(&taskFiles, "task", []string{}, "")
	return cmd
}
