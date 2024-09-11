package action

import (
	"context"
	"fmt"
	"time"
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
	Create(params Params, taskPath string) (Action, error)
	Name() string
}

type Params map[string]any

func (p Params) Duration(key string, def time.Duration) (time.Duration, error) {
	if p[key] == nil {
		return def, nil
	}

	val, ok := p[key].(string)
	if !ok {
		return def, fmt.Errorf("parameter `%s` is of type %T not string", key, p[key])
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return def, fmt.Errorf("parameter `%s` is not a Go duration", key)
	}

	return d, nil
}

func (p Params) String(key string, def string) (string, error) {
	if p[key] == nil {
		return def, nil
	}

	val, ok := p[key].(string)
	if !ok {
		return def, fmt.Errorf("parameter `%s` is of type %T not string", key, p[key])
	}

	return val, nil
}
