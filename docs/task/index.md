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

## commitMessage

[json-path:../../pkg/task/schema/task.schema.json:$.properties.commitMessage.description]

## createOnly

[json-path:../../pkg/task/schema/task.schema.json:$.properties.createOnly.description]

## disabled

[json-path:../../pkg/task/schema/task.schema.json:$.properties.disabled.description]

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

## prBody

[json-path:../../pkg/task/schema/task.schema.json:$.properties.prBody.description]

## prTitle

[json-path:../../pkg/task/schema/task.schema.json:$.properties.prTitle.description]
