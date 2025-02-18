# plugin

```{.text mdox-exec="./saturn-bot plugin --help" title="plugin"}
Commands to debug and test plugins.

Each sub-command executes one function of a plugin.

The flags --address and --path define how saturn-bot calls the plugin.
If --path is set, saturn-bot attempts to start the plugin file.
If --address is set, saturn-bot connects to the given address
without starting the plugin.

# Start the Python plugin and call its apply function.
saturn-bot plugin apply --path ./plugin.py

# Connect to the plugin using the connection string.
# Plugin has been started in another terminal.
saturn-bot plugin apply --address '1|1|tcp|127.0.0.1:11049|grpc'

# Start the Python plugin, pass configuration to it and call its filter function.
saturn-bot plugin apply --path ./plugin.py --config 'message=example'

Usage:
  saturn-bot plugin [command]

Available Commands:
  apply       Test the apply function of a plugin
  filter      Test the filter function of a plugin
  onPrClosed  Test the onPrClosed function of a plugin
  onPrCreated Test the onPrCreated function of a plugin
  onPrMerged  Test the onPrMerged function of a plugin

Flags:
      --address string          Address of the plugin to connect to.
                                Useful to debug a running plugin process.
                                Mutually exclusive with --path.
      --config stringToString   Key/value pairs to pass as configuration to the plugin.
                                Supply multiple times to add additional key/value pairs. (default [])
      --context string          Context data to send to the plugin.
  -h, --help                    help for plugin
      --log-format string       Log format of saturn-bot (auto,console,json). (default "auto")
      --log-level string        Log level of saturn-bot. (default "error")
      --path string             Path to the plugin file.
                                Starts the plugin before executing the requested function.
                                Mutually exclusive with --address.
      --workdir string          Path to the directory that contains files the apply function can modify.
                                Uses a temporary directory if not set.

Use "saturn-bot plugin [command] --help" for more information about a command.
```
