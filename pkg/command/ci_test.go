package command_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
)

const (
	taskValid = `
name: Valid Task One
---
name: Valid Task Two
`
	taskInvalid = `
name: Invalid Task
active: abc
`
)

func TestCiRunner_Run_Valid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	validFile, err := os.Create(filepath.Join(tmpDir, "valid.yaml"))
	require.NoError(t, err)
	_, err = validFile.WriteString(taskValid)
	require.NoError(t, err)
	validFile.Close()

	runner, err := command.NewCiRunner(options.Opts{
		Config: config.Configuration{WorkerLoopInterval: "1m"},
	})
	require.NoError(t, err)

	out := &bytes.Buffer{}
	err = runner.Run(out, filepath.Join(tmpDir, "*.yaml"))

	require.NoError(t, err)
	msg := `✅ Valid task Valid Task One found
✅ Valid task Valid Task Two found
`
	require.Equal(t, msg, out.String())
}

func TestCiRunner_Run_Invalid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	invalidFile, err := os.Create(filepath.Join(tmpDir, "invalid.yaml"))
	require.NoError(t, err)
	_, err = invalidFile.WriteString(taskInvalid)
	require.NoError(t, err)
	invalidFile.Close()

	runner, err := command.NewCiRunner(options.Opts{
		Config: config.Configuration{WorkerLoopInterval: "1m"},
	})
	require.NoError(t, err)

	out := &bytes.Buffer{}
	err = runner.Run(out, filepath.Join(tmpDir, "*.yaml"))

	require.Error(t, err, "validation failed")
	require.Contains(t, out.String(), "Validation failed: decode task from YAML file")
}

func TestNewCiRunnerFromConfig(t *testing.T) {
	configFile, err := os.CreateTemp("", "*.yaml")
	require.NoError(t, err)
	configFile.Close() // Empty file

	runner, err := command.NewCiRunnerFromConfig(configFile.Name(), false)

	require.NoError(t, err, "Should not error when config file is empty")
	require.IsType(t, &command.CiRunner{}, runner)
}
