# script

Execute a script.

saturn-bot sets the current working directory of the script to the checkout of a repository.

The script has access to the environment variable `TASK_DIR`, which contains the absolute
path to the directory that contains the task file.
It allows the script to load additional files stored next to the task file.

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

```yaml title="Execute script file in repository"
actions:
  - action: script
    params:
      # Path to file must be relative.
      # Otherwise the shell searches in locations of $PATH.
      script: |
        ./file-in-repo.sh
```

```yaml title="Load a file using the TASK_DIR environment variable"
actions:
  - action: script
    params:
      # Read the file "hello.txt", which is located
      # next to the task file and
      # write it to file "hello-repo.txt"
      # in the repository.
      script: |
        cat "${TASK_DIR}/hello.txt" > hello-repo.txt
```

```yaml title="Script file"
# Load the content of the script from a file.
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
