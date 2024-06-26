# repository

Match a repository if its host, owner and name match.

## Parameters

### `host`

The host of the repository, like `github.com` or `gitlab.com`.

Value can be a regular expression.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `owner`

The owner of the repository.

For GitHub, this is the user or organization that owns the repository.

For GitLab, this is the user or group that owns the repository.

Value can be a regular expression.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

### `name`

The name of the repository.

Value can be a regular expression.

| Name     | Value    |
| -------- | -------- |
| Type     | `string` |
| Required | **Yes**  |

## Examples

```yaml
# Match a repository on GitHub.
filters:
  - filter: repository
    params:
      host: "github.com"
      owner: "wndhydrnt"
      name: "saturn-bot"
```

```yaml
# Match all repositories on GitLab in a subgroup.
filters:
  - filter: repository
    params:
      host: "gitlab.com"
      owner: "gitlab-org/ci-cd/tests"
      name: ".+"
```
