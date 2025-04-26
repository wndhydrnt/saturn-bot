# Configuration

Configuration is stored in the YAML format.

```yaml title="Example"
# yaml-language-server: $schema=https://saturn-bot.readthedocs.io/en/latest/schemas/config.schema.json
dataDir: /tmp/saturn-bot
logFormat: console
githubToken: xxxxx
```

## dataDir

[json-path:../../pkg/config/config.schema.json:$.properties.dataDir.description]

| Name    | Value                |
| ------- | -------------------- |
| Default | -                    |
| Env Var | `SATURN_BOT_DATADIR` |
| Type    | `string`             |

## dryRun

[json-path:../../pkg/config/config.schema.json:$.properties.dryRun.description]

| Name    | Value               |
| ------- | ------------------- |
| Default | `false`             |
| Env Var | `SATURN_BOT_DRYRUN` |
| Type    | `bool`              |

## logFormat

[json-path:../../pkg/config/config.schema.json:$.properties.logFormat.description]

| Name    | Value                     |
| ------- | ------------------------- |
| Default | `auto`                    |
| Env Var | `SATURN_BOT_LOGFORMAT`    |
| Type    | `string`                  |
| Values  | `auto`, `console`, `json` |

## logLevel

[json-path:../../pkg/config/config.schema.json:$.properties.logLevel.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `info`                           |
| Env Var | `SATURN_BOT_LOGLEVEL`            |
| Type    | `string`                         |
| Values  | `debug`, `error`, `info`, `warn` |

## gitAuthor

[json-path:../../pkg/config/config.schema.json:$.properties.gitAuthor.description]

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

[json-path:../../pkg/config/config.schema.json:$.properties.gitCloneOptions.description]

| Name    | Value                        |
| ------- | ---------------------------- |
| Default | `["--filter", "blob:none"]`  |
| Env Var | `SATURN_BOT_GITCLONEOPTIONS` |
| Type    | `[string]`                   |

## gitCommitMessage

[json-path:../../pkg/config/config.schema.json:$.properties.gitCommitMessage.description]

| Name    | Value                         |
| ------- | ----------------------------- |
| Default | `changes by saturn-bot`       |
| Env Var | `SATURN_BOT_GITCOMMITMESSAGE` |
| Type    | `string`                      |

## gitLogLevel

[json-path:../../pkg/config/config.schema.json:$.properties.gitLogLevel.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `changes by saturn-bot`          |
| Env Var | `SATURN_BOT_GITCOMMITMESSAGE`    |
| Type    | `string`                         |
| Values  | `debug`, `error`, `info`, `warn` |

## gitPath

[json-path:../../pkg/config/config.schema.json:$.properties.gitPath.description]

| Name    | Value                |
| ------- | -------------------- |
| Default | `git`                |
| Env Var | `SATURN_BOT_GITPATH` |
| Type    | `string`             |

## gitUrl

[json-path:../../pkg/config/config.schema.json:$.properties.gitUrl.description]

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

[json-path:../../pkg/config/config.schema.json:$.properties.githubAddress.description]

| Name    | Value                      |
| ------- | -------------------------- |
| Default | -                          |
| Env Var | `SATURN_BOT_GITHUBADDRESS` |
| Type    | `string`                   |

## githubCacheDisabled

[json-path:../../pkg/config/config.schema.json:$.properties.githubCacheDisabled.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `false`                          |
| Env Var | `SATURN_BOT_GITHUBCACHEDISABLED` |
| Type    | `bool`                           |

## githubToken

[json-path:../../pkg/config/config.schema.json:$.properties.githubToken.description]

| Name    | Value                    |
| ------- | ------------------------ |
| Default | -                        |
| Env Var | `SATURN_BOT_GITHUBTOKEN` |
| Type    | `string`                 |

## gitlabAddress

[json-path:../../pkg/config/config.schema.json:$.properties.gitlabAddress.description]

| Name    | Value                      |
| ------- | -------------------------- |
| Default | -                          |
| Env Var | `SATURN_BOT_GITLABADDRESS` |
| Type    | `string`                   |

## gitlabToken

[json-path:../../pkg/config/config.schema.json:$.properties.gitlabToken.description]

| Name    | Value                    |
| ------- | ------------------------ |
| Default | -                        |
| Env Var | `SATURN_BOT_GITLABTOKEN` |
| Type    | `string`                 |

## goAutoMemLimitRatio

[json-path:../../pkg/config/config.schema.json:$.properties.goAutoMemLimitRatio.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `0.0`                            |
| Env Var | `SATURN_BOT_GOAUTOMEMLIMITRATIO` |
| Type    | `number`                         |

```yaml title="Example"
# Set the soft limit of the Go runtime to 90% of the current container or system memory
goAutoMemLimitRatio: 0.9
```

## goProfiling

[json-path:../../pkg/config/config.schema.json:$.properties.goProfiling.description]

| Name    | Value                    |
| ------- | ------------------------ |
| Default | `false`                  |
| Env Var | `SATURN_BOT_GOPROFILING` |
| Type    | `bool`                   |

## javaPath

[json-path:../../pkg/config/config.schema.json:$.properties.javaPath.description]

| Name    | Value                 |
| ------- | --------------------- |
| Default | `java`                |
| Env Var | `SATURN_BOT_JAVAPATH` |
| Type    | `string`              |

## pluginLogLevel

[json-path:../../pkg/config/config.schema.json:$.properties.pluginLogLevel.description]

| Name    | Value                            |
| ------- | -------------------------------- |
| Default | `debug`                          |
| Env Var | `SATURN_BOT_PLUGINLOGLEVEL`      |
| Type    | `string`                         |
| Values  | `debug`, `error`, `info`, `warn` |

## prometheusPushgatewayUrl

[json-path:../../pkg/config/config.schema.json:$.properties.prometheusPushgatewayUrl.description]

| Name    | Value                                 |
| ------- | ------------------------------------- |
| Default | -                                     |
| Env Var | `SATURN_BOT_PROMETHEUSPUSHGATEWAYURL` |
| Type    | `string`                              |

## pythonPath

[json-path:../../pkg/config/config.schema.json:$.properties.pythonPath.description]

| Name    | Value                   |
| ------- | ----------------------- |
| Default | `python`                |
| Env Var | `SATURN_BOT_PYTHONPATH` |
| Type    | `string`                |

## repositoryCacheTtl

[json-path:../../pkg/config/config.schema.json:$.properties.repositoryCacheTtl.description]

| Name    | Value                           |
| ------- | ------------------------------- |
| Default | `6h`                            |
| Env Var | `SATURN_BOT_REPOSITORYCACHETTL` |
| Type    | `string`                        |

## serverAccessLog

[json-path:../../pkg/config/config.schema.json:$.properties.serverAccessLog.description]

| Name    | Value                         |
| ------- | ----------------------------- |
| Default | `false`                       |
| Env Var | `SATURN_BOT_SERVERACCESSLOGS` |
| Type    | `bool`                        |

## serverAddr

[json-path:../../pkg/config/config.schema.json:$.properties.serverAddr.description]

| Name    | Value                   |
| ------- | ----------------------- |
| Default | `:3035`                 |
| Env Var | `SATURN_BOT_SERVERADDR` |
| Type    | `string`                |

## serverApiKey

[json-path:../../pkg/config/config.schema.json:$.properties.serverApiKey.description]

| Name    | Value                     |
| ------- | ------------------------- |
| Default | -                         |
| Env Var | `SATURN_BOT_SERVERAPIKEY` |
| Type    | `string`                  |

## serverBaseUrl

[json-path:../../pkg/config/config.schema.json:$.properties.serverBaseUrl.description]

| Name    | Value                      |
| ------- | -------------------------- |
| Default | `http://localhost:3035`    |
| Env Var | `SATURN_BOT_SERVERBASEURL` |
| Type    | `string`                   |

## serverCompress

[json-path:../../pkg/config/config.schema.json:$.properties.serverCompress.description]

| Name    | Value                       |
| ------- | --------------------------- |
| Default | `true`                      |
| Env Var | `SATURN_BOT_SERVERCOMPRESS` |
| Type    | `bool`                      |

## serverDatabaseLog

[json-path:../../pkg/config/config.schema.json:$.properties.serverDatabaseLog.description]

| Name    | Value                           |
| ------- | ------------------------------- |
| Default | `false`                         |
| Env Var | `SATURN_BOT_SERVERDATABASEPATH` |
| Type    | `bool`                          |

## serverDatabasePath

[json-path:../../pkg/config/config.schema.json:$.properties.serverDatabasePath.description]

| Name    | Value                           |
| ------- | ------------------------------- |
| Default | -                               |
| Env Var | `SATURN_BOT_SERVERDATABASEPATH` |
| Type    | `string`                        |

## serverWebhookSecretGithub

[json-path:../../pkg/config/config.schema.json:$.properties.serverWebhookSecretGithub.description]

| Name    | Value                                  |
| ------- | -------------------------------------- |
| Default | -                                      |
| Env Var | `SATURN_BOT_SERVERWEBHOOKSECRETGITHUB` |
| Type    | `string`                               |

## serverWebhookSecretGitlab

[json-path:../../pkg/config/config.schema.json:$.properties.serverWebhookSecretGitlab.description]

| Name    | Value                                  |
| ------- | -------------------------------------- |
| Default | -                                      |
| Env Var | `SATURN_BOT_SERVERWEBHOOKSECRETGITLAB` |
| Type    | `string`                               |

## serverServeUi

[json-path:../../pkg/config/config.schema.json:$.properties.serverServeUi.description]

| Name    | Value                      |
| ------- | -------------------------- |
| Default | `true`                     |
| Env Var | `SATURN_BOT_SERVERSERVEUI` |
| Type    | `bool`                     |

## workerLoopInterval

[json-path:../../pkg/config/config.schema.json:$.properties.workerLoopInterval.description]

| Name    | Value                           |
| ------- | ------------------------------- |
| Default | `10s`                           |
| Env Var | `SATURN_BOT_WORKERLOOPINTERVAL` |
| Type    | `string`                        |

## workerParallelExecutions

[json-path:../../pkg/config/config.schema.json:$.properties.workerParallelExecutions.description]

| Name    | Value                                 |
| ------- | ------------------------------------- |
| Default | `1`                                   |
| Env Var | `SATURN_BOT_WORKERPARALLELEXECUTIONS` |
| Type    | `integer`                             |

## workerServerAPIBaseURL

[json-path:../../pkg/config/config.schema.json:$.properties.workerServerAPIBaseURL.description]

| Name    | Value                               |
| ------- | ----------------------------------- |
| Default | `http://localhost:3035`             |
| Env Var | `SATURN_BOT_WORKERSERVERAPIBASEURL` |
| Type    | `string`                            |
