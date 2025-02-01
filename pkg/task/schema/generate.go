package schema

//go:generate go run -modfile=../../../tools/go.mod github.com/atombender/go-jsonschema --extra-imports -p schema -t ./task.schema.json --output ./schema.gen.go
