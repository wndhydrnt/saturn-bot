# exec

Execute a command.

saturn-bot sets the current working directory of the command to the checkout of a repository.

## Parameters

### `args`

Arguments to pass to the command.

| Name     | Value    |
| -------- | -------- |
| Type     | `array`  |
| Subtype  | `string` |
| Required | No       |
| Default  | `[]`     |

### `command`

The command to execute.

The value can be an absolute path to a binary. If the path is relative, it is interpreted as relative to the Task file.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `timeout`

Timeout after which the subprocess is shut down. Waits two minutes by default.

The value is a Go [duration string](https://pkg.go.dev/time#ParseDuration).

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `2m`     |

## Examples

```yaml
# Execute the script "update.sh".
# The script "update.sh" is located in the same directory as the Task file.
actions:
  - action: exec
    params:
      command: "./update.sh"
```

```yaml
# Sleep for a while.
actions:
  - action: exec
    params:
      args: ["5"]
      command: "/bin/sleep"
```

```yaml
# Execute the script "update.sh".
# Stop the process if it did not exit after 15 seconds.
actions:
  - action: exec
    params:
      command: "./update.sh"
      timeout: "15s"
```
