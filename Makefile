VERSION?=v0.0.0-dev
VERSION_HASH?=$(shell git rev-parse HEAD)
VERSION_DATETIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_FLAGS=-ldflags="-X 'github.com/wndhydrnt/saturn-bot/pkg/version.Version=$(VERSION)' -X 'github.com/wndhydrnt/saturn-bot/pkg/version.Hash=$(VERSION_HASH)' -X 'github.com/wndhydrnt/saturn-bot/pkg/version.DateTime=$(VERSION_DATETIME)'"
GO_JSONSCHEMA_VERSION=v0.16.0
OS=$(shell uname -s)
ARCH=$(shell uname -m)

build:
	go build $(BUILD_FLAGS) -o saturn-bot

build_darwin_amd64:
	GOARCH=amd64 GOOS=darwin go build $(BUILD_FLAGS) -o saturn-bot-$(VERSION).darwin-amd64

build_darwin_arm64:
	GOARCH=arm64 GOOS=darwin go build $(BUILD_FLAGS) -o saturn-bot-$(VERSION).darwin-arm64

build_linux_arm64:
	GOARCH=arm64 GOOS=linux go build $(BUILD_FLAGS) -o saturn-bot-$(VERSION).linux-arm64
	cp saturn-bot-$(VERSION).linux-arm64 saturn-bot-$(VERSION).linux-aarch64

build_linux_armv7:
	GOARCH=arm GOOS=linux go build $(BUILD_FLAGS) -o saturn-bot-$(VERSION).linux-armv7

build_linux_amd64:
	GOARCH=amd64 GOOS=linux go build $(BUILD_FLAGS) -o saturn-bot-$(VERSION).linux-amd64

build_all: build_darwin_amd64 build_darwin_arm64 build_linux_arm64 build_linux_armv7 build_linux_amd64 checksums

checksums:
	sha256sum saturn-bot-$(VERSION).* > sha256sums.txt

generate_go:
ifeq (, $(shell which mockgen))
	go install go.uber.org/mock/mockgen@latest
endif
ifeq (, $(shell which stringer))
	go install golang.org/x/tools/cmd/stringer@latest
endif
ifeq (, $(shell which go-jsonschema))
	go install github.com/atombender/go-jsonschema@v0.17.0
endif
	go generate ./...

generate_completion: build
	./saturn-bot completion bash > ./completion/saturn-bot.bash
	./saturn-bot completion zsh > ./completion/saturn-bot.zsh

mdox: build
ifeq (, $(shell which mdox))
	go install github.com/bwplotka/mdox@latest
endif
	mdox fmt --soft-wraps ./docs/commands/*.md

test_cover:
	go test -coverpkg=./... -coverprofile cover.out.tmp ./...
	grep -v -E ".*/pkg\/server\/api\/openapi\/.*|.*\/pkg\/worker\/.*" cover.out.tmp > cover.out
	go tool cover -html cover.out -o cover.html
	rm cover.out.tmp cover.out

docker_build:
	docker build -t ghcr.io/wndhydrnt/saturn-bot:${VERSION} .

docker_build_full: docker_build
	docker build --build-arg="BASE=${VERSION}" -t ghcr.io/wndhydrnt/saturn-bot:${VERSION}-full -f full.Dockerfile .
