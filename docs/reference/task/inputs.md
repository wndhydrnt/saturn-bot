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
    # Optional regular expression that validates the input
    validation: "^Hello|Hola$"
  # Optional input because a default value is set.
  - name: to
    description: Whom to great. # (1)
    default: World
    # Optional list of allowed values
    options:
      - World
      - Mundo
```

1.  The server UI supports Markdown formatting in this field.
    For example, `**bold**` gets rendered as **bold** by the UI.
    Links work too.

Each input is passed to `saturn-bot run`:

```shell
saturn-bot run --input greeting=Hola --input to=Mundo path/to/task/file.yaml
```

!!! note

    saturn-bot skips a task if one or more of its required inputs are missing.

## How to use

Once set, inputs are available as template data.
Plugins have access them too.

For example, the task above can be extended to customize title, description and branch name of the pull request:

```yaml title="Task with input template"
name: Inputs Example
branchName: '{{ .Run.greeting }}-{{ .Run.to }}'
prTitle: '{{ .Run.greeting }} {{ .Run.to }}'
prBody: |
    {{ .Run.greeting }}, {{ .Run.to }}!

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
