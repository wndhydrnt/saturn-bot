package action

import (
	"context"

	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

var (
	BuiltInFactories = []Factory{
		ExecFactory{},
		FileCreateFactory{},
		FileDeleteFactory{},
		LineDeleteFactory{},
		LineInsertFactory{},
		LineReplaceFactory{},
		ScriptFactory{},
	}
)

type Action interface {
	Apply(ctx context.Context) error
	String() string
}

type Factory interface {
	Create(params params.Params, taskPath string) (Action, error)
	Name() string
}
