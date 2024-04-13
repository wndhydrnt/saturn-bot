FROM golang:1.22.1-alpine as base
WORKDIR /src
RUN apk add make
COPY go.mod go.sum .
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x
COPY . .

FROM base AS builder
ARG VERSION=dev
ARG VERSION_HASH
RUN --mount=type=cache,target=/go/pkg/mod/ \
    make build

FROM alpine:3.19.1
RUN apk add gcompat
COPY --from=builder /src/saturn-sync /bin/saturn-sync
ENTRYPOINT ["saturn-sync"]
