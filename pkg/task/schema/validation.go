package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adhocore/gronx"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	schemaFile = "task.schema.json"
)

var (
	//go:embed task.schema.json
	schemaRaw string
)

func init() {
	// Add custom cron format.
	jsonschema.Formats["cron"] = isCron
}

// isCron checks if v is a valid cron expression
func isCron(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}

	return gronx.IsValid(s)
}

func Validate(t *Task) error {
	compiler := jsonschema.NewCompiler()
	// Set to true to validate custom formats, like cron
	compiler.AssertFormat = true
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
