package host

//go:generate go run -modfile=../../../tools/go.mod go.uber.org/mock/mockgen -package host -source ../../../pkg/host/host.go -destination host.gen.go
//go:generate go run -modfile=../../../tools/go.mod go.uber.org/mock/mockgen -package host -source ../../../pkg/host/pullrequestcache.go -destination pullrequestcache.gen.go
