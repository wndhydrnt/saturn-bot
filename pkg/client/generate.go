package client

//go:generate go run -modfile=../../tools/go.mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate models,client -o openapi.gen.go -package client -response-type-suffix ResponseBody ../server/api/openapi/openapi.yaml
