package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

// CustomClientWithResponsesOptions defines options for [NewCustomClientWithResponses].
type CustomClientWithResponsesOptions struct {
	ApiKey     string
	BaseUrl    string
	HttpClient *http.Client
}

// NewCustomClientWithResponses wraps [NewClientWithResponses].
// It provides useful defaults for the API client.
func NewCustomClientWithResponses(opts CustomClientWithResponsesOptions) (ClientWithResponsesInterface, error) {
	var httpClient *http.Client
	if opts.HttpClient == nil {
		httpClient = &http.Client{
			Timeout: 2 * time.Second,
		}
	} else {
		httpClient = opts.HttpClient
	}

	apiClient, err := NewClientWithResponses(
		opts.BaseUrl,
		WithRequestEditorFn(newApiKeyAddFunc(openapi.HeaderApiKey, opts.ApiKey)),
		WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("create custom OpenAPI client: %w", err)
	}

	return apiClient, nil
}

func newApiKeyAddFunc(key, value string) RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	}
}
