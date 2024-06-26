# file

Match a repository if a file exists in the repository.

## Parameters

### `path`

Path of the file in the repository.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```yaml
# Match a repository if it contains the file "hello-world.txt".
filters:
  - filter: file
    params:
      path: "hello-world.txt"
```

```yaml
# Match a repository if it
# contains a YAML file in the directory "config".
filters:
  - filter: file
    params:
      path: "config/*.yaml"
```
