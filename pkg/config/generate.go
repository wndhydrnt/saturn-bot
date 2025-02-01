package config

//go:generate go run -modfile=../../tools/go.mod github.com/atombender/go-jsonschema --extra-imports -p config -t ./config.schema.json --output ./schema.gen.go
