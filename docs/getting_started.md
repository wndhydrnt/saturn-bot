# Getting started

This tutorial will get you up and running with using saturn-bot to modify a repository that is hosted at GitHub or GitLab.

By the end, you'll understand how to write a task and use the `run` and `try` commands.

## Prerequisites

### Install saturn-bot

Before continuing, [install saturn-bot on your local machine](installation.md).

### Create an access token

saturn-bot queries the API of the host that hosts the repository.
It needs an access token to authenticate at the API.

Follow [Create an access token](./operation_guides/generate_access_token.md) for your host.

Store the access token in the file `config.yaml`. Replace `<token>` with the actual token.

=== "GitHub"

    ```yaml title="config.yaml"
    githubToken: <token>
    ```

=== "GitLab"

    ```yaml title="config.yaml"
    gitlabToken: <token>
    ```

### Create a repository

Create a repository for saturn-bot to modify.
Skip this step if you already have a repository you can use for testing purposes.

## Creating the task file

The task file contains all the information saturn-bot needs to discover which repositories to modify,
how to modify them and how to create a pull request to submit the changes.

Without the task file, saturn-bot cannot do anything.

The following task file creates lets saturn-bot create the file `hello-world.txt`
with the content `Hello World` in the root of a repository. Save it as `hello-world.yaml`.

```yaml title="hello-world.yaml"
# yaml-language-server: $schema=https://saturn-bot.readthedocs.io/en/latest/schemas/task.schema.json
name: "saturn-bot Hello World"
prTitle: "saturn-bot Hello World"
prBody: |
  saturn-bot "Getting started" tutorial.

  This pull request creates the file `hello-world.txt`.

# Filters tell saturn-bot which repositories to modify.
filters:
  - filter: repository
    params:
      host: github.com # Replace with your host
      owner: wndhydrnt # Replace with your owner
      name: saturn-bot-example # Replace with your repository

# Actions tell saturn-bot how to modify each repository.
actions:
  - action: fileCreate
    params:
      content: "Hello World"
      path: "hello-world.txt"
```

## Trying out the task

Let's check if the task modifies the repository.
saturn-bot provides the `try` command to check locally what a task would do without pushing the changes to the repository.
It allows to rapidly iterate when developing a task and inspect the changes and avoid,
for example, repeatedly triggered CI/CD pipeline.

```shell
saturn-bot try \
  --config config.yaml \
  --repository github.com/wndhydrnt/saturn-bot-example \
  ./hello-world.yaml
```

saturn-bot lets you know that it has modified files:

```text
✅ Filter repository(host=^github.com$,owner=^wndhydrnt$,name=^(saturn-bot-example)$) of task saturn-bot Hello World matches
🏗️ Cloning repository
😍 Actions modified files - view checkout in ~/.saturn-bot/data/git/github.com/wndhydrnt/saturn-bot-example
```

Go to `~/.saturn-bot/data/git/github.com/wndhydrnt/saturn-bot-example` and check that the file `hello-world.txt` exists:

```shell
cd ~/.saturn-bot/data/git/github.com/wndhydrnt/saturn-bot-example
cat hello-world.txt
```

Output:

```text
Hello World
```

## Executing the task

Now that you have checked what saturn-bot would change, it is time to
execute saturn-bot and create the pull request:

```shell
saturn-bot run \
  --config config.yaml \
  --repository github.com/wndhydrnt/saturn-bot-example \
  ./hello-world.yaml
```

Check that a new pull request with the title "saturn-bot Hello World" has been created.

## What's next

-   Read the [task reference](./reference/task/index.md) to learn about all options, actions and filters of a task.
