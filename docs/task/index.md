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

## autoCloseAfter

[json-path:../../pkg/task/schema/task.schema.json:$.properties.autoCloseAfter.description]

Examples

```yaml title="Close pull request after being open for 14 days"
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

Examples

```yaml
# Set a custom name.
branchName: "feature/hello-world"
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

## prTitle

[json-path:../../pkg/task/schema/task.schema.json:$.properties.prTitle.description]
