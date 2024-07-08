# Configuration

Configuration is stored in the YAML format.

```yaml title="Example"
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

| Name    | Value                                   |
| ------- | --------------------------------------- |
| Default | `saturn-bot <bot@saturn-bot.localhost>` |
| Env Var | `SATURN_BOT_GITAUTHOR`                  |
| Type    | `string`                                |

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
