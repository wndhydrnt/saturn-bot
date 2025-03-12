# schedule

```{.text mdox-exec="./saturn-bot schedule --help" title="run"}
Schedule a run via the server API.

"schedule" schedules a new run of TASK_NAME at the server
provided by --server-url.
It blocks until the run has finished and reports its result.

If blocking isn't desired, pass --wait=0.

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

Usage:
  saturn-bot schedule TASK_NAME [flags]

Flags:
  -h, --help                           help for schedule
      --input stringToString           Key/value pair in the format <key>=<value>
                                       to use as an input parameter of a task.
                                       Can be supplied multiple times to set multiple inputs. (default [])
      --server-api-key string          Key to authenticate at the server API.
      --server-url string              Base URL of the server API. (default "http://localhost:3035")
      --wait duration                  Wait for the run to finish.
                                       The command blocks until the duration is over.
                                       Useful to provide users with feedback on the result of the scheduled run. (default 15m0s)
      --wait-check-interval duration   Time to wait between checks. Only relevant if --wait is set. (default 10s)
```
