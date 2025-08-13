package host

//go:generate go tool go.uber.org/mock/mockgen -package host -source ../../../pkg/host/host.go -destination host.gen.go
//go:generate go tool go.uber.org/mock/mockgen -package host -source ../../../pkg/host/pullrequestcache.go -destination pullrequestcache.gen.go
