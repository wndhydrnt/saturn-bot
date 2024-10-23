---
title: Metrics
---

[saturn-bot run](../commands/run.md) pushes metrics
to a [Prometheus Pushgateway](https://github.com/prometheus/pushgateway)
if [prometheusPushgatewayUrl](../configuration.md#prometheuspushgatewayurl) is set.

The `job` label added to all metrics is `saturn-bot`.

The following metrics exist:

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
