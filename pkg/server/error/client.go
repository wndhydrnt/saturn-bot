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
	Client()
	Error() string
	ErrorID() int
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

func (e client) ToApiError() openapi.Error {
	return openapi.Error{Errors: []openapi.ErrorDetail{
		{Error: e.ID, Message: e.Message},
	}}
}

// NewTaskNotFoundError returns a client error that indicates that a task hasn't been found.
func NewTaskNotFoundError(taskName string) Client {
	return client{ID: ClientIDTaskNotFound, Message: "unknown task"}
}

type InputError struct {
	client

	errors []error
}

func (e InputError) Errors() []error {
	return e.errors
}

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
