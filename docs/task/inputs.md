# Inputs

Inputs allow customizing a task at runtime.

## How to set

A task file declares all expected inputs:

```yaml title="Task"
name: Inputs Example

# ... other settings

inputs:
  # Required input, no default value.
  - name: greeting
    description: How to greet.
  # Optional input because a default value is set.
  - name: to
    description: Whom to great.
    default: World
```

Each input is passed to `saturn-bot run`:

```shell
saturn-bot run --input greeting=Hola --input to=mundo path/to/task/file.yaml
```

!!! note

    saturn-bot skips a task if one or more of its required inputs are missing.

## How to use

Once set, inputs are available as template data.
Plugins have access them too.

For example, the task above can be extended to customize title, description and branch name of the pull request:

```yaml title="Task with input template"
name: Inputs Example
branchName: '{{ .Run["greeting"] }}-{{ .Run["to"] }}'
prTitle: '{{ .Run["greeting"] }} {{ .Run["to"] }}'
prBody: |
    {{ .Run["greeting"] }}, {{ .Run["to"] }}!

# ... other settings

inputs:
  # Required input, no default value.
  - name: greeting
    description: How to greet.
  # Optional input because a default value is set.
  - name: to
    description: Whom to great.
    default: World
```

Given the task file above and the inputs `--input greeting=Hello` and `--input to=World`
then `prTitle` is rendered to:

```text
Hello World
```
