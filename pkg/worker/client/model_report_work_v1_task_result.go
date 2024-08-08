/*
saturn-bot server API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the ReportWorkV1TaskResult type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ReportWorkV1TaskResult{}

// ReportWorkV1TaskResult Result of the run of a task.
type ReportWorkV1TaskResult struct {
	// Error encountered during the run, if any.
	Error *string `json:"error,omitempty"`
	// Name of the repository.
	RepositoryName string `json:"repositoryName"`
	// Identifier of the result.
	Result int32 `json:"result"`
	// Name of the task.
	TaskName string `json:"taskName"`
}

type _ReportWorkV1TaskResult ReportWorkV1TaskResult

// NewReportWorkV1TaskResult instantiates a new ReportWorkV1TaskResult object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewReportWorkV1TaskResult(repositoryName string, result int32, taskName string) *ReportWorkV1TaskResult {
	this := ReportWorkV1TaskResult{}
	this.RepositoryName = repositoryName
	this.Result = result
	this.TaskName = taskName
	return &this
}

// NewReportWorkV1TaskResultWithDefaults instantiates a new ReportWorkV1TaskResult object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewReportWorkV1TaskResultWithDefaults() *ReportWorkV1TaskResult {
	this := ReportWorkV1TaskResult{}
	return &this
}

// GetError returns the Error field value if set, zero value otherwise.
func (o *ReportWorkV1TaskResult) GetError() string {
	if o == nil || IsNil(o.Error) {
		var ret string
		return ret
	}
	return *o.Error
}

// GetErrorOk returns a tuple with the Error field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ReportWorkV1TaskResult) GetErrorOk() (*string, bool) {
	if o == nil || IsNil(o.Error) {
		return nil, false
	}
	return o.Error, true
}

// HasError returns a boolean if a field has been set.
func (o *ReportWorkV1TaskResult) HasError() bool {
	if o != nil && !IsNil(o.Error) {
		return true
	}

	return false
}

// SetError gets a reference to the given string and assigns it to the Error field.
func (o *ReportWorkV1TaskResult) SetError(v string) {
	o.Error = &v
}

// GetRepositoryName returns the RepositoryName field value
func (o *ReportWorkV1TaskResult) GetRepositoryName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RepositoryName
}

// GetRepositoryNameOk returns a tuple with the RepositoryName field value
// and a boolean to check if the value has been set.
func (o *ReportWorkV1TaskResult) GetRepositoryNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RepositoryName, true
}

// SetRepositoryName sets field value
func (o *ReportWorkV1TaskResult) SetRepositoryName(v string) {
	o.RepositoryName = v
}

// GetResult returns the Result field value
func (o *ReportWorkV1TaskResult) GetResult() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Result
}

// GetResultOk returns a tuple with the Result field value
// and a boolean to check if the value has been set.
func (o *ReportWorkV1TaskResult) GetResultOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Result, true
}

// SetResult sets field value
func (o *ReportWorkV1TaskResult) SetResult(v int32) {
	o.Result = v
}

// GetTaskName returns the TaskName field value
func (o *ReportWorkV1TaskResult) GetTaskName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TaskName
}

// GetTaskNameOk returns a tuple with the TaskName field value
// and a boolean to check if the value has been set.
func (o *ReportWorkV1TaskResult) GetTaskNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TaskName, true
}

// SetTaskName sets field value
func (o *ReportWorkV1TaskResult) SetTaskName(v string) {
	o.TaskName = v
}

func (o ReportWorkV1TaskResult) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ReportWorkV1TaskResult) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Error) {
		toSerialize["error"] = o.Error
	}
	toSerialize["repositoryName"] = o.RepositoryName
	toSerialize["result"] = o.Result
	toSerialize["taskName"] = o.TaskName
	return toSerialize, nil
}

func (o *ReportWorkV1TaskResult) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"repositoryName",
		"result",
		"taskName",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varReportWorkV1TaskResult := _ReportWorkV1TaskResult{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varReportWorkV1TaskResult)

	if err != nil {
		return err
	}

	*o = ReportWorkV1TaskResult(varReportWorkV1TaskResult)

	return err
}

type NullableReportWorkV1TaskResult struct {
	value *ReportWorkV1TaskResult
	isSet bool
}

func (v NullableReportWorkV1TaskResult) Get() *ReportWorkV1TaskResult {
	return v.value
}

func (v *NullableReportWorkV1TaskResult) Set(val *ReportWorkV1TaskResult) {
	v.value = val
	v.isSet = true
}

func (v NullableReportWorkV1TaskResult) IsSet() bool {
	return v.isSet
}

func (v *NullableReportWorkV1TaskResult) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableReportWorkV1TaskResult(val *ReportWorkV1TaskResult) *NullableReportWorkV1TaskResult {
	return &NullableReportWorkV1TaskResult{value: val, isSet: true}
}

func (v NullableReportWorkV1TaskResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableReportWorkV1TaskResult) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

