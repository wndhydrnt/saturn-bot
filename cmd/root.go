package cmd

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "saturn-bot",
	Short: "Synchronize code across many repositories",
	Long:  `Synchronize code across many repositories.`,
}

func Execute() {
	rootCmd.AddCommand(createRunCommand())
	rootCmd.AddCommand(createServerCommand())
	rootCmd.AddCommand(createTryCommand())
	rootCmd.AddCommand(createVersionCommand())
	rootCmd.AddCommand(createWorkerCommand())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func handleError(err error, out io.Writer) {
	if err == nil {
		return
	}

	var validationError *jsonschema.ValidationError
	if errors.As(err, &validationError) {
		slog.Error("Configuration file contains errors:")
		fmt.Fprintf(out, "%s\n", validationError.Error())
	} else {
		slog.Error("command failed", "err", err)
	}

	os.Exit(1)
}
