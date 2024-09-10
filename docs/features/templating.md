# Templating

Users can customize the content of text, such as the body and title of a pull request or the name of a branch.

Templates use [Go template notation](https://pkg.go.dev/text/template).

## Template variables

The following variables are available:

| Description | Usage | Example value |
| --- | --- | --- |
| Full name of the repository | `{{.Repository.FullName}}` | `github.com/wndhydrant/saturn-bot` |
| Host of the repository | `{{.Repository.Host}}` | `github.com` |
| Name of the repository | `{{.Repository.Name}}` | `saturn-bot` |
| Owner of the repository | `{{.Repository.Owner}}` | `wndhydrnt` |
| HTTP URL of the repository | `{{.Repository.WebUrl}}` | `http://github.com/wndhydrant/saturn-bot` |
| Name of the task | `{{.TaskName}}` | `template-example` |