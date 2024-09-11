# script

Execute a script.

saturn-bot sets the current working directory of the script to the checkout of a repository.

To execute a script file located in a repository, use the [exec action](exec.md).

## Parameters

### `script`

The script to execute. Mutually exclusive with [`scriptFromFile`](#scriptfromfile).

Supports [template variables](../../features/templating.md).

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `""`     |

### `scriptFromFile`

The script to execute. Reads the content from a file. The path to the file can be an absolute path or relative to the task file. Mutually exclusive with [`script`](#script).

Supports [template variables](../../features/templating.md).

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `""`     |

### `shell`

The path to the shell that executes the script.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `sh`     |

### `timeout`

Timeout after which the script process is shut down. Waits 10 seconds by default.

The value is a Go [duration string](https://pkg.go.dev/time#ParseDuration).

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `10s`    |

## Examples

```yaml title="Inline script"
actions:
  - action: script
    params:
      script: |
        echo 'hello world' > hello-world.txt
```

```yaml title="Script file"
# Execute the script "example.sh".
# The script "example.sh" is located in the same directory as the Task file.
actions:
  - action: script
    params:
      scriptFromFile: "./example.sh"
```

```yaml title="Template variables"
actions:
  - action: script
    params:
      script: |
        echo '{{.TaskName}}' > task-name.txt
```

```yaml title="Shell"
# Use "bash" to execute the script
actions:
  - action: script
    params:
      script: |
        echo 'hello world' > hello-world.txt
      shell: "/bin/bash"
```

```yaml title="Timeout"
actions:
  - action: script
    params:
      script: |
        sleep 30
        echo 'hello world' > hello-world.txt
      # Increase the timeout.
      timeout: "1m"
```
