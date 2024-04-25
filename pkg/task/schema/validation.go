package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	schemaFile = "task.schema.json"
)

var (
	//go:embed task.schema.json
	schemaRaw string
)

func Validate(t *Task) error {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	err := compiler.AddResource(schemaFile, strings.NewReader(schemaRaw))
	if err != nil {
		return fmt.Errorf("add resource to JSON schema compiler: %w", err)
	}

	jsonSchema, err := compiler.Compile(schemaFile)
	if err != nil {
		return fmt.Errorf("compile json schema: %w", err)
	}

	b, _ := json.Marshal(t)
	var validation interface{}
	_ = json.Unmarshal(b, &validation)
	return jsonSchema.Validate(validation)
}
