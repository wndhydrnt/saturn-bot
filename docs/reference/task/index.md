# Task

```yaml title="Example"
# yaml-language-server: $schema=https://saturn-bot.readthedocs.io/en/latest/schemas/task.schema.json
name: "saturn-bot Hello World"
prTitle: "saturn-bot Hello World"
prBody: |
  saturn-bot Quickstart.

  This pull request creates the file `hello-world.txt`.

# Filters tell saturn-bot which repositories to modify.
filters:
  - filter: repository
    params:
      host: github.com
      owner: wndhydrnt # Replace with your owner
      name: saturn-bot-example # Replace with your repository

# Actions tell saturn-bot how to modify each repository.
actions:
  - action: fileCreate
    params:
      content: "Hello World"
      path: "hello-world.txt"
```

## actions

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.actions.description]

All available actions can be found [here](./actions/index.md).

## active

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.active.description]

Examples

```yaml
# Activate the task (default)
active: true
```

```yaml
# Deactivate the task
active: false
```

## assignees

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.assignees.description]

!!! note

    If the task previously defined assignees and the list is set back to empty
    then saturn-bot doesn't remove the assignees from a pull request.

Examples

```yaml title="Set assignees"
assignees:
  - ellie
  - joel
```

## autoCloseAfter

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.autoCloseAfter.description]

Examples

```yaml title="Close pull request after it being open for 14 days"
autoCloseAfter: 1209600
```

```yaml title="Deactivate auto-close, the default"
autoCloseAfter: 0
```

## autoMerge

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.autoMerge.description]

Examples

```yaml
# Enable auto-merge behavior
autoMerge: true
```

```yaml
# Disable auto-merge behavior
autoMerge: false
```

## autoMergeAfter

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.autoMergeAfter.description]

Examples

```yaml
# Automatically merges pull requests once they’ve been open for 30 min.
autoMergeAfter: 30m
```

## branchName

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.branchName.description]

Supports [templating](../../user_guides/templating.md).

Some git hosts restrict the maximum length of branch names.
The branch name is automatically cut to 230 characters.

Examples

```yaml title="Set a custom branch"
branchName: "feature/hello-world"
```

```yaml title="Use a template variable"
branchName: "feature/{{.TaskName}}"
```

## changeLimit

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.changeLimit.description]

```yaml title="Allow up to 5 pull requests combined to be created or merged in one run of saturn-bot"
changeLimit: 5
```

```yaml title="Disable the feature"
changeLimit: 0
```

## commitMessage

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.commitMessage.description]

## createOnly

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.createOnly.description]

## filters

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.filters.description]

All available filters can be found [here](./filters/index.md).

## inputs

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.inputs.description]

[Inputs](./inputs.md) provides more details on how to use them.

## keepBranchAfterMerge

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.keepBranchAfterMerge.description]

## labels

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.labels.description]

## maxOpenPRs

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.maxOpenPRs.description]

```yaml title="Allow 5 pull requests to be open at the same time"
maxOpenPRs: 5
```

```yaml title="Disable the feature"
maxOpenPRs: 0
```

## mergeOnce

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.mergeOnce.description]

## metricLabels

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.metricLabels.description]

```yaml title="Define a custom label"
metricLabels:
  task_owner: some-team
```

## name

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.name.description]

## plugins

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.plugins.description]

```yaml title="Register a plugin"
plugins:
  - path: ./example # Plugin binary to execute. Path is relative to the task file.
    configuration:
      # Arbitrary configuration to pass to the plugin.
      message: "Hello Plugin"
```

Learn more about how to create plugins in the [documentation](plugins/index.md).

## prBody

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.prBody.description]

Supports [templating](../../user_guides/templating.md).

Examples

```yaml title="Custom pull request body"
prBody: |
  Describe what the change does.

  Supports multi-lines.
```

```yaml title="Use a template variable"
prBody: |
  This pull request modifies repository {{.Repository.FullName}}.
```

## prTitle

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.prTitle.description]

Supports [templating](../../user_guides/templating.md).

Examples

```yaml title="Custom pull request title"
prTitle: "feat: Custom title"
```

```yaml title="Use a template variable"
prTitle: "Apply task {{.TaskName}}"
```

## pushToDefaultBranch

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.pushToDefaultBranch.description]

Defaults to `false`.

Examples

```yaml title="Create a pull request if task changes content"
pushToDefaultBranch: false
```

```yaml title="Push changes to the default branch"
pushToDefaultBranch: true
```

## reviewers

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.reviewers.description]

!!! note

    If the task previously defined reviewers and the list is set back to empty
    then saturn-bot doesn't remove the reviewers from a pull request.

Examples

```yaml title="Set reviewers"
reviewers:
  - ellie
  - joel
```

## trigger

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.trigger.description]

### cron

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.trigger.properties.cron.description]

[crontab.guru](https://crontab.guru) supports with writing the cron schedule expression.

```yaml title="Define a cron trigger"
trigger:
  cron: "0 8,13 * * *"
```

### webhook

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.trigger.properties.webhook.description]

[Trigger a task when a webhook is received](../../user_guides/webhook.md) describes how to set up a webhook.

See [Webhook](./webhook.md) for reference.

#### delay

[json-path:../../../pkg/task/schema/task.schema.json:$.properties.trigger.properties.webhook.properties.delay.description]

```yaml title="Delay the task by 5 minutes"
trigger:
  webhook:
    delay: 300
    # ... other webhook settings
```
