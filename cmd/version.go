package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
)

func createVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  "Display version information.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", version.String())
		},
	}
	return cmd
}
