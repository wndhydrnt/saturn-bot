# Trigger a task when a webhook is received

`saturn-bot server` can trigger tasks when it receives webhooks from GitHub or GitLab.

This page describes the setup.

## Prerequisites

### Generate a webhook secret

Use Python to generate a random secret:

```shell
python -c 'import secrets; print(secrets.token_hex(32))'
```

Record the secret.

## GitHub

### Configure the webhook secret

Update the [configuration](../reference/configuration.md#serverwebhooksecretgithub) of saturn-bot:

```yaml
serverWebhookSecretGithub: <secret>
```

Restart the server to ensure that the change takes effect.

### Create the webhook

1.  Follow [Creating a repository webhook](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks#creating-a-repository-webhook) to create a webhook for a single repository.
    Follow [Creating an organization webhook](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks#creating-an-organization-webhook) to create a webhook for an organization.
2.  Set **Payload URL** to `https://$URL/webhooks/github`.
3.  Set **Content type** to **application/json**.
4.  Set **Secret** to the recorded secret.
5.  Select the events that trigger the webhook.
    The `push` event is the most common event.

### Trigger a task when a webhook is received

```yaml
name: "GitHub Webhook Trigger Example"
# ... other settings
trigger:
  webhook:
    github:
      - event: "push" # (1)
        filters: # (2)
          - '.repository.full_name == "wndhydrnt/saturn-bot"'
          - '.ref | startswith("refs/tags/")'
```

1.  This value must match one of the events selected during [Create the webhook](#create-the-webhook).
    [Webhook events and payloads](https://docs.github.com/en/webhooks/webhook-events-and-payloads) documents all available webhook events.
2.  `filters` allow to further select the webhook(s) that trigger a task by inspecting the body of the webhook.
    Each filter is a [`jq`](https://jqlang.org) expression.
    [Webhook events and payloads](https://docs.github.com/en/webhooks/webhook-events-and-payloads) documents payloads of webhook events.

## GitLab

### Configure the webhook secret

Update the [configuration](../reference/configuration.md#serverwebhooksecretgitlab) of saturn-bot:

```yaml
serverWebhookSecretGitlab: <secret>
```

Restart the server to ensure that the change takes effect.

### Create the webhook

1.  Follow [Create a webhook](https://docs.gitlab.com/user/project/integrations/webhooks/#create-a-webhook) to create a webhook for a project or group.
2.  Set **URL** to `https://$URL/webhooks/gitlab`.
3.  Set **Secret token** to the recorded secret.
4.  Select the events that trigger the webhook.
    The `Push event` is the most common event.

### Trigger a task when a webhook is received

```yaml
name: "GitLab Webhook Trigger Example"
# ... other settings
trigger:
  webhook:
    gitlab:
      - event: "Push Hook" # (1)
        filters: # (2)
          - '.project.path_with_namespace == "mike/diaspora"'
          - '.ref == "refs/heads/main"'
```

1.  This value must match one of the events selected during [Create the webhook](#create-the-webhook_1).
    [Webhook events](https://docs.gitlab.com/user/project/integrations/webhook_events/) documents the events.
2.  `filters` allow to further select the webhook(s) that trigger a task by inspecting the body of the webhook.
    Each filter is a [`jq`](https://jqlang.org) expression.
    [Webhook events](https://docs.gitlab.com/user/project/integrations/webhook_events/) documents payloads of webhook events.
