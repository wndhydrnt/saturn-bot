# lineDelete

Delete lines in a file.

## Parameters

### `path`

Path of the file.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `regexp`

Regular expression that gets matched against each line in the file. If the regular expression matches, the line is deleted.

| Name     | Value                           |
| -------- | ------------------------------- |
| Type     | `string`                        |
| Required | **Yes**, if `search` isn't set. |
| Default  | `""`                            |

### `search`

Search string that gets matched against each line in the file. If the line equals the search string, the line is deleted.

| Name     | Value                           |
| -------- | ------------------------------- |
| Type     | `string`                        |
| Required | **Yes**, if `regexp` isn't set. |
| Default  | `""`                            |

## Examples

```yaml
# Delete every line that equals "Hello World" in the file "hello-world.txt".
actions:
  - action: lineDelete
    params:
      path: "hello-world.txt"
      search: "Hello World"
```

```yaml
# Delete every line that starts with "Hello" in the file "hello-world.txt".
actions:
  - action: lineDelete
    params:
      path: "hello-world.txt"
      regexp: "Hello.+"
```
