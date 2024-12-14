package error

import "fmt"

const (
	ClientIDTaskNotFound = iota + 1000
	ClientIDInputMissing
)

// Client defines an interface for errors caused by invalid inputs sent by a client.
type Client interface {
	Client()
	Error() string
	ErrorID() int
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

// NewTaskNotFoundError returns a client error that indicates that a task hasn't been found.
func NewTaskNotFoundError(taskName string) Client {
	return client{ID: ClientIDTaskNotFound, Message: fmt.Sprintf("unknown task: %s", taskName)}
}

// NewInputMissingError returns a client error that indicates that an expected input of a task isn't set.
func NewInputMissingError(inputName string, taskName string) Client {
	return client{ID: ClientIDInputMissing, Message: fmt.Sprintf("required input %s not set for task %s", inputName, taskName)}
}
