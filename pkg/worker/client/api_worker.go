/*
saturn-bot server API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
)


type WorkerAPI interface {

	/*
	GetWorkV1 Get a unit of work.

	Let a worker retrieve the next unit of work.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@return ApiGetWorkV1Request
	*/
	GetWorkV1(ctx context.Context) ApiGetWorkV1Request

	// GetWorkV1Execute executes the request
	//  @return GetWorkV1Response
	GetWorkV1Execute(r ApiGetWorkV1Request) (*GetWorkV1Response, *http.Response, error)

	/*
	ReportWorkV1 Report the result of a unit of work

	Used by workers to report the result of a unit of work.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@return ApiReportWorkV1Request
	*/
	ReportWorkV1(ctx context.Context) ApiReportWorkV1Request

	// ReportWorkV1Execute executes the request
	//  @return ReportWorkV1Response
	ReportWorkV1Execute(r ApiReportWorkV1Request) (*ReportWorkV1Response, *http.Response, error)

	/*
	ScheduleRunV1 Schedule a run.

	Add a new run to the queue.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@return ApiScheduleRunV1Request
	*/
	ScheduleRunV1(ctx context.Context) ApiScheduleRunV1Request

	// ScheduleRunV1Execute executes the request
	//  @return ScheduleRunV1Response
	ScheduleRunV1Execute(r ApiScheduleRunV1Request) (*ScheduleRunV1Response, *http.Response, error)
}

// WorkerAPIService WorkerAPI service
type WorkerAPIService service

type ApiGetWorkV1Request struct {
	ctx context.Context
	ApiService WorkerAPI
}

func (r ApiGetWorkV1Request) Execute() (*GetWorkV1Response, *http.Response, error) {
	return r.ApiService.GetWorkV1Execute(r)
}

/*
GetWorkV1 Get a unit of work.

Let a worker retrieve the next unit of work.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiGetWorkV1Request
*/
func (a *WorkerAPIService) GetWorkV1(ctx context.Context) ApiGetWorkV1Request {
	return ApiGetWorkV1Request{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return GetWorkV1Response
func (a *WorkerAPIService) GetWorkV1Execute(r ApiGetWorkV1Request) (*GetWorkV1Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *GetWorkV1Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkerAPIService.GetWorkV1")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/worker/work"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiReportWorkV1Request struct {
	ctx context.Context
	ApiService WorkerAPI
	reportWorkV1Request *ReportWorkV1Request
}

func (r ApiReportWorkV1Request) ReportWorkV1Request(reportWorkV1Request ReportWorkV1Request) ApiReportWorkV1Request {
	r.reportWorkV1Request = &reportWorkV1Request
	return r
}

func (r ApiReportWorkV1Request) Execute() (*ReportWorkV1Response, *http.Response, error) {
	return r.ApiService.ReportWorkV1Execute(r)
}

/*
ReportWorkV1 Report the result of a unit of work

Used by workers to report the result of a unit of work.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiReportWorkV1Request
*/
func (a *WorkerAPIService) ReportWorkV1(ctx context.Context) ApiReportWorkV1Request {
	return ApiReportWorkV1Request{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return ReportWorkV1Response
func (a *WorkerAPIService) ReportWorkV1Execute(r ApiReportWorkV1Request) (*ReportWorkV1Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ReportWorkV1Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkerAPIService.ReportWorkV1")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/worker/work"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.reportWorkV1Request == nil {
		return localVarReturnValue, nil, reportError("reportWorkV1Request is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.reportWorkV1Request
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiScheduleRunV1Request struct {
	ctx context.Context
	ApiService WorkerAPI
	scheduleRunV1Request *ScheduleRunV1Request
}

func (r ApiScheduleRunV1Request) ScheduleRunV1Request(scheduleRunV1Request ScheduleRunV1Request) ApiScheduleRunV1Request {
	r.scheduleRunV1Request = &scheduleRunV1Request
	return r
}

func (r ApiScheduleRunV1Request) Execute() (*ScheduleRunV1Response, *http.Response, error) {
	return r.ApiService.ScheduleRunV1Execute(r)
}

/*
ScheduleRunV1 Schedule a run.

Add a new run to the queue.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiScheduleRunV1Request
*/
func (a *WorkerAPIService) ScheduleRunV1(ctx context.Context) ApiScheduleRunV1Request {
	return ApiScheduleRunV1Request{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return ScheduleRunV1Response
func (a *WorkerAPIService) ScheduleRunV1Execute(r ApiScheduleRunV1Request) (*ScheduleRunV1Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ScheduleRunV1Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkerAPIService.ScheduleRunV1")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/runs"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.scheduleRunV1Request == nil {
		return localVarReturnValue, nil, reportError("scheduleRunV1Request is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.scheduleRunV1Request
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
