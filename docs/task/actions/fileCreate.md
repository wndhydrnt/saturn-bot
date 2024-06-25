# fileCreate

Create or update a file.

## Parameters

### `content`

Content of the file. Mutually exclusive with `contentFromFile`.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `""`     |

### `contentFromFile`

Read the content of the file from the file at the given path. Mutually exclusive with `content`. The path can be absolute or relative to the task file.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `""`     |

### `path`

Path of the file to create in the repository.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `mode`

Mode of the file to create. If the file exists, the file mode gets updated.

| Name     | Value     |
| -------- | --------- |
| Type     | `integer` |
| Required | No        |
| Default  | `644`     |

### `overwrite`

If `true`, overwrite the file if it already exists.

| Name     | Value  |
| -------- | ------ |
| Type     | `bool` |
| Required | No     |
| Default  | `true` |

## Examples

```yaml
# Create the file hello-world.txt with content "Hello World"
# at the root of the repository.
actions:
  - action: fileCreate
    params:
      - content: "Hello World"
        path: "hello-world.txt"
```

```yaml
# Create the file hello-world.txt with content "Hello World"
# at the root of the repository.
# Do nothing if the file already exists.
actions:
  - action: fileCreate
    params:
      - content: "Hello World"
        path: "hello-world.txt"
        overwrite: false
```

```yaml
# Create the file hello-world.txt at the root of the repository.
# Read the content from the file content.txt.
# The path of content.txt is relative to the path of the task.
actions:
  - action: fileCreate
    params:
      - contentFromFile: "./content.txt"
        path: "hello-world.txt"
```

```yaml
# Create the file update.sh at the root of the repository.
# Make the file executable.
actions:
  - action: fileCreate
    params:
      - content: |
          echo "Updating..."
        path: "update.sh"
        mode: 755
```
