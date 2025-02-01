VERSION?=v0.0.0-dev
VERSION_HASH?=$(shell git rev-parse HEAD)
VERSION_DATETIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_FLAGS=-ldflags="-X 'github.com/wndhydrnt/saturn-bot/pkg/version.Version=$(VERSION)' -X 'github.com/wndhydrnt/saturn-bot/pkg/version.Hash=$(VERSION_HASH)' -X 'github.com/wndhydrnt/saturn-bot/pkg/version.DateTime=$(VERSION_DATETIME)'"
GO_JSONSCHEMA_VERSION=v0.16.0
OS=$(shell uname -s)
ARCH=$(shell uname -m)

build:
	go build $(BUILD_FLAGS) -o saturn-bot

build_darwin_x86_64:
	GOARCH=amd64 GOOS=darwin go build $(BUILD_FLAGS) -o saturn-bot

package_darwin_x86_64: build_darwin_x86_64
	tar -a -cf saturn-bot-$(VERSION).Darwin-x86_64.tar.gz saturn-bot LICENSE

build_darwin_arm64:
	GOARCH=arm64 GOOS=darwin go build $(BUILD_FLAGS) -o saturn-bot

package_darwin_arm64: build_darwin_arm64
	tar -a -cf saturn-bot-$(VERSION).Darwin-arm64.tar.gz saturn-bot LICENSE

build_linux_aarch64:
	GOARCH=arm64 GOOS=linux go build $(BUILD_FLAGS) -o saturn-bot

package_linux_aarch64: build_linux_aarch64
	tar -a -cf saturn-bot-$(VERSION).Linux-aarch64.tar.gz saturn-bot LICENSE

build_linux_x86_64:
	GOARCH=amd64 GOOS=linux go build $(BUILD_FLAGS) -o saturn-bot

package_linux_x86_64: build_linux_x86_64
	tar -a -cf saturn-bot-$(VERSION).Linux-x86_64.tar.gz saturn-bot LICENSE

build_all: build_darwin_x86_64 build_darwin_arm64 build_linux_aarch64 build_linux_x86_64

package_all: package_darwin_x86_64 package_darwin_arm64 package_linux_aarch64 package_linux_x86_64 checksums

checksums:
	sha256sum saturn-bot-$(VERSION).*.tar.gz > sha256sums.txt

test_cover:
	go test -coverpkg=./... -coverprofile cover.out.tmp ./...
	grep -v -E ".*\/pkg\/server\/api\/openapi\/.*|.*\/pkg\/worker\/.*|.*\/test\/.*" cover.out.tmp > cover.out
	go tool cover -html cover.out -o cover.html
	rm cover.out.tmp cover.out

docker_build:
	docker build -t ghcr.io/wndhydrnt/saturn-bot:${VERSION} .

docker_build_full: docker_build
	docker build --build-arg="BASE=${VERSION}" -t ghcr.io/wndhydrnt/saturn-bot:${VERSION}-full -f full.Dockerfile .
