package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func defaultDataDir() *string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("defaultDataDir: %w", err))
	}

	dir := filepath.Join(homeDir, ".saturn-bot", "data")
	return &dir
}

func TestReadConfig(t *testing.T) {
	testCases := []struct {
		name string
		in   Configuration
		out  Configuration
	}{
		{
			name: "default and required values",
			in:   Configuration{},
			out: Configuration{
				DataDir:          defaultDataDir(),
				GitCloneOptions:  []string{"--filter", "blob:none"},
				GitCommitMessage: "changes by saturn-bot",
				GitlabAddress:    "https://gitlab.com",
				GitLogLevel:      "warn",
				GitPath:          "git",
				LogFormat:        "auto",
				LogLevel:         "info",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "*config.yaml")
			require.NoError(t, err)
			defer f.Close()
			enc := yaml.NewEncoder(f)
			err = enc.Encode(tc.in)
			enc.Close()
			require.NoError(t, err)

			readCfg, err := Read(f.Name())
			require.NoError(t, err)

			require.EqualValues(t, tc.out, readCfg)
		})
	}
}
