# Configuration

Configuration is stored in the YAML format.

```yaml title="Example"
# yaml-language-server: $schema=https://saturn-bot.readthedocs.io/en/latest/schemas/config.schema.json
dataDir: /tmp/saturn-bot
logFormat: console
githubToken: xxxxx
```

## dataDir

[json-path:../pkg/config/config.schema.json:$.properties.dataDir.description]

| Name    | Value                |
| ------- | -------------------- |
| Default | -                    |
| Env Var | `SATURN_BOT_DATADIR` |
| Type    | `string`             |

## dryRun

[json-path:../pkg/config/config.schema.json:$.properties.dryRun.description]

| Name    | Value               |
| ------- | ------------------- |
| Default | `false`             |
| Env Var | `SATURN_BOT_DRYRUN` |
| Type    | `bool`              |

## logFormat

[json-path:../pkg/config/config.schema.json:$.properties.logFormat.description]

| Name    | Value                     |
| ------- | ------------------------- |
| Default | `auto`                    |
| Env Var | `SATURN_BOT_LOGFORMAT`    |
| Type    | `string`                  |
| Values  | `auto`, `console`, `json` |

## logLevel

[json-path:../pkg/config/config.schema.json:$.properties.logLevel.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `info`                           |
| Env Var | `SATURN_BOT_LOGLEVEL`            |
| Type    | `string`                         |
| Values  | `debug`, `error`, `info`, `warn` |

## gitAuthor

[json-path:../pkg/config/config.schema.json:$.properties.gitAuthor.description]

If not set, saturn-bot tries to discover the author via the host of the repository.

| Name    | Value                  |
| ------- | ---------------------- |
| Default | -                      |
| Env Var | `SATURN_BOT_GITAUTHOR` |
| Type    | `string`               |

```yaml title="Set a custom author"
gitAuthor: "Saturn Bot <saturn-bot@example.local>"
```

## gitCloneOptions

[json-path:../pkg/config/config.schema.json:$.properties.gitCloneOptions.description]

| Name    | Value                        |
| ------- | ---------------------------- |
| Default | `["--filter", "blob:none"]`  |
| Env Var | `SATURN_BOT_GITCLONEOPTIONS` |
| Type    | `[string]`                   |

## gitCommitMessage

[json-path:../pkg/config/config.schema.json:$.properties.gitCommitMessage.description]

| Name    | Value                         |
| ------- | ----------------------------- |
| Default | `changes by saturn-bot`       |
| Env Var | `SATURN_BOT_GITCOMMITMESSAGE` |
| Type    | `string`                      |

## gitLogLevel

[json-path:../pkg/config/config.schema.json:$.properties.gitLogLevel.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `changes by saturn-bot`          |
| Env Var | `SATURN_BOT_GITCOMMITMESSAGE`    |
| Type    | `string`                         |
| Values  | `debug`, `error`, `info`, `warn` |

## gitPath

[json-path:../pkg/config/config.schema.json:$.properties.gitPath.description]

| Name    | Value                |
| ------- | -------------------- |
| Default | `git`                |
| Env Var | `SATURN_BOT_GITPATH` |
| Type    | `string`             |

## gitUrl

[json-path:../pkg/config/config.schema.json:$.properties.gitUrl.description]

Set to `ssh` to clone repositories via SSH.

| Name    | Value               |
| ------- | ------------------- |
| Default | `https`             |
| Env Var | `SATURN_BOT_GITURL` |
| Type    | `string`            |
| Values  | `https`, `ssh`      |

!!! warning

    If set to `ssh`, git and ssh need to be configured.
    Follow instructions for your platform:

    - [GitHub](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)
    - [GitLab](https://docs.gitlab.com/ee/ci/ssh_keys/)

## githubAddress

[json-path:../pkg/config/config.schema.json:$.properties.githubAddress.description]

| Name    | Value                      |
| ------- | -------------------------- |
| Default | -                          |
| Env Var | `SATURN_BOT_GITHUBADDRESS` |
| Type    | `string`                   |

## githubCacheDisabled

[json-path:../pkg/config/config.schema.json:$.properties.githubCacheDisabled.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `false`                          |
| Env Var | `SATURN_BOT_GITHUBCACHEDISABLED` |
| Type    | `bool`                           |

## githubToken

[json-path:../pkg/config/config.schema.json:$.properties.githubToken.description]

| Name    | Value                    |
| ------- | ------------------------ |
| Default | -                        |
| Env Var | `SATURN_BOT_GITHUBTOKEN` |
| Type    | `string`                 |

## gitlabAddress

[json-path:../pkg/config/config.schema.json:$.properties.gitlabAddress.description]

| Name    | Value                      |
| ------- | -------------------------- |
| Default | -                          |
| Env Var | `SATURN_BOT_GITLABADDRESS` |
| Type    | `string`                   |

## gitlabToken

[json-path:../pkg/config/config.schema.json:$.properties.gitlabToken.description]

| Name    | Value                    |
| ------- | ------------------------ |
| Default | -                        |
| Env Var | `SATURN_BOT_GITLABTOKEN` |
| Type    | `string`                 |

## javaPath

[json-path:../pkg/config/config.schema.json:$.properties.javaPath.description]

| Name    | Value                 |
| ------- | --------------------- |
| Default | `java`                |
| Env Var | `SATURN_BOT_JAVAPATH` |
| Type    | `string`              |

## pluginLogLevel

[json-path:../pkg/config/config.schema.json:$.properties.pluginLogLevel.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `debug`                          |
| Env Var | `SATURN_BOT_PLUGINLOGLEVEL`      |
| Type    | `string`                         |
| Values  | `debug`, `error`, `info`, `warn` |

## prometheusPushgatewayUrl

[json-path:../pkg/config/config.schema.json:$.properties.prometheusPushgatewayUrl.description]

| Name    | Value                                 |
| ------- | ------------------------------------- |
| Default | -                                     |
| Env Var | `SATURN_BOT_PROMETHEUSPUSHGATEWAYURL` |
| Type    | `string`                              |

## pythonPath

[json-path:../pkg/config/config.schema.json:$.properties.pythonPath.description]

| Name    | Value                   |
| ------- | ----------------------- |
| Default | `python`                |
| Env Var | `SATURN_BOT_PYTHONPATH` |
| Type    | `string`                |

## repositoryCacheTtl

[json-path:../pkg/config/config.schema.json:$.properties.repositoryCacheTtl.description]

| Name    | Value                           |
| ------- | ------------------------------- |
| Default | `6h`                            |
| Env Var | `SATURN_BOT_REPOSITORYCACHETTL` |
| Type    | `string`                        |
