# ci

```{.text mdox-exec="./saturn-bot ci --help" title="ci"}
Loads and validates task files.

"ci" can be executed during a continuous integration process
to receive early feedback on syntax errors or invalid values
in task files.

It also starts any plugins defined in a task file
and calls their initialize function.
Pass --start-plugins=false to prevent this.

The command exits with exit code "1" if validation fails.

Examples:

# Validate one task file
saturn-bot ci ./task.yaml

# Validate multiple task files
saturn-bot ci ./*.yaml

Usage:
  saturn-bot ci FILE [FILE...] [flags]

Flags:
      --config string   Path to config file
  -h, --help            help for ci
      --skip-plugins    Skip starting plugins as part of the CI run.
```

`ci` starts and initializes plugins as part of its execution.
If this behavior isn't desired, the flag `--skip-plugins=true` can be passed to the command.
It is also possible to make each plugin
[skip initialization during CI runs](../task/plugins/index.md#skip-initialization-during-ci-runs).
