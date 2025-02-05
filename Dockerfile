# golang:1.23.6-bookworm
FROM golang@sha256:441f59f8a2104b99320e1f5aaf59a81baabbc36c81f4e792d5715ef09dd29355 AS base
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

# debian:bookworm-20241111-slim
FROM debian@sha256:ca3372ce30b03a591ec573ea975ad8b0ecaf0eb17a354416741f8001bbcae33d
ENV SATURN_BOT_DATADIR=/home/saturn-bot/data
RUN useradd --create-home --shell /usr/sbin/nologin --uid 1001 saturn-bot && \
    mkdir /home/saturn-bot/data && \
    chown 1001:1001 /home/saturn-bot/data && \
    apt-get update && \
    apt-get install --no-install-recommends -y git=1:2.39.5-0+deb12u1 ca-certificates=20230311 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder --chown=1001:1001 /src/saturn-bot /bin/saturn-bot
USER saturn-bot
WORKDIR /home/saturn-bot
ENTRYPOINT ["saturn-bot"]
