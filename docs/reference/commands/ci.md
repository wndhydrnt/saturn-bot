# ci

```text
--8<-- "docs/reference/commands/ci.txt"
```

`ci` starts and initializes plugins as part of its execution.
If this behavior isn't desired, the flag `--skip-plugins=true` can be passed to the command.
It is also possible to make each plugin
[skip initialization during CI runs](../task/plugins/index.md#skip-initialization-during-ci-runs).
