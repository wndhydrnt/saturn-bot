package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
)

var (
	runCommandHelp = `Execute tasks against repositories.

"run" executes all tasks in the given task file(s).
It first lists all repositories from the source, for example GitHub or GitLab.
Then it modifies a repository by executing the actions of a task if the filters
of that task match.
If files have been modified, it creates a pull request for the repository.

Examples:

# Execute task in file "task.yaml" against all repositories.
saturn-bot run task.yaml

# Execute tasks in files "task1.yaml" and "task2.yaml"
# against all repositories.
saturn-bot run \
  task1.yaml \
  task2.yaml

# Globbing support.
saturn-bot run *.yaml

# Execute task in file "task.yaml" against
# repository "github.com/wndhydrnt/saturn-bot-example".
saturn-bot run \
  --repository github.com/wndhydrnt/saturn-bot-example \
  task.yaml

# Set inputs "version" and "date".
# The task in file "task.yaml" defines the expected inputs.
saturn-bot run \
  --input version=1.2.3 \
  --input date=2024-11-10 \
  task.yaml
`
)

func createRunCommand() *cobra.Command {
	var inputs map[string]string
	var repositories []string

	var cmd = &cobra.Command{
		Use:   "run FILE [FILE...]",
		Short: "Execute tasks against repositories",
		Long:  runCommandHelp,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Read(cfgFile)
			handleError(err, cmd.ErrOrStderr())
			opts, err := options.ToOptions(cfg)
			handleError(err, cmd.ErrOrStderr())
			_, err = command.ExecuteRun(opts, repositories, args, inputs)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	cmd.Flags().StringToStringVar(&inputs, "input", map[string]string{}, `Key/value pairs to use as input parameters of tasks.
Can be supplied multiple times.`)
	cmd.Flags().StringArrayVar(&repositories, "repository", []string{}, `Name of a repository to apply the tasks to.
Filters of a task aren't executed if this flag
is set.
Can be supplied multiple times.`)
	return cmd
}
