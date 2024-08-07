/*
saturn-bot server API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
	"time"
	"bytes"
	"fmt"
)

// checks if the ScheduleRunV1Request type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ScheduleRunV1Request{}

// ScheduleRunV1Request struct for ScheduleRunV1Request
type ScheduleRunV1Request struct {
	// Name of the repository for which to add a run. If empty, the run uses the filters of the task.
	RepositoryName *string `json:"repositoryName,omitempty"`
	// Schedule the run after the given time. If empty, then the current time is used.
	ScheduleAfter *time.Time `json:"scheduleAfter,omitempty"`
	// Name of the task for which to add a run.
	TaskName string `json:"taskName"`
	RunData map[string]string `json:"runData,omitempty"`
}

type _ScheduleRunV1Request ScheduleRunV1Request

// NewScheduleRunV1Request instantiates a new ScheduleRunV1Request object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewScheduleRunV1Request(taskName string) *ScheduleRunV1Request {
	this := ScheduleRunV1Request{}
	this.TaskName = taskName
	return &this
}

// NewScheduleRunV1RequestWithDefaults instantiates a new ScheduleRunV1Request object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewScheduleRunV1RequestWithDefaults() *ScheduleRunV1Request {
	this := ScheduleRunV1Request{}
	return &this
}

// GetRepositoryName returns the RepositoryName field value if set, zero value otherwise.
func (o *ScheduleRunV1Request) GetRepositoryName() string {
	if o == nil || IsNil(o.RepositoryName) {
		var ret string
		return ret
	}
	return *o.RepositoryName
}

// GetRepositoryNameOk returns a tuple with the RepositoryName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ScheduleRunV1Request) GetRepositoryNameOk() (*string, bool) {
	if o == nil || IsNil(o.RepositoryName) {
		return nil, false
	}
	return o.RepositoryName, true
}

// HasRepositoryName returns a boolean if a field has been set.
func (o *ScheduleRunV1Request) HasRepositoryName() bool {
	if o != nil && !IsNil(o.RepositoryName) {
		return true
	}

	return false
}

// SetRepositoryName gets a reference to the given string and assigns it to the RepositoryName field.
func (o *ScheduleRunV1Request) SetRepositoryName(v string) {
	o.RepositoryName = &v
}

// GetScheduleAfter returns the ScheduleAfter field value if set, zero value otherwise.
func (o *ScheduleRunV1Request) GetScheduleAfter() time.Time {
	if o == nil || IsNil(o.ScheduleAfter) {
		var ret time.Time
		return ret
	}
	return *o.ScheduleAfter
}

// GetScheduleAfterOk returns a tuple with the ScheduleAfter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ScheduleRunV1Request) GetScheduleAfterOk() (*time.Time, bool) {
	if o == nil || IsNil(o.ScheduleAfter) {
		return nil, false
	}
	return o.ScheduleAfter, true
}

// HasScheduleAfter returns a boolean if a field has been set.
func (o *ScheduleRunV1Request) HasScheduleAfter() bool {
	if o != nil && !IsNil(o.ScheduleAfter) {
		return true
	}

	return false
}

// SetScheduleAfter gets a reference to the given time.Time and assigns it to the ScheduleAfter field.
func (o *ScheduleRunV1Request) SetScheduleAfter(v time.Time) {
	o.ScheduleAfter = &v
}

// GetTaskName returns the TaskName field value
func (o *ScheduleRunV1Request) GetTaskName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TaskName
}

// GetTaskNameOk returns a tuple with the TaskName field value
// and a boolean to check if the value has been set.
func (o *ScheduleRunV1Request) GetTaskNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TaskName, true
}

// SetTaskName sets field value
func (o *ScheduleRunV1Request) SetTaskName(v string) {
	o.TaskName = v
}

// GetRunData returns the RunData field value if set, zero value otherwise.
func (o *ScheduleRunV1Request) GetRunData() map[string]string {
	if o == nil || IsNil(o.RunData) {
		var ret map[string]string
		return ret
	}
	return o.RunData
}

// GetRunDataOk returns a tuple with the RunData field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ScheduleRunV1Request) GetRunDataOk() (map[string]string, bool) {
	if o == nil || IsNil(o.RunData) {
		return map[string]string{}, false
	}
	return o.RunData, true
}

// HasRunData returns a boolean if a field has been set.
func (o *ScheduleRunV1Request) HasRunData() bool {
	if o != nil && !IsNil(o.RunData) {
		return true
	}

	return false
}

// SetRunData gets a reference to the given map[string]string and assigns it to the RunData field.
func (o *ScheduleRunV1Request) SetRunData(v map[string]string) {
	o.RunData = v
}

func (o ScheduleRunV1Request) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ScheduleRunV1Request) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.RepositoryName) {
		toSerialize["repositoryName"] = o.RepositoryName
	}
	if !IsNil(o.ScheduleAfter) {
		toSerialize["scheduleAfter"] = o.ScheduleAfter
	}
	toSerialize["taskName"] = o.TaskName
	if !IsNil(o.RunData) {
		toSerialize["runData"] = o.RunData
	}
	return toSerialize, nil
}

func (o *ScheduleRunV1Request) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
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

	varScheduleRunV1Request := _ScheduleRunV1Request{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varScheduleRunV1Request)

	if err != nil {
		return err
	}

	*o = ScheduleRunV1Request(varScheduleRunV1Request)

	return err
}

type NullableScheduleRunV1Request struct {
	value *ScheduleRunV1Request
	isSet bool
}

func (v NullableScheduleRunV1Request) Get() *ScheduleRunV1Request {
	return v.value
}

func (v *NullableScheduleRunV1Request) Set(val *ScheduleRunV1Request) {
	v.value = val
	v.isSet = true
}

func (v NullableScheduleRunV1Request) IsSet() bool {
	return v.isSet
}

func (v *NullableScheduleRunV1Request) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableScheduleRunV1Request(val *ScheduleRunV1Request) *NullableScheduleRunV1Request {
	return &NullableScheduleRunV1Request{value: val, isSet: true}
}

func (v NullableScheduleRunV1Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableScheduleRunV1Request) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


