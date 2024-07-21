// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * saturn-bot server API
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0.0
 */

package openapi


import (
	"time"
)



type ScheduleRunV1Request struct {

	// Name of the repository for which to add a run. If empty, the run uses the filters of the task.
	RepositoryName string `json:"repositoryName,omitempty"`

	// Schedule the run after the given time. If empty, then the current time is used.
	ScheduleAfter time.Time `json:"scheduleAfter,omitempty"`

	// Name of the task for which to add a run.
	TaskName string `json:"taskName"`
}

// AssertScheduleRunV1RequestRequired checks if the required fields are not zero-ed
func AssertScheduleRunV1RequestRequired(obj ScheduleRunV1Request) error {
	elements := map[string]interface{}{
		"taskName": obj.TaskName,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertScheduleRunV1RequestConstraints checks if the values respects the defined constraints
func AssertScheduleRunV1RequestConstraints(obj ScheduleRunV1Request) error {
	return nil
}
