# fileContent

Match a repository if the content of a file in the repository matches.

## Parameters

### `path`

Path of the file in the repository.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `regexp`

Regular expression to match against the content of the file.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```yaml
# Match a repository if it the file "hello-world.txt"
# contains the string "Hello".
filters:
  - filter: fileContent
    params:
      - path: "hello-world.txt"
        regexp: "Hello"
```
