package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
)

type CiRunner struct {
	Opts options.Opts
}

func NewCiRunner(opts options.Opts) (*CiRunner, error) {
	err := options.Initialize(&opts)
	if err != nil {
		return nil, fmt.Errorf("initialize options: %w", err)
	}

	return &CiRunner{Opts: opts}, nil
}

// NewCiRunnerFromConfig creates a new CiRunner by reading from configFile.
// It ignores some errors that are not relevant in a CI run.
func NewCiRunnerFromConfig(configFile string, skipPlugins bool) (*CiRunner, error) {
	cfg, err := config.Read(configFile)
	// Ignore config.ErrNoToken because no token is needed during CI.
	if err != nil && !errors.Is(err, config.ErrNoToken) {
		return nil, err
	}

	opts, err := options.ToOptions(cfg)
	if err != nil && !errors.Is(err, options.ErrNoHosts) {
		return nil, err
	}

	opts.SkipPlugins = skipPlugins
	return NewCiRunner(opts)
}

// Run reads and validates taskFiles and writes a message to out for each file.
func (ci *CiRunner) Run(out io.Writer, taskFiles ...string) error {
	for _, taskFile := range taskFiles {
		// Create a new registry for each file.
		// Makes it easier to connect an error to a task file.
		reg := task.NewRegistry(ci.Opts)
		err := reg.ReadTasks(taskFile)
		if err != nil {
			fmt.Fprintf(out, "❌ Validation failed: %s\n", err)
			continue
		}

		for _, task := range reg.GetTasks() {
			fmt.Fprintf(out, "✅ Valid task %s found\n", task.Name)
			task.Stop()
		}
	}

	return nil
}
