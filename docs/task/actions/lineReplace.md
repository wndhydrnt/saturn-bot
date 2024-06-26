# lineReplace

Replace one or more lines in a file.

## Parameters

### `line`

The new line that replaces any matched lines.

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

### `regexp`

Regular expression that gets matched against each line in the file. If the regular expression matches, the line is replaced with the value of [`line`](#line).

| Name     | Value                           |
| -------- | ------------------------------- |
| Type     | `string`                        |
| Required | **Yes**, if `search` isn't set. |
| Default  | `""`                            |

### `search`

Search string that gets matched against each line in the file. If the line equals the search string, the line is replaced with the value of [`line`](#line).

| Name     | Value                           |
| -------- | ------------------------------- |
| Type     | `string`                        |
| Required | **Yes**, if `regexp` isn't set. |
| Default  | `""`                            |

## Examples

```yaml
# Replace the line "Hello Everyone"
# with the line "Hello World" in the file "hello-world.txt".
actions:
  - action: lineReplace
    params:
      - line: "Hello World"
        path: "hello-world.txt"
        search: "Hello Everyone"
```

```yaml
# Turn each line that equals "Hello World"
# into "World Hello" in the file "hello-world.txt".
actions:
  - action: lineReplace
    params:
      - line: "${2} ${1}"
        path: "hello-world.txt"
        regexp: "(Hello)\s(World)"
```
