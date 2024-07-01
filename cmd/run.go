package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
)

func createRunCommand() *cobra.Command {
	var repositories []string
	var taskFiles []string

	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Run locally",
		Long:  "Run locally",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Read(cfgFile)
			handleError(err, cmd.ErrOrStderr())
			opts, err := options.ToOptions(cfg)
			handleError(err, cmd.ErrOrStderr())
			err = pkg.ExecuteRun(opts, repositories, taskFiles)
			handleError(err, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	cmd.Flags().StringArrayVar(&repositories, "repository", []string{}, "Name of a repository to apply the tasks to. Can be supplied multiple times.")
	cmd.Flags().StringArrayVar(&taskFiles, "task", []string{}, "Path to a file to read Tasks from. Can be supplied multiple times.")
	return cmd
}
