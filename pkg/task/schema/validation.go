package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adhocore/gronx"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	schemaFile = "task.schema.json"
)

var (
	//go:embed task.schema.json
	schemaRaw string

	cronFormat = &jsonschema.Format{
		Name:     "cron",
		Validate: validateCron,
	}
)

// validateCron checks if v is a valid cron expression
func validateCron(v any) error {
	s, ok := v.(string)
	if !ok {
		return nil
	}

	if gronx.IsValid(s) {
		return nil
	}

	return fmt.Errorf("unsupported cron expression: %s", s)
}

func Validate(t *Task) error {
	schemaUnmarshal, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaRaw))
	if err != nil {
		return fmt.Errorf("unmarshal json schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.RegisterFormat(cronFormat)
	// Set to true to validate custom formats, like cron
	compiler.AssertFormat()
	err = compiler.AddResource(schemaFile, schemaUnmarshal)
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
