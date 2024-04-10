VERSION?=dev
VERSION_HASH?=$(shell git rev-parse HEAD)
VERSION_DATETIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_FLAGS=-ldflags="-X 'github.com/wndhydrnt/saturn-sync/pkg/version.Version=$(VERSION)' -X 'github.com/wndhydrnt/saturn-sync/pkg/version.Hash=$(VERSION_HASH)' -X 'github.com/wndhydrnt/saturn-sync/pkg/version.DateTime=$(VERSION_DATETIME)'"

build:
	go build $(BUILD_FLAGS) -o saturn-sync

build_darwin_amd64:
	GOARCH=amd64 GOOS=darwin go build $(BUILD_FLAGS) -o saturn-sync-$(VERSION).darwin-amd64

build_darwin_arm64:
	GOARCH=arm64 GOOS=darwin go build $(BUILD_FLAGS) -o saturn-sync-$(VERSION).darwin-arm64

build_linux_arm64:
	GOARCH=arm64 GOOS=linux go build $(BUILD_FLAGS) -o saturn-sync-$(VERSION).linux-arm64
	cp saturn-sync-$(VERSION).linux-arm64 saturn-sync-$(VERSION).linux-aarch64

build_linux_armv7:
	GOARCH=arm GOOS=linux go build $(BUILD_FLAGS) -o saturn-sync-$(VERSION).linux-armv7

build_linux_amd64:
	GOARCH=amd64 GOOS=linux go build $(BUILD_FLAGS) -o saturn-sync-$(VERSION).linux-amd64

build_all: build_darwin_amd64 build_darwin_arm64 build_linux_arm64 build_linux_armv7 build_linux_amd64

generate_grpc:
	buf generate
	protoc-go-inject-tag -input ./pkg/proto/saturnsync.pb.go -remove_tag_comment
	gofmt -w ./pkg/proto/saturnsync.pb.go

generate_mocks:
	mkdir -p pkg/mock
	rm pkg/mock/*.go
	mockgen -package mock -source pkg/filter/filter.go > pkg/mock/filter.go
	mockgen -package mock -source pkg/git/git.go > pkg/mock/git.go
	mockgen -package mock -source pkg/host/host.go > pkg/mock/host.go
	mockgen -package mock -source pkg/task/task.go > pkg/mock/task.go

test_cover:
	go test -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html
	rm cover.out
