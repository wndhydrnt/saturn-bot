VERSION?=dev
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

build_all: build_darwin_amd64 build_darwin_arm64 build_linux_arm64 build_linux_armv7 build_linux_amd64

generate_grpc:
	buf generate

generate_json_schema_config: bin/go-jsonschema/go-jsonschema-${GO_JSONSCHEMA_VERSION}
	bin/go-jsonschema/go-jsonschema-${GO_JSONSCHEMA_VERSION} --extra-imports -p config -t ./pkg/config/config.schema.json --output ./pkg/config/schema.go

generate_json_schema_task: bin/go-jsonschema/go-jsonschema-${GO_JSONSCHEMA_VERSION}
	bin/go-jsonschema/go-jsonschema-${GO_JSONSCHEMA_VERSION} --extra-imports -p schema -t ./pkg/task/schema/task.schema.json --output ./pkg/task/schema/schema.go

generate_json_schema: generate_json_schema_config generate_json_schema_task

generate_mocks:
ifeq (, $(shell which mockgen))
	go install go.uber.org/mock/mockgen@latest
endif
	mkdir -p pkg/mock
	rm -f pkg/mock/*.go
	mockgen -package mock -source pkg/filter/filter.go > pkg/mock/filter.go
	mockgen -package mock -source pkg/git/git.go > pkg/mock/git.go
	mockgen -package mock -source pkg/host/host.go > pkg/mock/host.go
	mockgen -package mock -source pkg/task/task.go > pkg/mock/task.go

test_cover:
	go test -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html
	rm cover.out

docker_build:
	docker build -t ghcr.io/wndhydrnt/saturn-bot:${VERSION} .

docker_build_full: docker_build
	docker build --build-arg="BASE=${VERSION}" -t ghcr.io/wndhydrnt/saturn-bot:${VERSION}-full -f full.Dockerfile .

bin/go-jsonschema/go-jsonschema-${GO_JSONSCHEMA_VERSION}:
	mkdir -p bin/go-jsonschema
ifeq (${OS}-${ARCH},Darwin-arm64)
	curl -L --silent --fail -o bin/go-jsonschema/go-jsonschema.tar.gz 'https://github.com/omissis/go-jsonschema/releases/download/${GO_JSONSCHEMA_VERSION}/go-jsonschema_Darwin_arm64.tar.gz'
endif
ifeq (${OS}-${ARCH},Darwin-amd64)
	curl -L --silent --fail -o bin/go-jsonschema/go-jsonschema.tar.gz 'https://github.com/omissis/go-jsonschema/releases/download/${GO_JSONSCHEMA_VERSION}/go-jsonschema_Darwin_amd64.tar.gz'
endif
ifeq (${OS}-${ARCH},Linux-aarch64)
	curl -L --silent --fail -o bin/go-jsonschema/go-jsonschema.tar.gz 'https://github.com/omissis/go-jsonschema/releases/download/${GO_JSONSCHEMA_VERSION}/go-jsonschema_Linux_arm64.tar.gz'
endif
ifeq (${OS}-${ARCH},Linux-x86_64)
	curl -L --silent --fail -o bin/go-jsonschema/go-jsonschema.tar.gz 'https://github.com/omissis/go-jsonschema/releases/download/${GO_JSONSCHEMA_VERSION}/go-jsonschema_Linux_x86_64.tar.gz'
endif
	tar -C bin/go-jsonschema/ -xzf bin/go-jsonschema/go-jsonschema.tar.gz
	mv bin/go-jsonschema/go-jsonschema bin/go-jsonschema/go-jsonschema-${GO_JSONSCHEMA_VERSION}
	rm bin/go-jsonschema/go-jsonschema.tar.gz
