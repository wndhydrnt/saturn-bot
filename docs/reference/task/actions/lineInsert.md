# lineInsert

Insert a line in a file. The line can be added at the beginning or the end of the file.

## Parameters

### `insertAt`

Define where to insert the line. `EOF` to insert at the end of the file. `BOF` to insert at the beginning of the file.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `EOF`    |

### `line`

The line to insert.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `path`

Path of the file.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```yaml
# Add the line "Hello End" at the end of the file "hello-world.txt".
actions:
  - action: lineInsert
    params:
      line: "Hello End"
      path: "hello-world.txt"
```

```yaml
# Add the line "Hello Beginning" at the beginning of the file "hello-world.txt".
actions:
  - action: lineInsert
    params:
      insertAt: "BOF"
      line: "Hello Beginning"
      path: "hello-world.txt"
```
