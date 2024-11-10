package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
)

var (
	tryCommandHelp = `Try out a task locally.

"try" verifies that all filters match and that actions modify files
in a repository.

Use this command during local development of a task to try it out
and iterate frequently.

It first executes all filters against the given repository and
provides feedback on whether they match or not.
If all filters match, it clones the repository, applies all
actions and provides feedback on whether files have changed or not.

Examples:

# Try all tasks in file "task.yaml" against
# repository "github.com/wndhydrnt/saturn-bot-example".
saturn-bot try \
  --repository github.com/wndhydrnt/saturn-bot-example \
  task.yaml

# Try task with name "example" in "task.yaml" against
# repository "github.com/wndhydrnt/saturn-bot-example".
saturn-bot try \
  --repository github.com/wndhydrnt/saturn-bot-example \
  --task-name example \
  task.yaml

# Set inputs "version" and "date".
# The task in file "task.yaml" defines the expected inputs.
saturn-bot try \
  --repository github.com/wndhydrnt/saturn-bot-example \
	--input version=1.2.3 \
	--input date=2024-11-10 \
  task.yaml
`
)

func createTryCommand() *cobra.Command {
	var dataDir string
	var inputs map[string]string
	var repository string
	var taskName string

	cmd := &cobra.Command{
		Use:   "try FILE",
		Short: "Try out a task locally",
		Long:  tryCommandHelp,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Read(cfgFile)
			handleError(err, cmd.ErrOrStderr())
			opts, err := options.ToOptions(cfg)
			handleError(err, cmd.ErrOrStderr())
			runner, err := command.NewTryRunner(opts, dataDir, repository, args[0], taskName, inputs)
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
	cmd.Flags().StringVar(&taskName, "task-name", "", `If set, try only the task that matches the name.
Useful if a task file contains multiple tasks.`)
	cmd.Flags().StringToStringVar(&inputs, "input", map[string]string{}, `Key/value pairs to use as input parameters of the task.
Can be supplied multiple times.`)
	return cmd
}
