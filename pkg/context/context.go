package context

import (
	"context"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"go.uber.org/zap"
)

type CheckoutPath struct{}
type PullRequestKey struct{}
type RepositoryKey struct{}
type RunDataKey struct{}

// RunData reads and returns plugin data from the context.
// It initialize the map if the context does not contain plugin data.
func RunData(ctx context.Context) map[string]string {
	data, ok := ctx.Value(RunDataKey{}).(map[string]string)
	if !ok {
		return make(map[string]string)
	}

	return data
}

type logKey struct{}

// WithLog returns a context with logger added to it.
func WithLog(parent context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(parent, logKey{}, logger)
}

// Log returns the logger from the context.
// If the context doesn't contain a logger then it returns the default logger.
func Log(ctx context.Context) *zap.SugaredLogger {
	v := ctx.Value(logKey{})
	if v == nil {
		return log.Log()
	}

	return v.(*zap.SugaredLogger)
}
