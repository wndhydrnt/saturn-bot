// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package schema

import "encoding/json"
import "fmt"
import yaml "gopkg.in/yaml.v3"

// An action tells saturn-bot how to modify a repository.
type Action struct {
	// Identifier of the action.
	Action string `json:"action" yaml:"action" mapstructure:"action"`

	// Key/value pairs passed as parameters to the action.
	Params ActionParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

// Key/value pairs passed as parameters to the action.
type ActionParams map[string]interface{}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Action) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["action"]; raw != nil && !ok {
		return fmt.Errorf("field action in Action: required")
	}
	type Plain Action
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Action(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Action) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["action"]; raw != nil && !ok {
		return fmt.Errorf("field action in Action: required")
	}
	type Plain Action
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Action(plain)
	return nil
}

type Filter struct {
	// Identifier of the filter.
	Filter string `json:"filter" yaml:"filter" mapstructure:"filter"`

	// Key/value pairs passed as parameters to the filter.
	Params FilterParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`

	// Reverse the result of the filter, i.e. negate it.
	Reverse bool `json:"reverse,omitempty" yaml:"reverse,omitempty" mapstructure:"reverse,omitempty"`
}

// Key/value pairs passed as parameters to the filter.
type FilterParams map[string]interface{}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Filter) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["filter"]; raw != nil && !ok {
		return fmt.Errorf("field filter in Filter: required")
	}
	type Plain Filter
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if v, ok := raw["reverse"]; !ok || v == nil {
		plain.Reverse = false
	}
	*j = Filter(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Filter) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["filter"]; raw != nil && !ok {
		return fmt.Errorf("field filter in Filter: required")
	}
	type Plain Filter
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if v, ok := raw["reverse"]; !ok || v == nil {
		plain.Reverse = false
	}
	*j = Filter(plain)
	return nil
}

type GithubTrigger struct {
	// Experimental: GitHub webhook event, like push. See
	// https://docs.github.com/en/webhooks/webhook-events-and-payloads for a list of
	// all available events.
	Event *string `json:"event,omitempty" yaml:"event,omitempty" mapstructure:"event,omitempty"`

	// Experimental: jq expressions to apply to the body of the webhook. If all
	// expressions match the content of the webhook then a new run of the task is
	// scheduled.
	Filters []string `json:"filters,omitempty" yaml:"filters,omitempty" mapstructure:"filters,omitempty"`
}

type GitlabTrigger struct {
	// Experimental: GitLab webhook event, like push. See
	// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html for a
	// list of all available events.
	Event *string `json:"event,omitempty" yaml:"event,omitempty" mapstructure:"event,omitempty"`

	// Experimental: jq expressions to apply to the body of the webhook. If all
	// expressions match the content of the webhook then a new run of the task is
	// scheduled.
	Filters []string `json:"filters,omitempty" yaml:"filters,omitempty" mapstructure:"filters,omitempty"`
}

// A input allows customizing a task at runtime.
type Input struct {
	// Default value to use if no input has been set via the command-line.
	Default *string `json:"default,omitempty" yaml:"default,omitempty" mapstructure:"default,omitempty"`

	// Text that describes the input value.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// Key that identifies the input. Set via the command-line to set the input value.
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// If not empty, a list of possible values for the input.
	Options []string `json:"options,omitempty" yaml:"options,omitempty" mapstructure:"options,omitempty"`

	// If not empty, a regular expression that validates the value of the input.
	Validation *string `json:"validation,omitempty" yaml:"validation,omitempty" mapstructure:"validation,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Input) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in Input: required")
	}
	type Plain Input
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Input(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Input) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in Input: required")
	}
	type Plain Input
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Input(plain)
	return nil
}

// A plugin extends saturn-bot and allows custom filtering or modification of
// repositories.
type Plugin struct {
	// Key/value pairs that hold additional configuration for the plugin. Sent to the
	// plugin once on startup.
	Configuration PluginConfiguration `json:"configuration,omitempty" yaml:"configuration,omitempty" mapstructure:"configuration,omitempty"`

	// Path corresponds to the JSON schema field "path".
	Path string `json:"path" yaml:"path" mapstructure:"path"`
}

// Key/value pairs that hold additional configuration for the plugin. Sent to the
// plugin once on startup.
type PluginConfiguration map[string]string

// UnmarshalJSON implements json.Unmarshaler.
func (j *Plugin) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["path"]; raw != nil && !ok {
		return fmt.Errorf("field path in Plugin: required")
	}
	type Plain Plugin
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Plugin(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Plugin) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["path"]; raw != nil && !ok {
		return fmt.Errorf("field path in Plugin: required")
	}
	type Plain Plugin
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Plugin(plain)
	return nil
}

