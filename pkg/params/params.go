package params

import (
	"fmt"
	"time"
)

// Params is the data container that holds the parameters set for an action.
type Params map[string]any

// Duration parses a parameter into a Go duration.
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

// String parses a parameter into a string.
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

func (p Params) StringSlice(key string, def []string) ([]string, error) {
	if p[key] == nil {
		return def, nil
	}

	rawV, ok := p[key].([]interface{})
	if !ok {
		return def, fmt.Errorf("parameter `%s` is of type %T not slice", key, p[key])
	}

	var vals []string
	for idx, rawItem := range rawV {
		v, ok := rawItem.(string)
		if !ok {
			return def, fmt.Errorf("parameter `%s[%d]` is of type %T not string", key, idx, rawItem)
		}

		vals = append(vals, v)
	}

	return vals, nil
}
