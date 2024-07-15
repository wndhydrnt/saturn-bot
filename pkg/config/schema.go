// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package config

import "encoding/json"
import "fmt"
import yaml "gopkg.in/yaml.v3"
import "reflect"

// Configuration settings of saturn-bot.
type Configuration struct {
	// Path to directory to store files and repository clones.
	DataDir *string `json:"dataDir,omitempty" yaml:"dataDir,omitempty" mapstructure:"dataDir,omitempty"`

	// Toggle dry-run mode. No pull requests will be created or merged when enabled.
	DryRun bool `json:"dryRun,omitempty" yaml:"dryRun,omitempty" mapstructure:"dryRun,omitempty"`

	// Author to use for git commits. Global git configuration applies if not set.
	// Must conform to RFC5322: `User Name <user@name.local>`.
	GitAuthor string `json:"gitAuthor,omitempty" yaml:"gitAuthor,omitempty" mapstructure:"gitAuthor,omitempty"`

	// Command-line options to pass to `git clone`.
	GitCloneOptions []string `json:"gitCloneOptions,omitempty" yaml:"gitCloneOptions,omitempty" mapstructure:"gitCloneOptions,omitempty"`

	// Default commit message to use if a task does not define a custom one.
	GitCommitMessage string `json:"gitCommitMessage,omitempty" yaml:"gitCommitMessage,omitempty" mapstructure:"gitCommitMessage,omitempty"`

	// Level for logs sent by the git sub-system. These logs can be very verbose and
	// can make it tricky to find logs of other sub-systems.
	GitLogLevel ConfigurationGitLogLevel `json:"gitLogLevel,omitempty" yaml:"gitLogLevel,omitempty" mapstructure:"gitLogLevel,omitempty"`

	// Path to `git` executable. PATH will be searched if not set.
	GitPath string `json:"gitPath,omitempty" yaml:"gitPath,omitempty" mapstructure:"gitPath,omitempty"`

	// Address of GitHub server to use.
	GithubAddress *string `json:"githubAddress,omitempty" yaml:"githubAddress,omitempty" mapstructure:"githubAddress,omitempty"`

	// If true, disables caching of HTTP responses received from the GitHub API.
	GithubCacheDisabled bool `json:"githubCacheDisabled,omitempty" yaml:"githubCacheDisabled,omitempty" mapstructure:"githubCacheDisabled,omitempty"`

	// Token to use for authentication at the GitHub API.
	GithubToken *string `json:"githubToken,omitempty" yaml:"githubToken,omitempty" mapstructure:"githubToken,omitempty"`

	// Address of GitLab server to use.
	GitlabAddress string `json:"gitlabAddress,omitempty" yaml:"gitlabAddress,omitempty" mapstructure:"gitlabAddress,omitempty"`

	// Token to use for authentication at the GitLab API.
	GitlabToken *string `json:"gitlabToken,omitempty" yaml:"gitlabToken,omitempty" mapstructure:"gitlabToken,omitempty"`

	// Format of log messages.
	LogFormat ConfigurationLogFormat `json:"logFormat,omitempty" yaml:"logFormat,omitempty" mapstructure:"logFormat,omitempty"`

	// Log level of the application.
	LogLevel ConfigurationLogLevel `json:"logLevel,omitempty" yaml:"logLevel,omitempty" mapstructure:"logLevel,omitempty"`
}

type ConfigurationGitLogLevel string

const ConfigurationGitLogLevelDebug ConfigurationGitLogLevel = "debug"
const ConfigurationGitLogLevelError ConfigurationGitLogLevel = "error"
const ConfigurationGitLogLevelInfo ConfigurationGitLogLevel = "info"
const ConfigurationGitLogLevelWarn ConfigurationGitLogLevel = "warn"

var enumValues_ConfigurationGitLogLevel = []interface{}{
	"debug",
	"error",
	"info",
	"warn",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConfigurationGitLogLevel) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationGitLogLevel {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationGitLogLevel, v)
	}
	*j = ConfigurationGitLogLevel(v)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *ConfigurationGitLogLevel) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationGitLogLevel {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationGitLogLevel, v)
	}
	*j = ConfigurationGitLogLevel(v)
	return nil
}

