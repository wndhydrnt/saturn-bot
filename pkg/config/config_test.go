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
		in   Configuration
		out  Configuration
	}{
		{
			name: "default and required values",
			in:   Configuration{},
			out: Configuration{
				DataDir:                  nil,
				GitCloneOptions:          []string{"--filter", "blob:none"},
				GitCommitMessage:         "changes by saturn-bot",
				GitlabAddress:            "https://gitlab.com",
				GitLogLevel:              "warn",
				GitPath:                  "git",
				LogFormat:                "auto",
				LogLevel:                 "info",
				JavaPath:                 "java",
				PythonPath:               "python",
				ServerAddr:               ":3035",
				ServerBaseUrl:            "http://localhost:3035",
				ServerCompress:           true,
				WorkerLoopInterval:       "10s",
				WorkerParallelExecutions: 4,
				WorkerServerAPIBaseURL:   "http://localhost:3035",
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
