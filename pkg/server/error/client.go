package error

import "fmt"

const (
	ClientIDTaskNotFound = iota + 1000
	ClientIDInputMissing
)

type Client interface {
	Client()
	Error() string
	ErrorID() int
}

type client struct {
	ID      int
	Message string
}

func (e client) Client() {}

func (e client) ErrorID() int {
	return e.ID
}

func (e client) Error() string {
	return e.Message
}

func NewTaskNotFoundError(taskName string) Client {
	return client{ID: ClientIDTaskNotFound, Message: fmt.Sprintf("unknown task: %s", taskName)}
}

func NewInputMissingError(inputName string, taskName string) Client {
	return client{ID: ClientIDInputMissing, Message: fmt.Sprintf("required input %s not set for task %s", inputName, taskName)}
}
