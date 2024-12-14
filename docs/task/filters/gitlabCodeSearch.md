# gitlabCodeSearch

Executes a [code search](https://docs.gitlab.com/ee/user/search/advanced_search.html#code-search) query against GitLab.
Every repository in the result set returned by GitLab is considered a match.

## Parameters

### `groupID`

Optional ID to limit the search to projects within a group.

The is either an integer of the group or the path to the group.

| Name     | Value                 |
| -------- | --------------------- |
| Type     | `integer` or `string` |
| Required | No                    |

### `query`

The code search query to execute.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```yaml title="Match all repositories that contain the file go.mod"
  filters:
    - filter: gitlabCodeSearch
      params:
        query: "file:go.mod"
```

```yaml title="Match all repositories that contain the file go.mod in group gitlab-org"
filters:
  - filter: gitlabCodeSearch
    params:
      groupID: "gitlab-org"
      query: "file:go.mod"
```

[Search result of the query](https://gitlab.com/search?group_id=9970&scope=blobs&search=file%3Ago.mod)

```yaml title="Same as previous example but using the unique ID of the group"
filters:
  - filter: gitlabCodeSearch
    params:
      groupID: 9970
      query: "file:go.mod"
```

[Search result of the query](https://gitlab.com/search?group_id=9970&scope=blobs&search=file%3Ago.mod)
