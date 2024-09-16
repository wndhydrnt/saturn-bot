package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	protoV1 "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
)

var (
	execPluginOpts    = command.ExecPluginOptions{}
	pluginContextJSON string
)

var pluginCommandHelp = `Commands to debug and test plugins.

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
`

func createPluginCommand() *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Commands to debug and test plugins",
		Long:  pluginCommandHelp,
	}
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.Addr, "address", "", `Address of the plugin to connect to.
Useful to debug a running plugin process.
Mutually exclusive with --path.`)
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.LogFormat, "log-format", "auto", "Log format of saturn-bot (auto,console,json).")
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.LogLevel, "log-level", "error", "Log level of saturn-bot.")
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.Path, "path", "", `Path to the plugin file.
Starts the plugin before executing the requested function.
Mutually exclusive with --address.`)
	pluginCmd.PersistentFlags().StringToStringVar(&execPluginOpts.Config, "config", map[string]string{}, `Key/value pairs to pass as configuration to the plugin.
Supply multiple times to add additional key/value pairs.`)
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.WorkDir, "workdir", "", `Path to the directory that contains files the apply function can modify.
Uses a temporary directory if not set.`)
	pluginCmd.Flags().StringVar(&pluginContextJSON, "context", "", "Context data to send to the plugin.")

	for funcName := range command.PluginFuncs {
		funcCmd := createPluginFuncCommand(funcName)
		pluginCmd.AddCommand(funcCmd)
	}

	return pluginCmd
}

func createPluginFuncCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Test the " + name + " function of a plugin",
		Long:  "Test the " + name + " function of a plugin.",
		Run: func(cmd *cobra.Command, args []string) {
			if pluginContextJSON != "" {
				execPluginOpts.Context = &protoV1.Context{}
				dec := json.NewDecoder(strings.NewReader(pluginContextJSON))
				err := dec.Decode(&execPluginOpts.Context)
				handleError(fmt.Errorf("decode plugin context from JSON: %w", err), cmd.ErrOrStderr())
			}

			execPluginOpts.Out = cmd.OutOrStdout()
			err := command.ExecPlugin(name, execPluginOpts)
			handleError(err, cmd.ErrOrStderr())
		},
	}
}
