---
title: Home
---

# saturn-bot Documentation

Create, modify or delete files in many repositories in parallel.

## Use cases

- Onboard repositories to CI workflows and keep those files in sync.
- Perform automatic rewrite of code scattered in many repositories and create pull requests.

## Features

- Create or delete files.
- Insert, replace or delete lines in files.
- [Filter](task/filters/index.md) which repositories to modify.
- Support for GitHub and GitLab.
- Write plugins in Go or Python to implement custom logic and complex changes.
- Automatically merge pull requests if all checks have passed and all approvals have been given.

## Quickstart

Requirements:

- saturn-bot [installed](installation.md).
- An access token.

Create the file `hello-world.yaml`:

```yaml title="hello-world.yaml"
name: "saturn-bot Hello World"
prTitle: "saturn-bot Hello World"
prBody: |
  saturn-bot Quickstart.

  This pull request creates the file `hello-world.txt`.

# Filters tell saturn-bot which repositories to modify.
filters:
  - filter: repository
    params:
      host: github.com
      owner: wndhydrnt # Replace with your owner
      name: saturn-bot-example # Replace with your repository

# Actions tell saturn-bot how to modify each repository.
actions:
  - action: fileCreate
    params:
      content: "Hello World"
      path: "hello-world.txt"
```

Run saturn-bot:

```shell
SATURN_BOT_GITHUB_TOKEN=<token> saturn-bot run --task hello-world.yaml
```
