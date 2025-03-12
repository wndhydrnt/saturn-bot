package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

var (
	scheduleCommandHelp = `Schedule a run via the server API.

"schedule" schedules a new run of TASK_NAME at the server
provided by --server-url.
It blocks until the run has finished and reports its result.

If blocking isn't desired, pass --wait 0.

Examples:

# Schedule a run of task with the name "hello-world"
# using default values.
saturn-bot schedule hello-world

# Schedule a run of task with the name "hello-world"
# and do not wait for it to finish.
saturn-bot schedule \
  --wait 0 \
  hello-world

# Schedule a run of task with the name "hello-world"
# and inputs.
saturn-bot schedule \
  --input greeting=Hello \
  --input to=World \
  hello-world
`
)

func createScheduleCommand() *cobra.Command {
	var inputs map[string]string
	var serverApiKey string
	var serverUrl string
	var waitFor time.Duration
	var waitCheckInterval time.Duration

	cmd := &cobra.Command{
		Use:   "schedule TASK_NAME",
		Short: "Schedule a run via the server API",
		Long:  scheduleCommandHelp,
		Run: func(cmd *cobra.Command, args []string) {
			runner, err := command.NewScheduleRunner(command.ScheduleOptions{
				WaitFor:      waitFor,
				WaitInterval: waitCheckInterval,
				ServerApiKey: serverApiKey,
				ServerUrl:    serverUrl,
			})
			handleError(err, cmd.ErrOrStderr())
			err = runner.Run(cmd.OutOrStdout(), client.ScheduleRunV1Request{
				RunData:  ptr.To(inputs),
				TaskName: args[0],
			})
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringToStringVar(&inputs, "input", map[string]string{}, `Key/value pair in the format <key>=<value>
to use as an input parameter of a task.
Can be supplied multiple times to set multiple inputs.`)
	cmd.Flags().StringVar(&serverApiKey, "server-api-key", "", "Key to authenticate at the server API.")
	cmd.Flags().StringVar(&serverUrl, "server-url", "http://localhost:3035", "Base URL of the server API.")
	cmd.Flags().DurationVar(&waitFor, "wait", 5*time.Minute, `Wait for the run to finish.
The command blocks until the duration is over.
Useful to provide users with feedback on the result of the scheduled run.`)
	cmd.Flags().DurationVar(&waitCheckInterval, "wait-check-interval", 10*time.Second, "Time to wait between checks. Only relevant if --wait is set.")

	return cmd
}
