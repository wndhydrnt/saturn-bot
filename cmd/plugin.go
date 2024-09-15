package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
)

var (
	execPluginOpts = command.ExecPluginOptions{}
)

func createPluginCommand() *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Commands to debug and test plugins",
		Long:  "Commands to debug and test plugins.",
	}
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.Addr, "address", "", `Address of the plugin to connect to.
Useful to debug a running plugin process.
Mutually exclusive with --path.`)
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.LogFormat, "log-format", "auto", "Log format of saturn-bot (auto,console,json).")
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.LogLevel, "log-level", "error", "Log level of saturn-bot.")
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.Path, "path", "", `Path to the plugin file.
Starts the plugin before executing the requested function.
Mutually exclusive with --address.`)
	pluginCmd.PersistentFlags().StringToStringVar(&execPluginOpts.Config, "config", map[string]string{}, "Key/value pairs to pass as configuration to the plugin.")
	pluginCmd.PersistentFlags().StringVar(&execPluginOpts.WorkDir, "workdir", "", `Path to the directory that contains files the apply function can modify.
Uses a temporary directory if not set.`)
	pluginCmd.Flags().StringVar(&execPluginOpts.Context, "context", command.DefaultContext, "Context data to send to the plugin.")

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
			execPluginOpts.Out = cmd.OutOrStdout()
			err := command.ExecPlugin(name, execPluginOpts)
			handleError(err, cmd.ErrOrStderr())
		},
	}
}
