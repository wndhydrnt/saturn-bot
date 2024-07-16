# Plugins

saturn-bot allows users to implement their own filter and processing logic. It is also possible to react to actions of saturn-bot, like creating, merging or closing a pull request.

Plugins can be written in [Go](go.md), [Kotlin](kotlin.md) and [Python](python.md).

## Lifecycle of a plugin

saturn-bot starts each plugin of a task in a new sub-process. It communicates with plugins over gRPC.

```mermaid
sequenceDiagram
    saturn-bot->>Plugin: start
    saturn-bot->>Plugin: send configuration
    saturn-bot->>saturn-bot: list repositories
    loop for each repository
        saturn-bot->>Plugin: call filter()
        Plugin-->>saturn-bot: return result
        saturn-bot->>Plugin: call apply()
        saturn-bot->>Plugin: call onPrCreated()
        saturn-bot->>Plugin: call onPrMerged()
        saturn-bot->>Plugin: call onPrClosed()
    end
    saturn-bot->>Plugin: stop
```
