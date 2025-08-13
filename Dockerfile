FROM golang@1.24.6-bookworm AS base
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x
COPY . .

FROM base AS builder
ARG VERSION=dev
ARG VERSION_HASH
RUN --mount=type=cache,target=/go/pkg/mod/ \
    make build

# debian:bookworm-20250610-slim
FROM debian@sha256:d42b86d7e24d78a33edcf1ef4f65a20e34acb1e1abd53cabc3f7cdf769fc4082
ENV SATURN_BOT_DATADIR=/var/lib/saturn-bot
RUN groupadd --system --gid 1001 saturn-bot && \
    useradd --system --gid saturn-bot --no-create-home --home /nonexistent --comment "saturn-bot user" --shell /bin/false --uid 1001 saturn-bot && \
    mkdir ${SATURN_BOT_DATADIR} && \
    chown 1001:1001 ${SATURN_BOT_DATADIR} && \
    apt-get update && \
    apt-get install --no-install-recommends -y git=1:2.39.5-0+deb12u2 ca-certificates=20230311 curl=7.88.1-10+deb12u12 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder --chown=1001:1001 /src/saturn-bot /bin/saturn-bot
USER saturn-bot
WORKDIR ${SATURN_BOT_DATADIR}
ENTRYPOINT ["saturn-bot"]
