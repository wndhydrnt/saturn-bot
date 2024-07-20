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

// checks if the GetTaskV1200Response type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetTaskV1200Response{}

// GetTaskV1200Response struct for GetTaskV1200Response
type GetTaskV1200Response struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Content string `json:"content"`
}

type _GetTaskV1200Response GetTaskV1200Response

// NewGetTaskV1200Response instantiates a new GetTaskV1200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetTaskV1200Response(name string, hash string, content string) *GetTaskV1200Response {
	this := GetTaskV1200Response{}
	this.Name = name
	this.Hash = hash
	this.Content = content
	return &this
}

// NewGetTaskV1200ResponseWithDefaults instantiates a new GetTaskV1200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetTaskV1200ResponseWithDefaults() *GetTaskV1200Response {
	this := GetTaskV1200Response{}
	return &this
}

// GetName returns the Name field value
func (o *GetTaskV1200Response) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *GetTaskV1200Response) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *GetTaskV1200Response) SetName(v string) {
	o.Name = v
}

// GetHash returns the Hash field value
func (o *GetTaskV1200Response) GetHash() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Hash
}

// GetHashOk returns a tuple with the Hash field value
// and a boolean to check if the value has been set.
func (o *GetTaskV1200Response) GetHashOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Hash, true
}

// SetHash sets field value
func (o *GetTaskV1200Response) SetHash(v string) {
	o.Hash = v
}

// GetContent returns the Content field value
func (o *GetTaskV1200Response) GetContent() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Content
}

// GetContentOk returns a tuple with the Content field value
// and a boolean to check if the value has been set.
func (o *GetTaskV1200Response) GetContentOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Content, true
}

// SetContent sets field value
func (o *GetTaskV1200Response) SetContent(v string) {
	o.Content = v
}

func (o GetTaskV1200Response) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetTaskV1200Response) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["hash"] = o.Hash
	toSerialize["content"] = o.Content
	return toSerialize, nil
}

func (o *GetTaskV1200Response) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"name",
		"hash",
		"content",
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

	varGetTaskV1200Response := _GetTaskV1200Response{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varGetTaskV1200Response)

	if err != nil {
		return err
	}

	*o = GetTaskV1200Response(varGetTaskV1200Response)

	return err
}

type NullableGetTaskV1200Response struct {
	value *GetTaskV1200Response
	isSet bool
}

func (v NullableGetTaskV1200Response) Get() *GetTaskV1200Response {
	return v.value
}

func (v *NullableGetTaskV1200Response) Set(val *GetTaskV1200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableGetTaskV1200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableGetTaskV1200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetTaskV1200Response(val *GetTaskV1200Response) *NullableGetTaskV1200Response {
	return &NullableGetTaskV1200Response{value: val, isSet: true}
}

func (v NullableGetTaskV1200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetTaskV1200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


