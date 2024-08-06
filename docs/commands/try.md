# try

```{.text mdox-exec="./saturn-bot try --help" title="try"}
Try out a task locally.

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
  --task-name example

Usage:
  saturn-bot try [flags]

Flags:
      --config string           Path to config file.
      --data-dir string         Path to directory to clone the repository.
  -h, --help                    help for try
      --inputs stringToString    (default [])
      --repository string       Name of the repository to test against.
      --task-file string        Path to the task file to try out.
      --task-name string        If set, try only the task that matches the name.
                                Useful if a task file contains multiple tasks.
```
