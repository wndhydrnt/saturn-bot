package action

import "context"

var (
	BuiltInFactories = []Factory{
		FileCreateFactory{},
		FileDeleteFactory{},
		LineDeleteFactory{},
		LineInsertFactory{},
		LineReplaceFactory{},
	}
)

type Action interface {
	Apply(ctx context.Context) error
	String() string
}

type Factory interface {
	Create(params map[string]any, taskPath string) (Action, error)
	Name() string
}
