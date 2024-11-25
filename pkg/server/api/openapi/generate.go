package openapi

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./server.gen.go -package openapi -generate types,chi-server,strict-server ./openapi.yaml
