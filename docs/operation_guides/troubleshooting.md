---
title: Troubleshooting
---

## Kubernetes reports high memory usage for `run` or `worker` commands

### Problem

The metric `container_memory_working_set_bytes`,
exported by [cadvisor](https://github.com/google/cadvisor),
shows that the process allocates more and more memory
and gets close to its configured memory limit.

### Likely cause

saturn-bot `run` and `worker` commands interact with a lot of files when processing tasks.

A non-exhaustive lists of actions that involve files:

-   Caching repository data on the local filesystem.
-   Cloning Git repositories or pulling in changes from Git repositories.
-   Creating or modifying files in Git repositories to apply the [actions](../reference/task/actions/index.md) of a task.

All these interactions with the filesystem make the [page cache](https://en.wikipedia.org/wiki/Page_cache)
of the process grow.
Specifically, the output of `cat /sys/fs/cgroup/memory.stat`, executed from within a container,
reports a high value for `active_list`.

Kubernetes considers `active_list` as part of the memory in use.
The section [active_file memory is not considered as available memory](https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction/#active-file-memory-is-not-considered-as-available-memory)
in the Kubernetes documentation explains this in more detail.

### Possible mitigations

Set memory limit and memory request of a container to the same value.
Some experimentation might be needed to arrive at that value.
A good starting point is to inspect the metric `process_resident_memory_bytes` exported by the `worker` command.
However, that metric doesn't capture memory usage by external processes like
[plugins](../reference/task/index.md#plugins) or [scripts](../reference/task/actions/script.md).
