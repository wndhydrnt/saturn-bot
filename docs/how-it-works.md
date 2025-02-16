# How it works

On each execution of [saturn-bot run](./reference/commands/run.md),
saturn-bot takes the following steps:

1. Read all task files.
1. List all repositories in the host it has access to, GitHub or GitLab.
1. For each repository and each task, do:
    1. Check if all [filters](./reference/task/filters/index.md) match the repository.
    1. If all filters match, clone the repository to the local file system.
    1. Create a new branch locally.
    1. Apply all [actions](./reference/task/actions/index.md) defined in the task file to the local clone of the repository.
    1. If any of the actions has added, modified or deleted files in the local clone, create a commit.
    1. Push the commit and branch to the remote.
    1. Create a pull request or the pull request, if it exists already.
