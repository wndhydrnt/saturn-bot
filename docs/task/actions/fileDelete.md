# fileDelete

Delete a file.

## Parameters

### `path`

Path of the file to create in the repository.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```yaml
# Delete the file hello-world.txt in the root of the repository.
actions:
  - action: fileDelete
    params:
      - path: "hello-world.txt"
```