type Task struct {
	// List of actions that modify a repository.
	Actions []Action `json:"actions,omitempty" yaml:"actions,omitempty" mapstructure:"actions,omitempty"`

	// Set to `false` to temporarily deactivate the task and prevent it from
	// executing.
	Active bool `json:"active,omitempty" yaml:"active,omitempty" mapstructure:"active,omitempty"`

	// A list of usernames to set as assignees of a pull request.
	Assignees []string `json:"assignees,omitempty" yaml:"assignees,omitempty" mapstructure:"assignees,omitempty"`

	// Automatically close a pull request if it has been unmerged for the duration.
	// Format is seconds. Set to `0`, the default, to deactivate.
	AutoCloseAfter int `json:"autoCloseAfter,omitempty" yaml:"autoCloseAfter,omitempty" mapstructure:"autoCloseAfter,omitempty"`

	// Merge a pull request automatically if all checks have passed and all approvals
	// have been given.
	AutoMerge bool `json:"autoMerge,omitempty" yaml:"autoMerge,omitempty" mapstructure:"autoMerge,omitempty"`

	// If set, automatically merge the pull request after it has been open for the
	// specified amount of time. Only applied if `autoMerge` is `true`. The value is a
	// Go duration, like 5m or 1h.
	AutoMergeAfter string `json:"autoMergeAfter,omitempty" yaml:"autoMergeAfter,omitempty" mapstructure:"autoMergeAfter,omitempty"`

	// If set, used as the name of the branch to commit changes to. Defaults to an
	// auto-generated name if not set.
	BranchName string `json:"branchName,omitempty" yaml:"branchName,omitempty" mapstructure:"branchName,omitempty"`

	// Number of pull requests to create or merge (combined) in one run. Useful to
	// reduce strain on a system caused by, for example, many CI/CD jobs created at
	// the same time.
	ChangeLimit int `json:"changeLimit,omitempty" yaml:"changeLimit,omitempty" mapstructure:"changeLimit,omitempty"`

	// If set, used as the message when changes get committed. Defaults to an
	// auto-generated message if not set.
	CommitMessage string `json:"commitMessage,omitempty" yaml:"commitMessage,omitempty" mapstructure:"commitMessage,omitempty"`

	// Create pull requests only. Don't attempt to update a pull request on a
	// subsequent run.
	CreateOnly bool `json:"createOnly,omitempty" yaml:"createOnly,omitempty" mapstructure:"createOnly,omitempty"`

	// Filters make saturn-bot pick the repositories to which it applies the task.
	Filters []Filter `json:"filters,omitempty" yaml:"filters,omitempty" mapstructure:"filters,omitempty"`

	// Inputs allows customizing a task at runtime.
	Inputs []Input `json:"inputs,omitempty" yaml:"inputs,omitempty" mapstructure:"inputs,omitempty"`

	// If `true`, keep the branch after a pull request has been merged.
	KeepBranchAfterMerge bool `json:"keepBranchAfterMerge,omitempty" yaml:"keepBranchAfterMerge,omitempty" mapstructure:"keepBranchAfterMerge,omitempty"`

	// List of labels to attach to a pull request.
	Labels []string `json:"labels,omitempty" yaml:"labels,omitempty" mapstructure:"labels,omitempty"`

	// The number of pull requests that can be open at the same time. 0 disables the
	// feature.
	MaxOpenPRs int `json:"maxOpenPRs,omitempty" yaml:"maxOpenPRs,omitempty" mapstructure:"maxOpenPRs,omitempty"`

	// If `true`, no new pull request is being created if a previous pull request has
	// been merged for this task.
	MergeOnce bool `json:"mergeOnce,omitempty" yaml:"mergeOnce,omitempty" mapstructure:"mergeOnce,omitempty"`

	// The name of the task. Used as an identifier.
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// List of plugins to start for the task.
	Plugins []Plugin `json:"plugins,omitempty" yaml:"plugins,omitempty" mapstructure:"plugins,omitempty"`

	// If set, used as the body of the pull request.
	PrBody string `json:"prBody,omitempty" yaml:"prBody,omitempty" mapstructure:"prBody,omitempty"`

	// If set, used as the title of the pull request.
	PrTitle string `json:"prTitle,omitempty" yaml:"prTitle,omitempty" mapstructure:"prTitle,omitempty"`

	// If `true`, push changes directly to the default branch, like "main". If
	// `false`, create a pull request to submit changes.
	PushToDefaultBranch bool `json:"pushToDefaultBranch,omitempty" yaml:"pushToDefaultBranch,omitempty" mapstructure:"pushToDefaultBranch,omitempty"`

	// A list of usernames to set as reviewers of the pull request.
	Reviewers []string `json:"reviewers,omitempty" yaml:"reviewers,omitempty" mapstructure:"reviewers,omitempty"`

	// Define when the task gets executed. Only relevant in server mode.
	Trigger *TaskTrigger `json:"trigger,omitempty" yaml:"trigger,omitempty" mapstructure:"trigger,omitempty"`
}

