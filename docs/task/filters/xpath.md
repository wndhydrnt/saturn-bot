# xpath

Downloads an XML document from a repository and queries it via an XPath expression.
Matches if the query returns at least one XML node.

## Parameters

### `expression`

The XPath expression to use for querying the XML document.

Use an online tool such as [XPather](http://xpather.com/) to test the expression.

The [XPath cheatsheet](https://devhints.io/xpath) can be a valuable reference.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `path`

Path to the XML document in a repository.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```yaml
# Match if the file pom.xml defines a dependency on the library kotlin-stdlib.
filters:
  - filter: xpath
    params:
      expression: '/project/dependencies/dependency/artifactId[text()="kotlin-stdlib"]'
      path: "pom.xml
```

```yaml
# Match if the file pom.xml doesn't define a dependency on the library kotlin-stdlib.
filters:
  - filter: xpath
    params:
      expression: '/project/dependencies/dependency/artifactId[text()="kotlin-stdlib"]'
      path: "pom.xml
    reverse: true
```
