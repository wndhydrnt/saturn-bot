# ci

```{.text mdox-exec="./saturn-bot ci --help" title="ci"}
Loads and validates task files.

"ci" can be executed during a continuous integration process
to receive early feedback on syntax errors or invalid values
in task files.

It also starts any plugins defined in a task file
and calls their initialize function.
Pass --start-plugins=false to prevent this.

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

## Prevent in plugins

`ci` initializes every plugin defined in a task file.
It might not be desirable that a plugin executes code that,
for example calls external APIs.
Plugin code can check if it runs as part of a CI run via the configuration:

```go
func (p Plugin) Init(config map[string]string) error {
	if config["saturn-bot.ci"] == "true" {
		return nil
	}

	// init code
}
```