// Define when the task gets executed. Only relevant in server mode.
type TaskTrigger struct {
	// Trigger the task based on a cron schedule.
	Cron *string `json:"cron,omitempty" yaml:"cron,omitempty" mapstructure:"cron,omitempty"`

	// Execute the task when the server receives a webhook.
	Webhook *TaskTriggerWebhook `json:"webhook,omitempty" yaml:"webhook,omitempty" mapstructure:"webhook,omitempty"`
}

// Execute the task when the server receives a webhook.
type TaskTriggerWebhook struct {
	// Delay the execution of the task, in seconds, after the webhook has been
	// received by the server.
	Delay int `json:"delay,omitempty" yaml:"delay,omitempty" mapstructure:"delay,omitempty"`

	// Experimental: Execute the task when the server receives a webhook from GitHub.
	Github []GithubTrigger `json:"github,omitempty" yaml:"github,omitempty" mapstructure:"github,omitempty"`

	// Experimental: Execute the task when the server receives a webhook from GitLab.
	Gitlab []GitlabTrigger `json:"gitlab,omitempty" yaml:"gitlab,omitempty" mapstructure:"gitlab,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *TaskTriggerWebhook) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	type Plain TaskTriggerWebhook
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if v, ok := raw["delay"]; !ok || v == nil {
		plain.Delay = 0.0
	}
	*j = TaskTriggerWebhook(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *TaskTriggerWebhook) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	type Plain TaskTriggerWebhook
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if v, ok := raw["delay"]; !ok || v == nil {
		plain.Delay = 0.0
	}
	*j = TaskTriggerWebhook(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Task) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in Task: required")
	}
	type Plain Task
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if v, ok := raw["active"]; !ok || v == nil {
		plain.Active = true
	}
	if v, ok := raw["autoCloseAfter"]; !ok || v == nil {
		plain.AutoCloseAfter = 0.0
	}
	if v, ok := raw["autoMerge"]; !ok || v == nil {
		plain.AutoMerge = false
	}
	if v, ok := raw["autoMergeAfter"]; !ok || v == nil {
		plain.AutoMergeAfter = ""
	}
	if v, ok := raw["branchName"]; !ok || v == nil {
		plain.BranchName = ""
	}
	if v, ok := raw["changeLimit"]; !ok || v == nil {
		plain.ChangeLimit = 0.0
	}
	if v, ok := raw["commitMessage"]; !ok || v == nil {
		plain.CommitMessage = ""
	}
	if v, ok := raw["createOnly"]; !ok || v == nil {
		plain.CreateOnly = false
	}
	if v, ok := raw["keepBranchAfterMerge"]; !ok || v == nil {
		plain.KeepBranchAfterMerge = false
	}
	if v, ok := raw["maxOpenPRs"]; !ok || v == nil {
		plain.MaxOpenPRs = 0.0
	}
	if v, ok := raw["mergeOnce"]; !ok || v == nil {
		plain.MergeOnce = false
	}
	if v, ok := raw["prBody"]; !ok || v == nil {
		plain.PrBody = ""
	}
	if v, ok := raw["prTitle"]; !ok || v == nil {
		plain.PrTitle = ""
	}
	if v, ok := raw["pushToDefaultBranch"]; !ok || v == nil {
		plain.PushToDefaultBranch = false
	}
	*j = Task(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Task) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in Task: required")
	}
	type Plain Task
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if v, ok := raw["active"]; !ok || v == nil {
		plain.Active = true
	}
	if v, ok := raw["autoCloseAfter"]; !ok || v == nil {
		plain.AutoCloseAfter = 0.0
	}
	if v, ok := raw["autoMerge"]; !ok || v == nil {
		plain.AutoMerge = false
	}
	if v, ok := raw["autoMergeAfter"]; !ok || v == nil {
		plain.AutoMergeAfter = ""
	}
	if v, ok := raw["branchName"]; !ok || v == nil {
		plain.BranchName = ""
	}
	if v, ok := raw["changeLimit"]; !ok || v == nil {
		plain.ChangeLimit = 0.0
	}
	if v, ok := raw["commitMessage"]; !ok || v == nil {
		plain.CommitMessage = ""
	}
	if v, ok := raw["createOnly"]; !ok || v == nil {
		plain.CreateOnly = false
	}
	if v, ok := raw["keepBranchAfterMerge"]; !ok || v == nil {
		plain.KeepBranchAfterMerge = false
	}
	if v, ok := raw["maxOpenPRs"]; !ok || v == nil {
		plain.MaxOpenPRs = 0.0
	}
	if v, ok := raw["mergeOnce"]; !ok || v == nil {
		plain.MergeOnce = false
	}
	if v, ok := raw["prBody"]; !ok || v == nil {
		plain.PrBody = ""
	}
	if v, ok := raw["prTitle"]; !ok || v == nil {
		plain.PrTitle = ""
	}
	if v, ok := raw["pushToDefaultBranch"]; !ok || v == nil {
		plain.PushToDefaultBranch = false
	}
	*j = Task(plain)
	return nil
}
