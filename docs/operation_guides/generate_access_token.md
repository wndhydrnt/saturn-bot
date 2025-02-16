---
title: Generate an access token
---

saturn-bot needs an access token to authenticate at the API of the host and push changes to repositories.

The following guides describe how to create an access token for the supported hosts.

## GitHub

1.  Create a [Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token).
2.  Create the token with scopes `repo` and `user:email`.
3.  Configure [`githubToken`](../reference/configuration.md#githubtoken).

## GitLab

1.  Depending on your use case, follow [Create a personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token)
    or [Create a group access token](https://docs.gitlab.com/ee/user/group/settings/group_access_tokens.html#create-a-group-access-token).
2.  Create the token with scopes `api` and `write_repository`.
3.  Configure [`gitlabToken`](../reference/configuration.md#gitlabtoken).
