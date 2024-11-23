// Package client provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime"
)

// Defines values for ReportWorkV1ResponseResult.
const (
	Ok ReportWorkV1ResponseResult = "ok"
)

// Error defines model for Error.
type Error struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GetTaskV1Response defines model for GetTaskV1Response.
type GetTaskV1Response struct {
	Content string `json:"content"`
	Hash    string `json:"hash"`
	Name    string `json:"name"`
}

// GetWorkV1Response defines model for GetWorkV1Response.
type GetWorkV1Response struct {
	// Repositories Names of repositories for which to apply the tasks.
	Repositories *[]string `json:"repositories,omitempty"`

	// RunID Internal identifier of the unit of work.
	RunID int `json:"runID"`

	// Tasks Names of the tasks to execute.
	Tasks []GetWorkV1Task `json:"tasks"`
}

// GetWorkV1Task defines model for GetWorkV1Task.
type GetWorkV1Task struct {
	// Hash Hash of the task. Used to detect if server and worker are out of sync.
	Hash string `json:"hash"`

	// Name Name of the task to execute.
	Name string `json:"name"`
}

// ListTasksV1Response defines model for ListTasksV1Response.
type ListTasksV1Response struct {
	// Tasks Names of registered tasks.
	Tasks []string `json:"tasks"`
}

// ReportWorkV1Request defines model for ReportWorkV1Request.
type ReportWorkV1Request struct {
	// Error General that occurred during the run, if any.
	Error *string `json:"error,omitempty"`

	// RunID Internal identifier of the unit of work.
	RunID int `json:"runID"`

	// TaskResults Results of each task.
	TaskResults []ReportWorkV1TaskResult `json:"taskResults"`
}

// ReportWorkV1Response defines model for ReportWorkV1Response.
type ReportWorkV1Response struct {
	// Result Indicator of the result of the operation.
	Result ReportWorkV1ResponseResult `json:"result"`
}

// ReportWorkV1ResponseResult Indicator of the result of the operation.
type ReportWorkV1ResponseResult string

// ReportWorkV1TaskResult Result of the run of a task.
type ReportWorkV1TaskResult struct {
	// Error Error encountered during the run, if any.
	Error *string `json:"error,omitempty"`

	// RepositoryName Name of the repository.
	RepositoryName string `json:"repositoryName"`

	// Result Identifier of the result.
	Result int `json:"result"`

	// TaskName Name of the task.
	TaskName string `json:"taskName"`
}

// ScheduleRunV1Request defines model for ScheduleRunV1Request.
type ScheduleRunV1Request struct {
	// RepositoryNames Names of the repositories for which to add a run.
	// Leave empty to schedule a run for all repositories the task matches.
	RepositoryNames *[]string          `json:"repositoryNames,omitempty"`
	RunData         *map[string]string `json:"runData,omitempty"`

	// ScheduleAfter Schedule the run after the given time.
	// Uses the current time if empty.
	ScheduleAfter *time.Time `json:"scheduleAfter,omitempty"`

	// TaskName Name of the task for which to add a run.
	TaskName string `json:"taskName"`
}

// ScheduleRunV1Response defines model for ScheduleRunV1Response.
type ScheduleRunV1Response struct {
	// RunID Identifier of the newly scheduled run.
	RunID int `json:"runID"`
}

// ScheduleRunV1JSONRequestBody defines body for ScheduleRunV1 for application/json ContentType.
type ScheduleRunV1JSONRequestBody = ScheduleRunV1Request

