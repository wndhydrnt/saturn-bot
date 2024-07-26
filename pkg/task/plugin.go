package task

import (
	"context"
	"fmt"
	"hash"
	"log/slog"
	"os/exec"
	"path/filepath"

	goPlugin "github.com/hashicorp/go-plugin"
	"github.com/wndhydrnt/saturn-bot-go/plugin"
	proto "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	gsContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	gsLog "github.com/wndhydrnt/saturn-bot/pkg/log"
)

type startPluginOptions struct {
	hash       hash.Hash
	config     map[string]string
	filePath   string
	pathJava   string
	pathPython string
}

type pluginWrapper struct {
	action   *PluginAction
	client   *goPlugin.Client
	filter   *PluginFilter
	provider plugin.Provider
}

func newPluginWrapper(opts startPluginOptions) (*pluginWrapper, error) {
	ext := filepath.Ext(opts.filePath)
	var args []string
	var executable string
	switch ext {
	case "":
		executable = opts.filePath
	case ".jar":
		executable = opts.pathJava
		args = append(args, "-jar", opts.filePath)
	case ".py":
		executable = opts.pathPython
		args = append(args, opts.filePath)
	}

	client := goPlugin.NewClient(&goPlugin.ClientConfig{
		HandshakeConfig: plugin.Handshake,
		Logger:          gsLog.DefaultHclogAdapter(),
		Plugins: map[string]goPlugin.Plugin{
			plugin.ID: &plugin.ProviderPlugin{},
		},
		Cmd:              exec.Command(executable, args...), // #nosec G204 -- arguments are controlled by saturn-bot
		AllowedProtocols: []goPlugin.Protocol{goPlugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("get client of plugin: %w", err)
	}

	raw, err := rpcClient.Dispense(plugin.ID)
	if err != nil {
		return nil, fmt.Errorf("dispense tasks plugin: %w", err)
	}

	provider := raw.(plugin.Provider)
	getPluginResp, err := provider.GetPlugin(&proto.GetPluginRequest{Config: opts.config})
	if err != nil {
		return nil, fmt.Errorf("get plugin: %w", err)
	}

	err = calculateChecksum(opts.filePath, opts.hash)
	if err != nil {
		return nil, fmt.Errorf("calculate checksum of plugin: %w", err)
	}

	slog.Debug("Registered plugin", "name", getPluginResp.GetName())
	return &pluginWrapper{
		action:   &PluginAction{Provider: provider},
		client:   client,
		filter:   &PluginFilter{Provider: provider},
		provider: provider,
	}, nil
}

func (pw *pluginWrapper) onPrClosed(repo host.Repository) error {
	resp, err := pw.provider.OnPrClosed(&proto.OnPrClosedRequest{
		Context: &proto.Context{
			Repository: &proto.Repository{
				FullName:     repo.FullName(),
				CloneUrlHttp: repo.CloneUrlHttp(),
				CloneUrlSsh:  repo.CloneUrlSsh(),
				WebUrl:       repo.WebUrl(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("send 'PR Closed' event: %w", err)
	}

	if resp.GetError() != "" {
		return fmt.Errorf("'PR Closed' event failed: %s", resp.GetError())
	}

	return nil
}

func (pw *pluginWrapper) onPrCreated(repo host.Repository) error {
	resp, err := pw.provider.OnPrCreated(&proto.OnPrCreatedRequest{
		Context: &proto.Context{
			Repository: &proto.Repository{
				FullName:     repo.FullName(),
				CloneUrlHttp: repo.CloneUrlHttp(),
				CloneUrlSsh:  repo.CloneUrlSsh(),
				WebUrl:       repo.WebUrl(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("send 'PR Created' event: %w", err)
	}

	if resp.GetError() != "" {
		return fmt.Errorf("'PR Created' event failed: %s", resp.GetError())
	}

	return nil
}

func (pw *pluginWrapper) onPrMerged(repo host.Repository) error {
	resp, err := pw.provider.OnPrMerged(&proto.OnPrMergedRequest{
		Context: &proto.Context{
			Repository: &proto.Repository{
				FullName:     repo.FullName(),
				CloneUrlHttp: repo.CloneUrlHttp(),
				CloneUrlSsh:  repo.CloneUrlSsh(),
				WebUrl:       repo.WebUrl(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("send 'PR Merged' event: %w", err)
	}

	if resp.GetError() != "" {
		return fmt.Errorf("'PR Merged' event failed: %s", resp.GetError())
	}

	return nil
}

func (pw *pluginWrapper) stop() {
	// It is safe to call Kill() multiple times.
	pw.client.Kill()
}

// PluginAction wraps remote actions provided by a plugin.
type PluginAction struct {
	Provider plugin.Provider
}

// Apply implements action.Apply().
func (a *PluginAction) Apply(ctx context.Context) error {
	path := ctx.Value(gsContext.CheckoutPath{}).(string)
	pluginData := gsContext.PluginData(ctx)
	repo := ctx.Value(gsContext.RepositoryKey{}).(host.Repository)
	reply, err := a.Provider.ExecuteActions(&proto.ExecuteActionsRequest{
		Path: path,
		Context: &proto.Context{
			PluginData:  pluginData,
			PullRequest: newPullRequestPayload(ctx.Value(gsContext.PullRequestKey{})),
			Repository: &proto.Repository{
				FullName:     repo.FullName(),
				CloneUrlHttp: repo.CloneUrlHttp(),
				CloneUrlSsh:  repo.CloneUrlSsh(),
				WebUrl:       repo.WebUrl(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("execute actions of plugin: %w", err)
	}

	if reply.GetError() != "" {
		return fmt.Errorf("plugin failed to execute actions: %s", reply.GetError())
	}

	updatePluginData(pluginData, reply.PluginData)
	updateTemplateVars(ctx, reply.TemplateVars)
	return nil
}

// Apply implements action.String().
func (a *PluginAction) String() string {
	return "pluginAction"
}

// PluginFilter wraps remote filters provided by a plugin.
type PluginFilter struct {
	Provider plugin.Provider
}

// Do implements filter.Do().
func (f *PluginFilter) Do(ctx context.Context) (bool, error) {
	pluginData := gsContext.PluginData(ctx)
	repo := ctx.Value(gsContext.RepositoryKey{}).(host.Repository)
	reply, err := f.Provider.ExecuteFilters(&proto.ExecuteFiltersRequest{
		Context: &proto.Context{
			PluginData:  pluginData,
			PullRequest: newPullRequestPayload(ctx.Value(gsContext.PullRequestKey{})),
			Repository: &proto.Repository{
				FullName:     repo.FullName(),
				CloneUrlHttp: repo.CloneUrlHttp(),
				CloneUrlSsh:  repo.CloneUrlSsh(),
				WebUrl:       repo.WebUrl(),
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("execute filters of plugin: %w", err)
	}

	if reply.GetError() != "" {
		return false, fmt.Errorf("plugin failed to execute filters: %s", reply.GetError())
	}

	updatePluginData(pluginData, reply.PluginData)
	updateTemplateVars(ctx, reply.TemplateVars)
	return reply.GetMatch(), nil
}

// Name implements filter.Name().
func (f *PluginFilter) Name() string {
	return "pluginFilter"
}

// String implements filter.String().
func (f *PluginFilter) String() string {
	return "pluginFilter"
}

func newPullRequestPayload(value any) *proto.PullRequest {
	pr, ok := value.(host.PullRequest)
	if !ok {
		return nil
	}

	return &proto.PullRequest{
		Number: pr.Number,
		WebUrl: pr.WebURL,
	}
}

func updatePluginData(current, received map[string]string) {
	for k, v := range received {
		current[k] = v
	}
}

func updateTemplateVars(ctx context.Context, templateVarsReply map[string]string) {
	templateVars, ok := ctx.Value(gsContext.TemplateVarsKey{}).(map[string]string)
	if !ok {
		templateVars = make(map[string]string)
	}

	for k, v := range templateVarsReply {
		_, exists := templateVars[k]
		// Plugin should not update existing keys to avoid confusion.
		if !exists {
			templateVars[k] = v
		}
	}
}
