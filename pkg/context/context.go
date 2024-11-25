package context

import (
	"context"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"go.uber.org/zap"
)

type CheckoutPath struct{}
type PullRequestKey struct{}
type RepositoryKey struct{}
type runDataKey struct{}

// RunData reads and returns plugin data from the context.
// It initialize the map if the context does not contain plugin data.
func RunData(ctx context.Context) map[string]string {
	data, ok := ctx.Value(runDataKey{}).(map[string]string)
	if !ok {
		return make(map[string]string)
	}

	return data
}

// WithRunData adds run data to the [context.Context].
// If the context already contains run data, then it updates the existing map.
// If a key in the map already exists then it gets overwritten.
func WithRunData(ctx context.Context, data map[string]string) context.Context {
	runData, ok := ctx.Value(runDataKey{}).(map[string]string)
	if !ok {
		runData = make(map[string]string)
	}

	// Copy to not modify the original map.
	for k, v := range data {
		runData[k] = v
	}

	return context.WithValue(ctx, runDataKey{}, runData)
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