// ReportWorkV1JSONRequestBody defines body for ReportWorkV1 for application/json ContentType.
type ReportWorkV1JSONRequestBody = ReportWorkV1Request

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ScheduleRunV1WithBody request with any body
	ScheduleRunV1WithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ScheduleRunV1(ctx context.Context, body ScheduleRunV1JSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListTasksV1 request
	ListTasksV1(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetTaskV1 request
	GetTaskV1(ctx context.Context, task string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetWorkV1 request
	GetWorkV1(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ReportWorkV1WithBody request with any body
	ReportWorkV1WithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ReportWorkV1(ctx context.Context, body ReportWorkV1JSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) ScheduleRunV1WithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScheduleRunV1RequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ScheduleRunV1(ctx context.Context, body ScheduleRunV1JSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScheduleRunV1Request(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListTasksV1(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListTasksV1Request(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetTaskV1(ctx context.Context, task string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetTaskV1Request(c.Server, task)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetWorkV1(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetWorkV1Request(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ReportWorkV1WithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewReportWorkV1RequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ReportWorkV1(ctx context.Context, body ReportWorkV1JSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewReportWorkV1Request(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewScheduleRunV1Request calls the generic ScheduleRunV1 builder with application/json body
func NewScheduleRunV1Request(server string, body ScheduleRunV1JSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewScheduleRunV1RequestWithBody(server, "application/json", bodyReader)
}

// NewScheduleRunV1RequestWithBody generates requests for ScheduleRunV1 with any type of body
func NewScheduleRunV1RequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/runs")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListTasksV1Request generates requests for ListTasksV1
func NewListTasksV1Request(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/tasks")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetTaskV1Request generates requests for GetTaskV1
func NewGetTaskV1Request(server string, task string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "task", runtime.ParamLocationPath, task)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/tasks/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetWorkV1Request generates requests for GetWorkV1
func NewGetWorkV1Request(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/worker/work")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewReportWorkV1Request calls the generic ReportWorkV1 builder with application/json body
func NewReportWorkV1Request(server string, body ReportWorkV1JSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewReportWorkV1RequestWithBody(server, "application/json", bodyReader)
}

// NewReportWorkV1RequestWithBody generates requests for ReportWorkV1 with any type of body
func NewReportWorkV1RequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/worker/work")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ScheduleRunV1WithBodyWithResponse request with any body
	ScheduleRunV1WithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScheduleRunV1ResponseBody, error)

	ScheduleRunV1WithResponse(ctx context.Context, body ScheduleRunV1JSONRequestBody, reqEditors ...RequestEditorFn) (*ScheduleRunV1ResponseBody, error)

	// ListTasksV1WithResponse request
	ListTasksV1WithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListTasksV1ResponseBody, error)

	// GetTaskV1WithResponse request
	GetTaskV1WithResponse(ctx context.Context, task string, reqEditors ...RequestEditorFn) (*GetTaskV1ResponseBody, error)

	// GetWorkV1WithResponse request
	GetWorkV1WithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetWorkV1ResponseBody, error)

	// ReportWorkV1WithBodyWithResponse request with any body
	ReportWorkV1WithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ReportWorkV1ResponseBody, error)

	ReportWorkV1WithResponse(ctx context.Context, body ReportWorkV1JSONRequestBody, reqEditors ...RequestEditorFn) (*ReportWorkV1ResponseBody, error)
}

type ScheduleRunV1ResponseBody struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ScheduleRunV1Response
}

// Status returns HTTPResponse.Status
func (r ScheduleRunV1ResponseBody) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ScheduleRunV1ResponseBody) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListTasksV1ResponseBody struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ListTasksV1Response
}

// Status returns HTTPResponse.Status
func (r ListTasksV1ResponseBody) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListTasksV1ResponseBody) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetTaskV1ResponseBody struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *GetTaskV1Response
	JSON404      *Error
	JSON500      *Error
}

// Status returns HTTPResponse.Status
func (r GetTaskV1ResponseBody) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetTaskV1ResponseBody) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetWorkV1ResponseBody struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *GetWorkV1Response
}

// Status returns HTTPResponse.Status
func (r GetWorkV1ResponseBody) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetWorkV1ResponseBody) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ReportWorkV1ResponseBody struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *ReportWorkV1Response
}

// Status returns HTTPResponse.Status
func (r ReportWorkV1ResponseBody) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ReportWorkV1ResponseBody) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ScheduleRunV1WithBodyWithResponse request with arbitrary body returning *ScheduleRunV1ResponseBody
func (c *ClientWithResponses) ScheduleRunV1WithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScheduleRunV1ResponseBody, error) {
	rsp, err := c.ScheduleRunV1WithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScheduleRunV1ResponseBody(rsp)
}

func (c *ClientWithResponses) ScheduleRunV1WithResponse(ctx context.Context, body ScheduleRunV1JSONRequestBody, reqEditors ...RequestEditorFn) (*ScheduleRunV1ResponseBody, error) {
	rsp, err := c.ScheduleRunV1(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScheduleRunV1ResponseBody(rsp)
}

// ListTasksV1WithResponse request returning *ListTasksV1ResponseBody
func (c *ClientWithResponses) ListTasksV1WithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListTasksV1ResponseBody, error) {
	rsp, err := c.ListTasksV1(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListTasksV1ResponseBody(rsp)
}

// GetTaskV1WithResponse request returning *GetTaskV1ResponseBody
func (c *ClientWithResponses) GetTaskV1WithResponse(ctx context.Context, task string, reqEditors ...RequestEditorFn) (*GetTaskV1ResponseBody, error) {
	rsp, err := c.GetTaskV1(ctx, task, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetTaskV1ResponseBody(rsp)
}

// GetWorkV1WithResponse request returning *GetWorkV1ResponseBody
func (c *ClientWithResponses) GetWorkV1WithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetWorkV1ResponseBody, error) {
	rsp, err := c.GetWorkV1(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetWorkV1ResponseBody(rsp)
}

// ReportWorkV1WithBodyWithResponse request with arbitrary body returning *ReportWorkV1ResponseBody
func (c *ClientWithResponses) ReportWorkV1WithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ReportWorkV1ResponseBody, error) {
	rsp, err := c.ReportWorkV1WithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseReportWorkV1ResponseBody(rsp)
}

func (c *ClientWithResponses) ReportWorkV1WithResponse(ctx context.Context, body ReportWorkV1JSONRequestBody, reqEditors ...RequestEditorFn) (*ReportWorkV1ResponseBody, error) {
	rsp, err := c.ReportWorkV1(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseReportWorkV1ResponseBody(rsp)
}

// ParseScheduleRunV1ResponseBody parses an HTTP response from a ScheduleRunV1WithResponse call
func ParseScheduleRunV1ResponseBody(rsp *http.Response) (*ScheduleRunV1ResponseBody, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ScheduleRunV1ResponseBody{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ScheduleRunV1Response
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListTasksV1ResponseBody parses an HTTP response from a ListTasksV1WithResponse call
func ParseListTasksV1ResponseBody(rsp *http.Response) (*ListTasksV1ResponseBody, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListTasksV1ResponseBody{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ListTasksV1Response
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetTaskV1ResponseBody parses an HTTP response from a GetTaskV1WithResponse call
func ParseGetTaskV1ResponseBody(rsp *http.Response) (*GetTaskV1ResponseBody, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetTaskV1ResponseBody{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest GetTaskV1Response
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseGetWorkV1ResponseBody parses an HTTP response from a GetWorkV1WithResponse call
func ParseGetWorkV1ResponseBody(rsp *http.Response) (*GetWorkV1ResponseBody, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetWorkV1ResponseBody{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest GetWorkV1Response
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseReportWorkV1ResponseBody parses an HTTP response from a ReportWorkV1WithResponse call
func ParseReportWorkV1ResponseBody(rsp *http.Response) (*ReportWorkV1ResponseBody, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ReportWorkV1ResponseBody{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest ReportWorkV1Response
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	}

	return response, nil
}
