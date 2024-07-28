package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
)

var (
	tryCommandHelp = `Try out a task locally.

"try" verifies that all filters and that actions modify files
in a repository.

It first executes all filters against the given repository and
provides feedback on whether they match or not.
If all filters match, it clones the repository, applies all
actions and provides feedback on whether files have changed or not.

Use this command during local development of a task to try it out
and iterate frequently.

Examples:

Try all tasks in file "task.yaml" against
repository "github.com/wndhydrnt/saturn-bot-example".

saturn-bot try \
  --config config.yaml \
  --repository github.com/wndhydrnt/saturn-bot-example \
  --task-file task.yaml

Try task "example" in "task.yaml" against
repository "github.com/wndhydrnt/saturn-bot-example".

saturn-bot try \
  --config config.yaml \
  --repository github.com/wndhydrnt/saturn-bot-example \
  --task-file task.yaml \
  --task-name example`
)

func createTryCommand() *cobra.Command {
	var dataDir string
	var repository string
	var taskFile string
	var taskName string

	cmd := &cobra.Command{
		Use:   "try",
		Short: "Try out a task locally",
		Long:  tryCommandHelp,
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
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file.")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Path to directory to clone the repository.")
	cmd.Flags().StringVar(&repository, "repository", "", "Name of the repository to test against.")
	cmd.Flags().StringVar(&taskFile, "task-file", "", "Path to the task file to try out.")
	cmd.Flags().StringVar(&taskName, "task-name", "", `If set, try only the task that matches the name.
Useful if a task file contains multiple tasks.`)
	return cmd
}
