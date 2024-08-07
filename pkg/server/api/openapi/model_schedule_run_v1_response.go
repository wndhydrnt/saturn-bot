// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * saturn-bot server API
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0.0
 */

package openapi




type ScheduleRunV1Response struct {

	// Identifier of the newly scheduled run.
	RunID int32 `json:"runID"`
}

// AssertScheduleRunV1ResponseRequired checks if the required fields are not zero-ed
func AssertScheduleRunV1ResponseRequired(obj ScheduleRunV1Response) error {
	elements := map[string]interface{}{
		"runID": obj.RunID,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertScheduleRunV1ResponseConstraints checks if the values respects the defined constraints
func AssertScheduleRunV1ResponseConstraints(obj ScheduleRunV1Response) error {
	return nil
}
