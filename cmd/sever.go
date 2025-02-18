package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/server"
)

var (
	serverCommandHelp = `Starts the server component.

"server" serves the API and the UI.

Examples:

# Start the server
saturn-bot server --config config.yaml ./tasks/**/*.yaml
`
)

func createServerCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "server FILE [FILE...]",
		Short: "Starts the server component",
		Long:  serverCommandHelp,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := server.Run(cfgFile, args)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	return cmd
}
