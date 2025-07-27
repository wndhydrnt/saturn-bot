package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"gopkg.in/yaml.v3"
)

// Holds all default values set after a configuration file has been parsed.
var defaultConfiguration = Configuration{
	DataDir:                  nil,
	GitCloneOptions:          []string{"--filter", "blob:none"},
	GitCommitMessage:         "changes by saturn-bot",
	GitlabAddress:            "https://gitlab.com",
	GitLogLevel:              "warn",
	GitPath:                  "git",
	GitUrl:                   "https",
	LogFormat:                "auto",
	LogLevel:                 "info",
	JavaPath:                 "java",
	Labels:                   []string{},
	PluginLogLevel:           "debug",
	PythonPath:               "python",
	RepositoryCacheTtl:       "6h",
	ServerAddr:               ":3035",
	ServerBaseUrl:            "http://localhost:3035",
	ServerCompress:           true,
	ServerServeUi:            true,
	ServerShutdownTimeout:    "5m",
	WorkerLoopInterval:       "10s",
	WorkerParallelExecutions: 1,
	WorkerServerAPIBaseURL:   "http://localhost:3035",
}

func TestReadConfig(t *testing.T) {
	testCases := []struct {
		name    string
		environ map[string]string
		in      Configuration
		out     Configuration
		readErr string
	}{
		{
			name: "default and required values",
			in: Configuration{
				GitlabToken: ptr.To("abc"),
			},
			out: func(cfg Configuration) Configuration {
				cfg.GitlabToken = ptr.To("abc")
				return cfg
			}(defaultConfiguration),
		},
		{
			name: "github token or gitlab token required",
			in: Configuration{
				GithubToken: nil,
				GitlabToken: nil,
			},
			readErr: "no githubToken or gitlabToken configured - https://saturn-bot.readthedocs.io/en/latest/configuration/",
		},
		{
			name: "env vars take precedence",
			environ: map[string]string{
				"SATURN_BOT_GITLABADDRESS": "http://gitlab.local",
				"SATURN_BOT_GITLABTOKEN":   "def",
			},
			in: Configuration{
				GitlabAddress: "https://gitlab.com",
				GitlabToken:   ptr.To("abc"),
			},
			out: func(cfg Configuration) Configuration {
				cfg.GitlabAddress = "http://gitlab.local"
				cfg.GitlabToken = ptr.To("def")
				return cfg
			}(defaultConfiguration),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.environ != nil {
				for k, v := range tc.environ {
					err := os.Setenv(k, v)
					require.NoErrorf(t, err, "set environment variable %s", k)
				}

				defer func() {
					for k := range tc.environ {
						err := os.Unsetenv(k)
						require.NoErrorf(t, err, "unset environment variable %s", k)
					}
				}()
			}

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
