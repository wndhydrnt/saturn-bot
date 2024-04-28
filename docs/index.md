---
title: Home
---

# saturn-sync Documentation

Create, modify or delete files in many repositories in parallel.

## Features

- [Create]() or [delete]() files.
- [Insert](), [replace]() or [delete]() lines in files.
- [Filter]() which repositories to modify.
- Support for [GitHub]() and [GitLab]().
- Write plugins in [Go]() or [Python]() to implement custom logic and complex changes.
- [Automatically merge pull requests]() if all checks have passed and all approvals have been given.

## Quickstart

Requirements:

- saturn-sync [installed]().
- An access token.

Create the file `hello-world.yaml`:

```yaml title="hello-world.yaml"
name: "saturn-sync Hello World"
prTitle: "saturn-sync Hello World"
prBody: |
  This pull request creates the file `hello-world.txt`.

filters:
  repositoryName:
    # Replace the name of the repository with your repository.
    - names: ["github.com/wndhydrnt/saturn-sync-example"]

actions:
  fileCreate:
    - content: "Hello World"
      path: "hello-world.txt"
```

Run saturn-sync:

```shell
SATURN_SYNC_GITHUB_TOKEN=<token> saturn-sync run --task hello-world.yaml
```
