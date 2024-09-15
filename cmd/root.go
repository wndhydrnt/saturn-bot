package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/cobra"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "saturn-bot",
	Short: "Synchronize code across many repositories",
	Long:  `Synchronize code across many repositories.`,
}

func Execute() int {
	rootCmd.AddCommand(createExperimentalCommand())
	rootCmd.AddCommand(createPluginCommand())
	rootCmd.AddCommand(createRunCommand())
	rootCmd.AddCommand(createTryCommand())
	rootCmd.AddCommand(createVersionCommand())
	if err := rootCmd.Execute(); err != nil {
		return 1
	}

	return 0
}

func handleError(err error, out io.Writer) {
	if err == nil {
		return
	}

	var validationError *jsonschema.ValidationError
	if errors.As(err, &validationError) {
		log.Log().Error("Configuration file contains errors:")
		fmt.Fprintf(out, "%s\n", validationError.Error())
	} else {
		log.Log().Errorf("command failed: %v", err)
	}

	os.Exit(1)
}
