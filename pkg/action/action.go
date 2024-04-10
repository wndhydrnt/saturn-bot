package action

import "context"

type Action interface {
	Apply(ctx context.Context) error
	String() string
}
