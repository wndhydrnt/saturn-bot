package task

import (
	"context"
	"fmt"
	"hash"
	"log/slog"
	"os/exec"

	goPlugin "github.com/hashicorp/go-plugin"
	"github.com/wndhydrnt/saturn-sync-go/plugin"
	proto "github.com/wndhydrnt/saturn-sync-go/protocol/v1"
	gsContext "github.com/wndhydrnt/saturn-sync/pkg/context"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
	gsLog "github.com/wndhydrnt/saturn-sync/pkg/log"
)

type startPluginOptions struct {
	args         []string
	hash         hash.Hash
	customConfig []byte
	executable   string
	filePath     string
}

type pluginWrapper struct {
	action   *PluginAction
	client   *goPlugin.Client
	filter   *PluginFilter
	provider plugin.Provider
}

func newPluginWrapper(opts startPluginOptions) (*pluginWrapper, error) {
	client := goPlugin.NewClient(&goPlugin.ClientConfig{
		HandshakeConfig: plugin.Handshake,
		Logger:          gsLog.DefaultHclogAdapter(),
		Plugins: map[string]goPlugin.Plugin{
			"tasks": &plugin.ProviderPlugin{},
		},
		Cmd:              exec.Command(opts.executable, opts.args...), // #nosec G204 -- arguments are controlled by saturn-sync
		AllowedProtocols: []goPlugin.Protocol{goPlugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("get client of plugin: %w", err)
	}

	raw, err := rpcClient.Dispense("tasks")
	if err != nil {
		return nil, fmt.Errorf("dispense tasks plugin: %w", err)
	}

	provider := raw.(plugin.Provider)
	getPluginResp, err := provider.GetPlugin(&proto.GetPluginRequest{CustomConfig: opts.customConfig})
	if err != nil {
		return nil, fmt.Errorf("get plugin: %w", err)
	}

	err = calculateChecksum(opts.filePath, opts.hash)
	if err != nil {
		return nil, fmt.Errorf("calculate checksum of plugin: %w", err)
	}

	slog.Debug("Registered plugin", "name", getPluginResp.GetName())
	return &pluginWrapper{
		action:   &PluginAction{provider: provider},
		filter:   &PluginFilter{provider: provider},
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
	provider plugin.Provider
}

// Apply implements action.Apply().
func (a *PluginAction) Apply(ctx context.Context) error {
	path := ctx.Value(gsContext.CheckoutPath{}).(string)
	repo := ctx.Value(gsContext.RepositoryKey{}).(host.Repository)
	reply, err := a.provider.ExecuteActions(&proto.ExecuteActionsRequest{
		Path: path,
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
		return fmt.Errorf("execute actions of plugin: %w", err)
	}

	if reply.GetError() != "" {
		return fmt.Errorf("plugin failed to execute actions: %s", reply.GetError())
	}

	return nil
}

// Apply implements action.String().
func (a *PluginAction) String() string {
	return "pluginAction"
}

// PluginFilter wraps remote filters provided by a plugin.
type PluginFilter struct {
	provider plugin.Provider
}

// Do implements filter.Do().
func (f *PluginFilter) Do(ctx context.Context) (bool, error) {
	repo := ctx.Value(gsContext.RepositoryKey{}).(host.Repository)
	reply, err := f.provider.ExecuteFilters(&proto.ExecuteFiltersRequest{
		Context: &proto.Context{
			Repository: &proto.Repository{
				FullName:     repo.FullName(),
				CloneUrlHttp: repo.CloneUrlHttp(),
				CloneUrlSsh:  repo.CloneUrlHttp(),
				WebUrl:       repo.BaseBranch(),
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("execute filters of plugin: %w", err)
	}

	if reply.GetError() != "" {
		return false, fmt.Errorf("plugin failed to execute filters: %s", reply.GetError())
	}

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
