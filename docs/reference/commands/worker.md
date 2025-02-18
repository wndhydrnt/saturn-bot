# worker

```{.text mdox-exec="./saturn-bot worker --help" title="worker"}
Starts the worker component.

"worker" queries the server component for tasks to execute,
executes them and reports the results back to the server.

Examples:

# Start the worker
saturn-bot worker --config config.yaml ./tasks/**/*.yaml

Usage:
  saturn-bot worker FILE [FILE...] [flags]

Flags:
      --config string   Path to config file
  -h, --help            help for worker
```
