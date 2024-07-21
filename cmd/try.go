package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
)

func createTryCommand() *cobra.Command {
	var dataDir string
	var repository string
	var taskFile string
	var taskName string

	cmd := &cobra.Command{
		Use:   "try",
		Short: "Try out a task locally",
		Long:  "Try out a task locally",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Read(cfgFile)
			handleError(err, cmd.ErrOrStderr())
			opts, err := options.ToOptions(cfg)
			handleError(err, cmd.ErrOrStderr())
			runner, err := command.NewTryRunner(opts, dataDir, repository, taskFile, taskName)
			if err != nil {
				handleError(err, cmd.ErrOrStderr())
			}

			err = runner.Run()
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "path to config file.")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "path to directory to clone the repository.")
	cmd.Flags().StringVar(&repository, "repository", "", "name of the repository to test against.")
	cmd.Flags().StringVar(&taskFile, "task-file", "", "path to the task file to try out.")
	cmd.Flags().StringVar(&taskName, "task-name", "", "if set, try only the task that matches the name. useful if a task file contains multiple tasks.")
	return cmd
}
