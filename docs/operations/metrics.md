---
title: Metrics
---

[saturn-bot run](../commands/run.md) can optionally push metrics
to a [Prometheus Pushgateway](https://github.com/prometheus/pushgateway).

The following metrics exist:

## `saturn_bot_http_client_requests_total`

## `saturn_bot_run_finish_time_seconds`

## `saturn_bot_run_start_time_seconds`

## `saturn_bot_run_task_success`

Status of the last run of a task. 1 indicates a successful run. 0 indicates a failed run.

Use this metric to alert if a task fails.
