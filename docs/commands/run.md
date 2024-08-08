# run

```{.text mdox-exec="./saturn-bot run --help" title="run"}
Execute tasks against repositories.

"run" executes all tasks in the given task file(s).
It first lists all repositories from the source, for example GitHub or GitLab.
Then it modifies a repository by executing the actions of a task if the filters
of that task match.
If files have been modified, it creates a pull request for the repository.

Examples:

Execute task in file "task.yaml" against all repositories:

saturn-bot run \
  --config config.yaml \
  task.yaml

Execute tasks in files "task1.yaml" and "task2.yaml"
against all repositories:

saturn-bot run \
  --config config.yaml \
  task1.yaml \
  task2.yaml

Globbing support:

saturn-bot run \
  --config config.yaml \
  *.yaml

Execute task in file "task.yaml" against
repository "github.com/wndhydrnt/saturn-bot-example":

saturn-bot run \
  --config config.yaml \
  --repository github.com/wndhydrnt/saturn-bot-example \
  task.yaml

Usage:
  saturn-bot run FILE [FILE...] [flags]

Flags:
      --config string            Path to config file
  -h, --help                     help for run
      --repository stringArray   Name of a repository to apply the tasks to.
                                 Filters of a task aren't executed if this flag
                                 is set.
                                 Can be supplied multiple times.
```
