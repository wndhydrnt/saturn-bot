package context

import "context"

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