type ConfigurationLogFormat string

const ConfigurationLogFormatAuto ConfigurationLogFormat = "auto"
const ConfigurationLogFormatConsole ConfigurationLogFormat = "console"
const ConfigurationLogFormatJson ConfigurationLogFormat = "json"

var enumValues_ConfigurationLogFormat = []interface{}{
	"auto",
	"console",
	"json",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConfigurationLogFormat) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationLogFormat {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationLogFormat, v)
	}
	*j = ConfigurationLogFormat(v)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *ConfigurationLogFormat) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationLogFormat {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationLogFormat, v)
	}
	*j = ConfigurationLogFormat(v)
	return nil
}

type ConfigurationLogLevel string

const ConfigurationLogLevelDebug ConfigurationLogLevel = "debug"
const ConfigurationLogLevelError ConfigurationLogLevel = "error"
const ConfigurationLogLevelInfo ConfigurationLogLevel = "info"
const ConfigurationLogLevelWarn ConfigurationLogLevel = "warn"

var enumValues_ConfigurationLogLevel = []interface{}{
	"debug",
	"error",
	"info",
	"warn",
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *ConfigurationLogLevel) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationLogLevel {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationLogLevel, v)
	}
	*j = ConfigurationLogLevel(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConfigurationLogLevel) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationLogLevel {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationLogLevel, v)
	}
	*j = ConfigurationLogLevel(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Configuration) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	type Plain Configuration
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if v, ok := raw["dryRun"]; !ok || v == nil {
		plain.DryRun = false
	}
	if v, ok := raw["gitAuthor"]; !ok || v == nil {
		plain.GitAuthor = ""
	}
	if v, ok := raw["gitCloneOptions"]; !ok || v == nil {
		plain.GitCloneOptions = []string{
			"--filter",
			"blob:none",
		}
	}
	if v, ok := raw["gitCommitMessage"]; !ok || v == nil {
		plain.GitCommitMessage = "changes by saturn-bot"
	}
	if v, ok := raw["gitLogLevel"]; !ok || v == nil {
		plain.GitLogLevel = "warn"
	}
	if v, ok := raw["gitPath"]; !ok || v == nil {
		plain.GitPath = "git"
	}
	if v, ok := raw["githubCacheDisabled"]; !ok || v == nil {
		plain.GithubCacheDisabled = false
	}
	if v, ok := raw["gitlabAddress"]; !ok || v == nil {
		plain.GitlabAddress = "https://gitlab.com"
	}
	if v, ok := raw["logFormat"]; !ok || v == nil {
		plain.LogFormat = "auto"
	}
	if v, ok := raw["logLevel"]; !ok || v == nil {
		plain.LogLevel = "info"
	}
	*j = Configuration(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Configuration) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	type Plain Configuration
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if v, ok := raw["dryRun"]; !ok || v == nil {
		plain.DryRun = false
	}
	if v, ok := raw["gitAuthor"]; !ok || v == nil {
		plain.GitAuthor = ""
	}
	if v, ok := raw["gitCloneOptions"]; !ok || v == nil {
		plain.GitCloneOptions = []string{
			"--filter",
			"blob:none",
		}
	}
	if v, ok := raw["gitCommitMessage"]; !ok || v == nil {
		plain.GitCommitMessage = "changes by saturn-bot"
	}
	if v, ok := raw["gitLogLevel"]; !ok || v == nil {
		plain.GitLogLevel = "warn"
	}
	if v, ok := raw["gitPath"]; !ok || v == nil {
		plain.GitPath = "git"
	}
	if v, ok := raw["githubCacheDisabled"]; !ok || v == nil {
		plain.GithubCacheDisabled = false
	}
	if v, ok := raw["gitlabAddress"]; !ok || v == nil {
		plain.GitlabAddress = "https://gitlab.com"
	}
	if v, ok := raw["logFormat"]; !ok || v == nil {
		plain.LogFormat = "auto"
	}
	if v, ok := raw["logLevel"]; !ok || v == nil {
		plain.LogLevel = "info"
	}
	*j = Configuration(plain)
	return nil
}
