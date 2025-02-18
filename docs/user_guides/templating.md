# Templating

Users can customize the content of text, such as the body and title of a pull request or the name of a branch.

Templates use [Go template notation](https://pkg.go.dev/text/template).

## Template variables

The following variables are available:

| Description                 | Usage                      | Example value                             |
| --------------------------- | -------------------------- | ----------------------------------------- |
| Full name of the repository | `{{.Repository.FullName}}` | `github.com/wndhydrant/saturn-bot`        |
| Host of the repository      | `{{.Repository.Host}}`     | `github.com`                              |
| Name of the repository      | `{{.Repository.Name}}`     | `saturn-bot`                              |
| Owner of the repository     | `{{.Repository.Owner}}`    | `wndhydrnt`                               |
| HTTP URL of the repository  | `{{.Repository.WebUrl}}`   | `http://github.com/wndhydrant/saturn-bot` |
| Name of the task            | `{{.TaskName}}`            | `template-example`                        |

### Run data

Run data is dynamic data that is known only when the task is running.
[Inputs](../reference/task/inputs.md) are available via run data,
as well as data set by [plugins](../reference/task/plugins/index.md).

Run data can be accessed in a template in two ways.

```text
{{ .Run.key }}
```

or, if the key contains a `-`, via the built-in index function:

```text
{{ index .Run "key-with-hyphen" }}
```
