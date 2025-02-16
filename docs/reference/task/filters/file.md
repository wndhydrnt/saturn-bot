# file

Match a repository if one or more files exist in the repository.

## Parameters

### `op`

Operator to connect the entries in `paths`. Can be `and` or `or`.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | No       |
| Default  | `and`    |

### `paths`

One or more paths to files in the repository.

| Name     | Value      |
| -------- | ---------- |
| Type     | `string[]` |
| Required | **Yes**    |

## Examples

```yaml
# Match a repository if it contains the file "hello-world.txt".
filters:
  - filter: file
    params:
      paths: ["hello-world.txt"]
```

```yaml
# Match a repository if it
# contains a YAML file in the directory "config".
filters:
  - filter: file
    params:
      paths: ["config/*.yaml"]
```

```yaml
# Match a repository if it
# contains the files "pom.xml" and "README.md".
filters:
  - filter: file
    params:
      paths: ["pom.xml", "README.md"]
```

```yaml
# Match a repository if it
# contains either "pom.xml" or "settings.gradle".
filters:
  - filter: file
    params:
      op: "or"
      paths: ["pom.xml", "settings.gradle"]
```
