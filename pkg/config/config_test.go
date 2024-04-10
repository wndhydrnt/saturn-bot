package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestReadConfig(t *testing.T) {
	testCases := []struct {
		name string
		in   Config
		out  Config
	}{
		{
			name: "default and required values",
			in:   Config{},
			out: Config{
				Custom:           map[string]string{},
				customMarshaled:  []byte("{}"),
				GitAuthor:        "saturn-sync <bot@saturn-sync.localhost>",
				gitUserEmail:     "bot@saturn-sync.localhost",
				gitUserName:      "saturn-sync",
				GitCloneOptions:  []string{"--filter", "blob:none"},
				GitCommitMessage: "changes by saturn-sync",
				GitLogLevel:      "warn",
				GitPath:          "git",
				LogFormat:        "auto",
				LogLevel:         "info",
			},
		},
		{
			name: "keys in custom configuration get parsed to lower-case",
			in: Config{
				Custom: map[string]string{"customKey": "customValue"},
			},
			out: Config{
				Custom:           map[string]string{"customkey": "customValue"},
				customMarshaled:  []byte(`{"customkey":"customValue"}`),
				GitAuthor:        "saturn-sync <bot@saturn-sync.localhost>",
				gitUserEmail:     "bot@saturn-sync.localhost",
				gitUserName:      "saturn-sync",
				GitCloneOptions:  []string{"--filter", "blob:none"},
				GitCommitMessage: "changes by saturn-sync",
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
