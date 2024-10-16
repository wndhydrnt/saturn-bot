# jsonpath

Downloads a JSON or YAML file from a repository and queries it via one or more JSONPath expressions.
Matches if every query returns at least one result.

## Parameters

### `expressions`

List of JSONPath expressions to query the file.

The implementation is based on [ojg](https://github.com/ohler55/ojg).

Refer to [goessner.net](https://goessner.net/articles/JsonPath/index.html) for an introduction
to JSONPath.

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
  - filter: jsonpath
    params:
      expressions: ["$.dependencies.react"]
      path: package.json
```

```yaml
# Match if the file package.json defines a dependency on the library react and react-dom.
filters:
  - filter: jsonpath
    params:
      expressions:
        - "$.dependencies.react"
        - "$.dependencies.react-dom"
      path: package.json
```

```yaml
# Match if the file package.json doesn't define a dependency on the library react.
filters:
  - filter: jsonpath
    params:
      expressions: ["$.dependencies.react"]
      path: package.json
    reverse: true
```
