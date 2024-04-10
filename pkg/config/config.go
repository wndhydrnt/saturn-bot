package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/viper"
)

var (
	//go:embed config.schema.json
	schemaRaw string
)

type Config struct {
	Custom              map[string]string `json:"custom,omitempty" yaml:",omitempty"`
	customMarshaled     []byte
	DataDir             string `json:"dataDir,omitempty" yaml:",omitempty"`
	DryRun              bool   `json:"dryRun,omitempty" yaml:",omitempty"`
	LogFormat           string `json:"logFormat,omitempty" yaml:",omitempty"`
	LogLevel            string `json:"logLevel,omitempty" yaml:",omitempty"`
	GitAuthor           string `json:"gitAuthor,omitempty" yaml:",omitempty"`
	gitUserEmail        string
	gitUserName         string
	GitCloneOptions     []string `json:"gitCloneOptions,omitempty" yaml:",omitempty"`
	GitCommitMessage    string   `json:"gitCommitMessage,omitempty" yaml:",omitempty"`
	GitLogLevel         string   `json:"gitLogLevel,omitempty" yaml:",omitempty"`
	GitPath             string   `json:"gitPath,omitempty" yaml:",omitempty"`
	GitHubAddress       string   `json:"githubAddress,omitempty" yaml:",omitempty"`
	GitHubCacheDisabled bool     `json:"githubCacheDisabled,omitempty" yaml:",omitempty"`
	GitHubToken         string   `json:"githubToken,omitempty" yaml:",omitempty"`
	GitLabAddress       string   `json:"gitlabAddress,omitempty" yaml:",omitempty"`
	GitLabToken         string   `json:"gitlabToken,omitempty" yaml:",omitempty"`
}

func (c Config) CustomMarshaled() []byte {
	if c.customMarshaled == nil {
		return []byte("{}")
	}

	return c.customMarshaled
}

func (c Config) GitUserEmail() string {
	return c.gitUserEmail
}

func (c Config) GitUserName() string {
	return c.gitUserName
}

func (c Config) toValidation() interface{} {
	var data interface{}
	b, _ := json.Marshal(&c)
	_ = json.Unmarshal(b, &data)
	return data
}

func Read(cfgFile string) (Config, error) {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	err := compiler.AddResource("config.schema.json", strings.NewReader(schemaRaw))
	if err != nil {
		return Config{}, fmt.Errorf("add resource to JSON schema compiler: %w", err)
	}

	schema, err := compiler.Compile("config.schema.json")
	if err != nil {
		return Config{}, fmt.Errorf("compile json schema: %w", err)
	}

	setDefaultsFromJSONSchema(schema)
	viper.SetEnvPrefix("SATURN_")
	viper.AutomaticEnv()
	if cfgFile != "" {
		slog.Debug("adding config file", "file", cfgFile)
		viper.SetConfigFile(cfgFile)
		err := viper.ReadInConfig()
		if err != nil {
			slog.Error("read config file", "err", err, "file", cfgFile)
		}
	}
	cfg := Config{}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("unmarshal configuration: %w", err)
	}

	if cfg.Custom == nil {
		cfg.Custom = map[string]string{}
	}

	err = schema.Validate(cfg.toValidation())
	if err != nil {
		return cfg, fmt.Errorf("schema validation failed: %w", err)
	}

	cfg.customMarshaled, err = json.Marshal(cfg.Custom)
	if err != nil {
		return cfg, fmt.Errorf("marshal custom configuration into JSON: %w", err)
	}

	mailAddress, err := mail.ParseAddress(cfg.GitAuthor)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse field `gitAuthor`: %w", err)
	}

	cfg.gitUserEmail = mailAddress.Address
	cfg.gitUserName = mailAddress.Name

	return cfg, nil
}

func setDefaultsFromJSONSchema(schema *jsonschema.Schema) {
	for name, item := range schema.Properties {
		if item.Default != nil {
			viper.SetDefault(name, item.Default)
		}
	}
}
