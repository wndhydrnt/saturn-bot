# Installation

## Homebrew

Install from the brew tap:

```shell
brew install wndhydrnt/tap/saturn-bot
```

## Docker

Run the Docker container:

<!-- x-release-please-start-version -->

```shell
docker run --rm -it ghcr.io/wndhydrnt/saturn-bot:v0.12.0 version
```

The tag `<version>-full` contains runtimes for Java and Python to execute plugins:

```shell
docker run --rm -it ghcr.io/wndhydrnt/saturn-bot:v0.12.0-full version
```

<!-- x-release-please-end -->
