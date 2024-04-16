# golang:1.22.2-bookworm
FROM golang@sha256:b03f3ba515751657c75475b20941fef47341fccb3341c3c0b64283ff15d3fb46 as base
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

# debian:bookworm-20240408-slim
FROM debian@sha256:3d5df92588469a4c503adbead0e4129ef3f88e223954011c2169073897547cac
ENV SATURN_SYNC_DATADIR=/home/saturn-sync/data
RUN useradd --create-home --shell /usr/sbin/nologin --uid 1001 saturn-sync && \
    mkdir /home/saturn-sync/data && \
    chown 1001:1001 /home/saturn-sync/data && \
    apt-get update && \
    apt-get install --no-install-recommends -y git=1:2.39.2-1.1 ca-certificates=20230311 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder --chown=1001:1001 /src/saturn-sync /bin/saturn-sync
USER saturn-sync
WORKDIR /home/saturn-sync
ENTRYPOINT ["saturn-sync"]
