//go:build tools
// +build tools

package main

import (
	_ "github.com/atombender/go-jsonschema"
	_ "github.com/bwplotka/mdox"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "go.uber.org/mock/mockgen"
	_ "golang.org/x/tools/cmd/stringer"
)
