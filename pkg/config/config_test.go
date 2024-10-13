package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"gopkg.in/yaml.v3"
)

func TestReadConfig(t *testing.T) {
	testCases := []struct {
		name    string
		in      Configuration
		out     Configuration
		readErr string
	}{
		{
			name: "default and required values",
			in: Configuration{
				GitlabToken: ptr.To("abc"),
			},
			out: Configuration{
				DataDir:                  nil,
				GitCloneOptions:          []string{"--filter", "blob:none"},
				GitCommitMessage:         "changes by saturn-bot",
				GitlabAddress:            "https://gitlab.com",
				GitlabToken:              ptr.To("abc"),
				GitLogLevel:              "warn",
				GitPath:                  "git",
				GitUrl:                   "https",
				LogFormat:                "auto",
				LogLevel:                 "info",
				JavaPath:                 "java",
				PluginLogLevel:           "debug",
				PythonPath:               "python",
				ServerAddr:               ":3035",
				ServerBaseUrl:            "http://localhost:3035",
				ServerCompress:           true,
				WorkerLoopInterval:       "10s",
				WorkerParallelExecutions: 4,
				WorkerServerAPIBaseURL:   "http://localhost:3035",
			},
		},
		{
			name: "github token or gitlab token required",
			in: Configuration{
				GithubToken: nil,
				GitlabToken: nil,
			},
			readErr: "no GitHub token or GitLab token configured",
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
			if tc.readErr == "" {
				require.NoError(t, err)
				require.EqualValues(t, tc.out, readCfg)
			} else {
				require.EqualError(t, err, tc.readErr)
			}
		})
	}
}
