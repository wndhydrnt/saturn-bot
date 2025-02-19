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

	// Configure how to clone git repositories.
	GitUrl ConfigurationGitUrl `json:"gitUrl,omitempty" yaml:"gitUrl,omitempty" mapstructure:"gitUrl,omitempty"`

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

	// Activate Go profiling endpoints for server or worker. The endpoints are
	// available at /debug/pprof/. See https://go.dev/blog/pprof.
	GoProfiling bool `json:"goProfiling,omitempty" yaml:"goProfiling,omitempty" mapstructure:"goProfiling,omitempty"`

	// Path to the Java binary to execute plugins. If not set explicitly, then
	// saturn-bot searches for the binary in $PATH.
	JavaPath string `json:"javaPath,omitempty" yaml:"javaPath,omitempty" mapstructure:"javaPath,omitempty"`

	// Format of log messages.
	LogFormat ConfigurationLogFormat `json:"logFormat,omitempty" yaml:"logFormat,omitempty" mapstructure:"logFormat,omitempty"`

	// Log level of the application.
	LogLevel ConfigurationLogLevel `json:"logLevel,omitempty" yaml:"logLevel,omitempty" mapstructure:"logLevel,omitempty"`

	// Level of logs sent by plugins. Set this to the same value as `logLevel` in
	// order to display the logs of a plugin.
	PluginLogLevel ConfigurationPluginLogLevel `json:"pluginLogLevel,omitempty" yaml:"pluginLogLevel,omitempty" mapstructure:"pluginLogLevel,omitempty"`

	// Address of a Prometheus Pushgateway to send metrics to.
	PrometheusPushgatewayUrl *string `json:"prometheusPushgatewayUrl,omitempty" yaml:"prometheusPushgatewayUrl,omitempty" mapstructure:"prometheusPushgatewayUrl,omitempty"`

	// Path to the Python binary to execute plugins. If not set explicitly, then
	// saturn-bot searches for the binary in $PATH.
	PythonPath string `json:"pythonPath,omitempty" yaml:"pythonPath,omitempty" mapstructure:"pythonPath,omitempty"`

	// Time-to-live of all items in the repository file cache. saturn-bot performs a
	// full update of the cache once the TTL has expired. The format is a Go duration,
	// like `30m` or `12h`.
	RepositoryCacheTtl string `json:"repositoryCacheTtl,omitempty" yaml:"repositoryCacheTtl,omitempty" mapstructure:"repositoryCacheTtl,omitempty"`

	// Turn HTTP access log of server on or off.
	ServerAccessLog bool `json:"serverAccessLog,omitempty" yaml:"serverAccessLog,omitempty" mapstructure:"serverAccessLog,omitempty"`

	// Address of the server in the format `<host>:<port>`.
	ServerAddr string `json:"serverAddr,omitempty" yaml:"serverAddr,omitempty" mapstructure:"serverAddr,omitempty"`

	// URL of the API server. The value is used to populate the `servers` array in the
	// OpenAPI definition.
	ServerBaseUrl string `json:"serverBaseUrl,omitempty" yaml:"serverBaseUrl,omitempty" mapstructure:"serverBaseUrl,omitempty"`

	// Turn compression of responses on or off.
	ServerCompress bool `json:"serverCompress,omitempty" yaml:"serverCompress,omitempty" mapstructure:"serverCompress,omitempty"`

	// If `true`, display executed SQL queries and errors of the database. Useful for
	// debugging.
	ServerDatabaseLog bool `json:"serverDatabaseLog,omitempty" yaml:"serverDatabaseLog,omitempty" mapstructure:"serverDatabaseLog,omitempty"`

	// Path to the sqlite database of the server. If unset, defaults to
	// `<dataDir>/db/saturn-bot.db`.
	ServerDatabasePath string `json:"serverDatabasePath,omitempty" yaml:"serverDatabasePath,omitempty" mapstructure:"serverDatabasePath,omitempty"`

	// If `true`, serves the user interface.
	ServerServeUi bool `json:"serverServeUi,omitempty" yaml:"serverServeUi,omitempty" mapstructure:"serverServeUi,omitempty"`

	// Secret to authenticate webhook requests sent by GitHub.
	ServerWebhookSecretGithub string `json:"serverWebhookSecretGithub,omitempty" yaml:"serverWebhookSecretGithub,omitempty" mapstructure:"serverWebhookSecretGithub,omitempty"`

	// Secret to authenticate webhook requests sent by GitLab. See
	// https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#create-a-webhook
	// for how to set up the token.
	ServerWebhookSecretGitlab string `json:"serverWebhookSecretGitlab,omitempty" yaml:"serverWebhookSecretGitlab,omitempty" mapstructure:"serverWebhookSecretGitlab,omitempty"`

	// Interval at which a worker queries the server to receive new tasks to execute.
	WorkerLoopInterval string `json:"workerLoopInterval,omitempty" yaml:"workerLoopInterval,omitempty" mapstructure:"workerLoopInterval,omitempty"`

	// Number of parallel executions of tasks per worker.
	WorkerParallelExecutions int `json:"workerParallelExecutions,omitempty" yaml:"workerParallelExecutions,omitempty" mapstructure:"workerParallelExecutions,omitempty"`

	// Base URL of the server API to query for new tasks to execute.
	WorkerServerAPIBaseURL string `json:"workerServerAPIBaseURL,omitempty" yaml:"workerServerAPIBaseURL,omitempty" mapstructure:"workerServerAPIBaseURL,omitempty"`
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

