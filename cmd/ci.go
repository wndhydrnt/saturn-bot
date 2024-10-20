package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
)

var (
	ciCommandHelp = `Loads and validates task files.

"ci" can be executed during a continuous integration process
to receive early feedback on syntax errors or invalid values
in task files.

It also starts any plugins defined in a task file
and calls their initialize function.
Pass --start-plugins=false to prevent this.

Examples:

# Validate one task file
saturn-bot ci ./task.yaml

# Validate multiple task files
saturn-bot ci ./*.yaml
`
)

func createCiCommand() *cobra.Command {
	var skipPlugins bool

	cmd := &cobra.Command{
		Use:   "ci FILE [FILE...]",
		Short: "Loads and validates task files",
		Long:  ciCommandHelp,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runner, err := command.NewCiRunnerFromConfig(cfgFile, skipPlugins)
			handleError(err, cmd.ErrOrStderr())
			err = runner.Run(cmd.OutOrStdout(), args...)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	cmd.Flags().BoolVar(&skipPlugins, "skip-plugins", false, "Skip starting plugins as part of the CI run.")
	return cmd
}
