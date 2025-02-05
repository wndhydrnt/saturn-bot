package error

import (
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

const (
	ClientIDTaskNotFound = iota + 1000
	ClientIDInput
	ClientIDRunNotFound
)

// Client defines an interface for errors caused by invalid inputs sent by a client.
type Client interface {
	// Client indicates that this is a client error.
	Client()
	// Error implements [error].
	// It returns the human-readable message of the error.
	Error() string
	// ErrorID returns the internal identifier of the error.
	ErrorID() int
	// ToApiError is a helper method that maps the error to an [github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi.Error].
	ToApiError() openapi.Error
}

type client struct {
	ID      int
	Message string
}

// Client implements [Client].
func (e client) Client() {}

// Error implements [error].
func (e client) Error() string {
	return e.Message
}

// ErrorID implements [Client].
func (e client) ErrorID() int {
	return e.ID
}

// ToApiError implements [ToApiError].
func (e client) ToApiError() openapi.Error {
	return openapi.Error{Errors: []openapi.ErrorDetail{
		{Error: e.ID, Message: e.Message},
	}}
}

// NewTaskNotFoundError returns a client error that indicates that a task hasn't been found.
func NewTaskNotFoundError(taskName string) Client {
	return client{ID: ClientIDTaskNotFound, Message: "unknown task"}
}

// InputError is a specific implementation of [Client].
// It wraps all errors that occurred during validation of inputs.
type InputError struct {
	client

	errors []error
}

// ToApiError implements [Client].
func (e InputError) ToApiError() openapi.Error {
	const msg = "missing required input"
	var details []openapi.ErrorDetail
	for _, err := range e.errors {
		details = append(details, openapi.ErrorDetail{
			Detail:  ptr.To(err.Error()),
			Message: msg,
			Error:   ClientIDInput,
		})
	}

	return openapi.Error{Errors: details}
}

// NewInputError returns a client error that indicates that an expected input of a task isn't set.
func NewInputError(errors []error, taskName string) InputError {
	return InputError{
		client: client{ID: ClientIDInput, Message: "missing inputs for task " + taskName},
		errors: errors,
	}
}

// NewTaskNotFoundError returns a client error that indicates that the run identified by id doesn't exist.
func NewRunNotFoundError(id int) Client {
	return client{ID: ClientIDRunNotFound, Message: "unknown run"}
}
