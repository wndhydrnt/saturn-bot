package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	envPrefix = "SATURN_SYNC_"
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

	koanfDefaults := createKoanfDefaults(schema)
	var k = koanf.New("")
	// Define defaults first
	if err := k.Load(confmap.Provider(koanfDefaults, ""), nil); err != nil {
		return Config{}, fmt.Errorf("load configuration defaults: %w", err)
	}

	if cfgFile != "" {
		// If file is defined, load configuration from file
		if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
			return Config{}, fmt.Errorf("load configuration file: %w", err)
		}
	}

	// Load configuration from env vars last. Env vars overwrite previously set configuration.
	if err := k.Load(env.Provider(envPrefix, ".", createEnvMapper(schema)), nil); err != nil {
		return Config{}, fmt.Errorf("load configuration from environment: %w", err)
	}

	var c Config
	err = k.Unmarshal("", &c)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if c.Custom == nil {
		c.Custom = map[string]string{}
	}

	err = schema.Validate(c.toValidation())
	if err != nil {
		return c, fmt.Errorf("schema validation failed: %w", err)
	}

	c.customMarshaled, err = json.Marshal(c.Custom)
	if err != nil {
		return c, fmt.Errorf("marshal custom configuration into JSON: %w", err)
	}

	mailAddress, err := mail.ParseAddress(c.GitAuthor)
	if err != nil {
		return c, fmt.Errorf("failed to parse field `gitAuthor`: %w", err)
	}

	c.gitUserEmail = mailAddress.Address
	c.gitUserName = mailAddress.Name

	return c, nil
}

func createKoanfDefaults(schema *jsonschema.Schema) map[string]interface{} {
	defaults := map[string]interface{}{}
	for name, item := range schema.Properties {
		if item.Default != nil {
			defaults[name] = item.Default
		}
	}

	return defaults
}

func createEnvMapper(schema *jsonschema.Schema) func(s string) string {
	mapping := map[string]string{}
	for name := range schema.Properties {
		envKey := strings.ToUpper(fmt.Sprintf("%s%s", envPrefix, name))
		mapping[envKey] = name
	}

	return func(s string) string {
		realKey, ok := mapping[s]
		if ok {
			return realKey
		}

		return s
	}
}
