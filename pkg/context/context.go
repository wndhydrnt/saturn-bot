package context

import "context"

type CheckoutPath struct{}
type PluginDataKey struct{}
type PullRequestKey struct{}
type RepositoryKey struct{}
type TemplateVarsKey struct{}

// PluginData reads and returns plugin data from the context.
// It initialize the map if the context does not contain plugin data.
func PluginData(ctx context.Context) map[string]string {
	data, ok := ctx.Value(PluginDataKey{}).(map[string]string)
	if !ok {
		return make(map[string]string)
	}

	return data
}
