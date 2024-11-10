# try

```{.text mdox-exec="./saturn-bot try --help" title="try"}
Try out a task locally.

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

Usage:
  saturn-bot try FILE [flags]

Flags:
      --config string          Path to config file.
      --data-dir string        Path to directory to clone the repository.
  -h, --help                   help for try
      --input stringToString   Key/value pairs to use as input parameters of the task.
                               Can be supplied multiple times. (default [])
      --repository string      Name of the repository to test against.
      --task-name string       If set, try only the task that matches the name.
                               Useful if a task file contains multiple tasks.
```
