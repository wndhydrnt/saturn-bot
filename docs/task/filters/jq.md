# jq

Downloads a JSON or YAML file from a repository and queries it via one or more [jq](https://jqlang.github.io/jq/) expressions.
Matches if every query returns at least one result.

## Parameters

### `expressions`

List of jq expressions to query the file.

The implementation is based on [gojq](https://github.com/itchyny/gojq).

Use the [jq playground](https://jqplay.org/) to test expressions.

| Name     | Value      |
| -------- | ---------- |
| Type     | `string[]` |
| Required | **Yes**    |

### `path`

Path to the file in a repository.
Supports JSON and YAML formats.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```json title="package.json"
{
  "dependencies": {
    "react": "^18",
    "react-dom": "^18"
  }
}
```

```yaml
# Match if the file package.json defines a dependency on the library react.
filters:
  - filter: jq
    params:
      expressions: [".dependencies.react"]
      path: package.json
```

```yaml
# Match if the file package.json defines a dependency on the library react and react-dom.
filters:
  - filter: jq
    params:
      expressions:
        - ".dependencies.react"
        - ".dependencies.react-dom"
      path: package.json
```

```yaml
# Match if the file package.json doesn't define a dependency on the library react.
filters:
  - filter: jq
    params:
      expressions: [".dependencies.react"]
      path: package.json
    reverse: true
```

```yaml
# Same as previous example, but not using `reverse`.
filters:
  - filter: jq
    params:
      expressions: [".dependencies.react == null"]
      path: package.json
```
