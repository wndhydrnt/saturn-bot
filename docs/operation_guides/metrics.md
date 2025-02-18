---
title: Metrics
---

[saturn-bot run](../reference/commands/run.md) pushes metrics
to a [Prometheus Pushgateway](https://github.com/prometheus/pushgateway)
if [prometheusPushgatewayUrl](../reference/configuration.md#prometheuspushgatewayurl) is set.

The `job` label added to all metrics is `saturn-bot`.

The following metrics exist:

## `git_commands_duration_seconds_count`

Total number of commands executed by git.

Use this metric to understand how much load saturn-bot puts
on a repository host by, for example, cloning repositories or pushing commits.

## `git_commands_duration_seconds_sum`

Total duration it took for git to execute commands.

Use together with `git_commands_duration_seconds_count` to calculate the average duration
and understand the repository host is slow.

## `http_client_requests_total`

Total number of requests sent via HTTP clients.

This metric includes HTTP requests sent to GitHub or GitLab.

Useful to understand how much load saturn-bot puts on an API.

## `run_finish_time_seconds`

Last unix time when the run finished.

Use this metric together with [`run_start_time_seconds`](#run_start_time_seconds)
to understand how long the execution of `saturn-bot run` took.

## `run_start_time_seconds`

Last unix time when the run started.

Use this metric together with [`run_finish_time_seconds`](#run_finish_time_seconds)
to understand how long the execution of `saturn-bot run` took.

## `run_task_success`

Status of the last run of a task. 1 indicates success. 0 indicates failure.

Use this metric to alert that a task has failed.
