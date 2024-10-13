# jsonpath

Downloads a JSON or YAML file from a repository and queries it via a JSONPath expression.
Matches if the query returns at least one result.

## Parameters

### `expression`

The JSONPath expression to query the file.

The implementation is based on [ojg](https://github.com/ohler55/ojg).

Refer to [goessner.net](https://goessner.net/articles/JsonPath/index.html) for an introduction
to JSONPath.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

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
    "react": "^18"
  }
}
```

```yaml
# Match if the file package.json defines a dependency on the library react.
filters:
  - filter: jsonpath
    params:
      expression: '$.dependencies.react'
      path: package.json
```

```yaml
# Match if the file package.json doesn't define a dependency on the library react.
filters:
  - filter: jsonpath
    params:
      expression: '$.dependencies.react'
      path: package.json
    reverse: true
```
