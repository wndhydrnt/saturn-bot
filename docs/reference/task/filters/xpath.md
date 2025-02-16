# xpath

Downloads an XML document from a repository and queries it via one or more XPath expressions.
Matches if every query returns at least one XML node.

## Parameters

### `expressions`

List of XPath expressions to query the XML document.

Use an online tool such as [XPather](http://xpather.com/) to test the expression.

The [XPath cheatsheet](https://devhints.io/xpath) can be a valuable reference.

| Name     | Value      |
| -------- | ---------- |
| Type     | `string[]` |
| Required | **Yes**    |

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
      expressions: ['/project/dependencies/dependency/artifactId[text()="kotlin-stdlib"]']
      path: pom.xml
```

```yaml
# Match if the file pom.xml defines a dependency
# on the library kotlin-stdlib and kotlinx.coroutines.
filters:
  - filter: xpath
    params:
      expressions:
        - '/project/dependencies/dependency/artifactId[text()="kotlin-stdlib"]'
        - '/project/dependencies/dependency/artifactId[text()="kotlinx-coroutines-core"]'
      path: pom.xml
```

```yaml
# Match if the file pom.xml doesn't define a dependency on the library kotlin-stdlib.
filters:
  - filter: xpath
    params:
      expressions: ['/project/dependencies/dependency/artifactId[text()="kotlin-stdlib"]']
      path: pom.xml
    reverse: true
```
