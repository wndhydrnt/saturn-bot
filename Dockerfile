# golang:1.23.6-bookworm
FROM golang@sha256:441f59f8a2104b99320e1f5aaf59a81baabbc36c81f4e792d5715ef09dd29355 AS base
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

# debian:bookworm-20250203-slim
FROM debian@sha256:40b107342c492725bc7aacbe93a49945445191ae364184a6d24fedb28172f6f7
ENV SATURN_BOT_DATADIR=/home/saturn-bot/data
RUN useradd --create-home --shell /usr/sbin/nologin --uid 1001 --system saturn-bot && \
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
