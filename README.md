# saturn-bot 🪐🤖

Create, modify or delete files across many repositories in parallel.

## Use cases

- Onboard repositories to CI workflows.
- Keep files in sync across repositories.
- Automate code rewrites.

## Features

- Create or delete files.
- Insert, replace or delete lines in files.
- Filter which repositories to modify.
- Automatic creation of pull requests.
- Support for GitHub and GitLab.
- Implement custom logic and complex changes through plugins in Go and Python.
- Automatically merge pull requests if all checks have passed and all approvals have been given.

## Quickstart

Requirements:

- saturn-bot installed.
- An access token for GitHub or GitLab.

Create the file `hello-world.yaml`:

```yaml title="hello-world.yaml"
name: "saturn-bot Hello World"
prTitle: "saturn-bot Hello World"
prBody: |
  saturn-bot Quickstart.

  This pull request creates the file `hello-world.txt`.

# Filters tell saturn-bot which repositories to modify.
filters:
  repositoryName:
    # Replace the name of the repository with your repository.
    - names: ["github.com/wndhydrnt/saturn-bot-example"]

# Actions tell saturn-bot how to modify each repository.
actions:
  fileCreate:
    - content: "Hello World"
      path: "hello-world.txt"
```

Run saturn-bot:

```shell
SATURN_BOT_GITHUB_TOKEN=<token> saturn-bot run --task hello-world.yaml
```
