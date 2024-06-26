package action

import "context"

type ExecFactory struct{}

func (f ExecFactory) Create(params map[string]string, taskPath string) (Action, error) {

}

type execAction struct {
}

func (a *execAction) Apply(_ context.Context) error {

}
