package api

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed openapi/openapi.yaml
	openApiDef  string
	serverError = openapi.Error{Error: "Internal Server Error", Message: ""}
)

func RegisterOpenAPIDefinitionRoute(baseURL string, router chi.Router) error {
	h, err := NewOpenApiDefinitionHandler(baseURL)
	if err != nil {
		return err
	}

	router.Get("/openapi.yaml", h)
	return nil
}

func NewOpenApiDefinitionHandler(baseURL string) (http.HandlerFunc, error) {
	dec := yaml.NewDecoder(strings.NewReader(openApiDef))
	var data map[string]interface{}
	err := dec.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decode OpenAPI definition from YAML: %w", err)
	}

	data["servers"] = []map[string]string{
		{"url": baseURL, "description": "saturn-bot API server"},
	}

	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	err = enc.Encode(&data)
	if err != nil {
		return nil, fmt.Errorf("encode OpenAPI definition back to YAML: %w", err)
	}

	err = enc.Close()
	if err != nil {
		return nil, fmt.Errorf("close YAML encoder of OpenAPI definition: %w", err)
	}

	content := buf.Bytes()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(content)
	}, nil
}

func RegisterHealthRoute(router chi.Router) {
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		const up = "UP"
		_, _ = fmt.Fprint(w, up)
	})
}
