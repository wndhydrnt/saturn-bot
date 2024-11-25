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

[json-path:../../pkg/task/schema/task.schema.json:$.properties.actions.description]

All available actions can be found [here](./actions/index.md).

## active

[json-path:../../pkg/task/schema/task.schema.json:$.properties.active.description]

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

[json-path:../../pkg/task/schema/task.schema.json:$.properties.assignees.description]

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

[json-path:../../pkg/task/schema/task.schema.json:$.properties.autoCloseAfter.description]

Examples

```yaml title="Close pull request after it being open for 14 days"
autoCloseAfter: 1209600
```

```yaml title="Deactivate auto-close, the default"
autoCloseAfter: 0
```

## autoMerge

[json-path:../../pkg/task/schema/task.schema.json:$.properties.autoMerge.description]

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

[json-path:../../pkg/task/schema/task.schema.json:$.properties.autoMergeAfter.description]

Examples

```yaml
# Merge pull request automatically.
autoMerge: true
```

```yaml
# Don't merge pull request automatically.
autoMerge: false
```

## branchName

[json-path:../../pkg/task/schema/task.schema.json:$.properties.branchName.description]

Supports [templating](../features/templating.md).

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

[json-path:../../pkg/task/schema/task.schema.json:$.properties.changeLimit.description]

```yaml title="Allow up to 5 pull requests combined to be created or merged in one run of saturn-bot"
changeLimit: 5
```

```yaml title="Disable the feature"
changeLimit: 0
```

## commitMessage

[json-path:../../pkg/task/schema/task.schema.json:$.properties.commitMessage.description]

## createOnly

[json-path:../../pkg/task/schema/task.schema.json:$.properties.createOnly.description]

## filters

[json-path:../../pkg/task/schema/task.schema.json:$.properties.filters.description]

All available filters can be found [here](./filters/index.md).

## inputs

[json-path:../../pkg/task/schema/task.schema.json:$.properties.inputs.description]

[Inputs](./inputs.md) provides more details on how to use them.

## keepBranchAfterMerge

[json-path:../../pkg/task/schema/task.schema.json:$.properties.keepBranchAfterMerge.description]

## labels

[json-path:../../pkg/task/schema/task.schema.json:$.properties.labels.description]

## maxOpenPRs

[json-path:../../pkg/task/schema/task.schema.json:$.properties.maxOpenPRs.description]

```yaml title="Allow 5 pull requests to be open at the same time"
maxOpenPRs: 5
```

```yaml title="Disable the feature"
maxOpenPRs: 0
```

## mergeOnce

[json-path:../../pkg/task/schema/task.schema.json:$.properties.mergeOnce.description]

## name

[json-path:../../pkg/task/schema/task.schema.json:$.properties.name.description]

## plugins

[json-path:../../pkg/task/schema/task.schema.json:$.properties.plugins.description]

```yaml title="Register a plugin"
plugins:
  - path: ./example # Plugin binary to execute. Path is relative to the task file.
    configuration:
      # Arbitrary configuration to pass to the plugin.
      message: "Hello Plugin"
```

Learn more about how to create plugins in the [documentation](plugins/index.md).

## prBody

[json-path:../../pkg/task/schema/task.schema.json:$.properties.prBody.description]

Supports [templating](../features/templating.md).

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

[json-path:../../pkg/task/schema/task.schema.json:$.properties.prTitle.description]

Supports [templating](../features/templating.md).

Examples

```yaml title="Custom pull request title"
prTitle: "feat: Custom title"
```

```yaml title="Use a template variable"
prTitle: "Apply task {{.TaskName}}"
```

## reviewers

[json-path:../../pkg/task/schema/task.schema.json:$.properties.reviewers.description]

!!! note

    If the task previously defined reviewers and the list is set back to empty
    then saturn-bot doesn't remove the reviewers from a pull request.

Examples

```yaml title="Set reviewers"
reviewers:
  - ellie
  - joel
```

## schedule

[json-path:../../pkg/task/schema/task.schema.json:$.properties.schedule.description]

Helps constrain the number of executions of a task. For example, if the task is too "noisy" or should run only once a week or once a month.

!!! note

    This setting uses Cron syntax to define a time range within which a task can be executed.
    It doesn't define the exact point in time at which the task gets executed.

```yaml title="Allow execution at any time on a Wednesday"
schedule: "* * * * WED"
```

```yaml title="Allow execution at any time on the 7th day of each month"
schedule: "* * 7 * *"
```

```yaml title="Allow execution every day after 14:00"
schedule: "* 14 * * *"
```
