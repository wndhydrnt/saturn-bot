package config

import (
	_ "embed"
	"encoding/json"
	"errors"
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
	envPrefix = "SATURN_BOT_"
)

var (
	//go:embed config.schema.json
	schemaRaw string
	// ErrNoToken informs the user that at least one token is required.
	ErrNoToken = errors.New("no githubToken or gitlabToken configured - https://saturn-bot.readthedocs.io/en/latest/configuration/")
)

func (c Configuration) GitUserEmail() string {
	mailAddress, err := mail.ParseAddress(c.GitAuthor)
	if err != nil {
		return ""
	}

	return mailAddress.Address
}

func (c Configuration) GitUserName() string {
	mailAddress, err := mail.ParseAddress(c.GitAuthor)
	if err != nil {
		return ""
	}

	return mailAddress.Name
}

func toValidation(c Configuration) interface{} {
	var data interface{}
	b, _ := json.Marshal(c)
	_ = json.Unmarshal(b, &data)
	return data
}

func Read(cfgFile string) (cfg Configuration, err error) {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	err = compiler.AddResource("config.schema.json", strings.NewReader(schemaRaw))
	if err != nil {
		return cfg, fmt.Errorf("add resource to JSON schema compiler: %w", err)
	}

	schema, err := compiler.Compile("config.schema.json")
	if err != nil {
		return cfg, fmt.Errorf("compile json schema: %w", err)
	}

	koanfDefaults := createKoanfDefaults(schema)
	var k = koanf.New("")
	// Define defaults first
	if err := k.Load(confmap.Provider(koanfDefaults, ""), nil); err != nil {
		return cfg, fmt.Errorf("load configuration defaults: %w", err)
	}

	if cfgFile != "" {
		// If file is defined, load configuration from file
		if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
			return cfg, fmt.Errorf("load configuration file: %w", err)
		}
	}

	// Load configuration from env vars last. Env vars overwrite previously set configuration.
	if err := k.Load(env.Provider(envPrefix, ".", createEnvMapper(schema)), nil); err != nil {
		return cfg, fmt.Errorf("load configuration from environment: %w", err)
	}

	err = k.Unmarshal("", &cfg)
	if err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}

	err = schema.Validate(toValidation(cfg))
	if err != nil {
		return cfg, fmt.Errorf("schema validation failed: %w", err)
	}

	if cfg.GitAuthor != "" {
		_, err = mail.ParseAddress(cfg.GitAuthor)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse field `gitAuthor`: %w", err)
		}
	}

	if cfg.GithubToken == nil && cfg.GitlabToken == nil {
		return cfg, ErrNoToken
	}

	return cfg, nil
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
