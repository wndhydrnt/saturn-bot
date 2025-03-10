# Contributing

## Code generation

This repository uses code generation via `go generate`.

To update all code and documentation, execute

```shell
go generate ./...
```

Tools that generate code are installed via [./tools/tools.go](./tools/tools.go).
[This blog post](https://www.jvt.me/posts/2024/09/30/go-tools-module/) explains the pattern.

## Database migrations

Requirements:

-   [golang-migrate](https://github.com/golang-migrate) to create new migrations.
    See [installation instructions](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation).

Tutorial:

1. Create a new migration:
    ```shell
    migrate create -dir ./pkg/server/db/migrations -ext sql example
    ```
2. Fill the `up` migration:
    ```sql
    -- pkg/server/db/migrations/20241223120923_example.up.sql
    CREATE TABLE `example` (
      `id` integer PRIMARY KEY AUTOINCREMENT,
      `data` text
    );
    ```
3. Fill the `down` migration:
    ```sql
    -- pkg/server/db/migrations/20241223120923_example.down.sql
    DROP TABLE `example`;
    ```

Conventions:

-   Use the schema `<table>-<action>` to name migrations.
    For example, `runs-create` to create the table runs or `runs-add-index-status`.
-   Prefer modifying a single table per migration file.

## Documentation

Documentation is written in [Markdown](https://www.markdownguide.org/) and rendered by [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/).

### Serve the documentation locally

1. (Optional) Install Python version via [pyenv](https://github.com/pyenv/pyenv):
    ```shell
    pyenv install
    ```
1. Create a Python virtualenv (needs to be done only once):
    ```shell
    python -m venv --prompt saturn-bot-docs ./venv
    ```
1. Activate the virtualenv:
    ```shell
    source venv/bin/activate
    ```
1. Install dependencies:
    ```shell
    pip install -r docs/requirements.txt
    ```
1. Start the server:
    ```shell
    mkdocs serve
    ```

Open [http://localhost:8000](http://localhost:8000) in a browser.

### Update dependencies

1. Activate the virtualenv:
    ```shell
    source venv/bin/activate
    ```
2. Install whatever dependencies are necessary:
    ```shell
    pip install ...
    ```
3. Update the requirements file:
    ```shell
    pip freeze -l > docs/requirements.txt
    ```
