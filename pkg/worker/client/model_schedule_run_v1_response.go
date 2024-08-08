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

// checks if the ScheduleRunV1Response type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ScheduleRunV1Response{}

// ScheduleRunV1Response struct for ScheduleRunV1Response
type ScheduleRunV1Response struct {
	// Identifier of the newly scheduled run.
	RunID int32 `json:"runID"`
}

type _ScheduleRunV1Response ScheduleRunV1Response

// NewScheduleRunV1Response instantiates a new ScheduleRunV1Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewScheduleRunV1Response(runID int32) *ScheduleRunV1Response {
	this := ScheduleRunV1Response{}
	this.RunID = runID
	return &this
}

// NewScheduleRunV1ResponseWithDefaults instantiates a new ScheduleRunV1Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewScheduleRunV1ResponseWithDefaults() *ScheduleRunV1Response {
	this := ScheduleRunV1Response{}
	return &this
}

// GetRunID returns the RunID field value
func (o *ScheduleRunV1Response) GetRunID() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.RunID
}

// GetRunIDOk returns a tuple with the RunID field value
// and a boolean to check if the value has been set.
func (o *ScheduleRunV1Response) GetRunIDOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RunID, true
}

// SetRunID sets field value
func (o *ScheduleRunV1Response) SetRunID(v int32) {
	o.RunID = v
}

func (o ScheduleRunV1Response) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ScheduleRunV1Response) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["runID"] = o.RunID
	return toSerialize, nil
}

func (o *ScheduleRunV1Response) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"runID",
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

	varScheduleRunV1Response := _ScheduleRunV1Response{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varScheduleRunV1Response)

	if err != nil {
		return err
	}

	*o = ScheduleRunV1Response(varScheduleRunV1Response)

	return err
}

type NullableScheduleRunV1Response struct {
	value *ScheduleRunV1Response
	isSet bool
}

func (v NullableScheduleRunV1Response) Get() *ScheduleRunV1Response {
	return v.value
}

func (v *NullableScheduleRunV1Response) Set(val *ScheduleRunV1Response) {
	v.value = val
	v.isSet = true
}

func (v NullableScheduleRunV1Response) IsSet() bool {
	return v.isSet
}

func (v *NullableScheduleRunV1Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableScheduleRunV1Response(val *ScheduleRunV1Response) *NullableScheduleRunV1Response {
	return &NullableScheduleRunV1Response{value: val, isSet: true}
}

func (v NullableScheduleRunV1Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableScheduleRunV1Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