type ConfigurationGitUrl string

const ConfigurationGitUrlHttps ConfigurationGitUrl = "https"
const ConfigurationGitUrlSsh ConfigurationGitUrl = "ssh"

var enumValues_ConfigurationGitUrl = []interface{}{
	"https",
	"ssh",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConfigurationGitUrl) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationGitUrl {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationGitUrl, v)
	}
	*j = ConfigurationGitUrl(v)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *ConfigurationGitUrl) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationGitUrl {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationGitUrl, v)
	}
	*j = ConfigurationGitUrl(v)
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

type ConfigurationPluginLogLevel string

const ConfigurationPluginLogLevelDebug ConfigurationPluginLogLevel = "debug"
const ConfigurationPluginLogLevelError ConfigurationPluginLogLevel = "error"
const ConfigurationPluginLogLevelInfo ConfigurationPluginLogLevel = "info"
const ConfigurationPluginLogLevelWarn ConfigurationPluginLogLevel = "warn"

var enumValues_ConfigurationPluginLogLevel = []interface{}{
	"debug",
	"error",
	"info",
	"warn",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConfigurationPluginLogLevel) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationPluginLogLevel {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationPluginLogLevel, v)
	}
	*j = ConfigurationPluginLogLevel(v)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *ConfigurationPluginLogLevel) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConfigurationPluginLogLevel {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConfigurationPluginLogLevel, v)
	}
	*j = ConfigurationPluginLogLevel(v)
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
	if v, ok := raw["gitUrl"]; !ok || v == nil {
		plain.GitUrl = "https"
	}
	if v, ok := raw["githubCacheDisabled"]; !ok || v == nil {
		plain.GithubCacheDisabled = false
	}
	if v, ok := raw["gitlabAddress"]; !ok || v == nil {
		plain.GitlabAddress = "https://gitlab.com"
	}
	if v, ok := raw["goProfiling"]; !ok || v == nil {
		plain.GoProfiling = false
	}
	if v, ok := raw["javaPath"]; !ok || v == nil {
		plain.JavaPath = "java"
	}
	if v, ok := raw["logFormat"]; !ok || v == nil {
		plain.LogFormat = "auto"
	}
	if v, ok := raw["logLevel"]; !ok || v == nil {
		plain.LogLevel = "info"
	}
	if v, ok := raw["pluginLogLevel"]; !ok || v == nil {
		plain.PluginLogLevel = "debug"
	}
	if v, ok := raw["pythonPath"]; !ok || v == nil {
		plain.PythonPath = "python"
	}
	if v, ok := raw["repositoryCacheTtl"]; !ok || v == nil {
		plain.RepositoryCacheTtl = "6h"
	}
	if v, ok := raw["serverAccessLog"]; !ok || v == nil {
		plain.ServerAccessLog = false
	}
	if v, ok := raw["serverAddr"]; !ok || v == nil {
		plain.ServerAddr = ":3035"
	}
	if v, ok := raw["serverBaseUrl"]; !ok || v == nil {
		plain.ServerBaseUrl = "http://localhost:3035"
	}
	if v, ok := raw["serverCompress"]; !ok || v == nil {
		plain.ServerCompress = true
	}
	if v, ok := raw["serverDatabaseLog"]; !ok || v == nil {
		plain.ServerDatabaseLog = false
	}
	if v, ok := raw["serverDatabasePath"]; !ok || v == nil {
		plain.ServerDatabasePath = ""
	}
	if v, ok := raw["serverServeUi"]; !ok || v == nil {
		plain.ServerServeUi = true
	}
	if v, ok := raw["serverWebhookSecretGithub"]; !ok || v == nil {
		plain.ServerWebhookSecretGithub = ""
	}
	if v, ok := raw["serverWebhookSecretGitlab"]; !ok || v == nil {
		plain.ServerWebhookSecretGitlab = ""
	}
	if v, ok := raw["workerLoopInterval"]; !ok || v == nil {
		plain.WorkerLoopInterval = "10s"
	}
	if v, ok := raw["workerParallelExecutions"]; !ok || v == nil {
		plain.WorkerParallelExecutions = 1.0
	}
	if v, ok := raw["workerServerAPIBaseURL"]; !ok || v == nil {
		plain.WorkerServerAPIBaseURL = "http://localhost:3035"
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
	if v, ok := raw["gitUrl"]; !ok || v == nil {
		plain.GitUrl = "https"
	}
	if v, ok := raw["githubCacheDisabled"]; !ok || v == nil {
		plain.GithubCacheDisabled = false
	}
	if v, ok := raw["gitlabAddress"]; !ok || v == nil {
		plain.GitlabAddress = "https://gitlab.com"
	}
	if v, ok := raw["goProfiling"]; !ok || v == nil {
		plain.GoProfiling = false
	}
	if v, ok := raw["javaPath"]; !ok || v == nil {
		plain.JavaPath = "java"
	}
	if v, ok := raw["logFormat"]; !ok || v == nil {
		plain.LogFormat = "auto"
	}
	if v, ok := raw["logLevel"]; !ok || v == nil {
		plain.LogLevel = "info"
	}
	if v, ok := raw["pluginLogLevel"]; !ok || v == nil {
		plain.PluginLogLevel = "debug"
	}
	if v, ok := raw["pythonPath"]; !ok || v == nil {
		plain.PythonPath = "python"
	}
	if v, ok := raw["repositoryCacheTtl"]; !ok || v == nil {
		plain.RepositoryCacheTtl = "6h"
	}
	if v, ok := raw["serverAccessLog"]; !ok || v == nil {
		plain.ServerAccessLog = false
	}
	if v, ok := raw["serverAddr"]; !ok || v == nil {
		plain.ServerAddr = ":3035"
	}
	if v, ok := raw["serverBaseUrl"]; !ok || v == nil {
		plain.ServerBaseUrl = "http://localhost:3035"
	}
	if v, ok := raw["serverCompress"]; !ok || v == nil {
		plain.ServerCompress = true
	}
	if v, ok := raw["serverDatabaseLog"]; !ok || v == nil {
		plain.ServerDatabaseLog = false
	}
	if v, ok := raw["serverDatabasePath"]; !ok || v == nil {
		plain.ServerDatabasePath = ""
	}
	if v, ok := raw["serverServeUi"]; !ok || v == nil {
		plain.ServerServeUi = true
	}
	if v, ok := raw["serverWebhookSecretGithub"]; !ok || v == nil {
		plain.ServerWebhookSecretGithub = ""
	}
	if v, ok := raw["serverWebhookSecretGitlab"]; !ok || v == nil {
		plain.ServerWebhookSecretGitlab = ""
	}
	if v, ok := raw["workerLoopInterval"]; !ok || v == nil {
		plain.WorkerLoopInterval = "10s"
	}
	if v, ok := raw["workerParallelExecutions"]; !ok || v == nil {
		plain.WorkerParallelExecutions = 1.0
	}
	if v, ok := raw["workerServerAPIBaseURL"]; !ok || v == nil {
		plain.WorkerServerAPIBaseURL = "http://localhost:3035"
	}
	*j = Configuration(plain)
	return nil
}
