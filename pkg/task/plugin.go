package task

import (
	"context"
	"fmt"
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
	customConfig []byte
	executable   string
	filePath     string
}

// startPlugin starts a go-plugin.
// The protocol to communicate with the plugin is GRPC.
// It requests all tasks that the plugin provides, initializes them and returns them.
func startPlugin(opts *startPluginOptions) ([]Task, error) {
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
	listTasksReply, err := provider.ListTasks(&proto.ListTasksRequest{CustomConfig: opts.customConfig})
	if err != nil {
		return nil, fmt.Errorf("list tasks of plugin: %w", err)
	}

	var result []Task
	checksum, err := calculateChecksum(opts.filePath)
	if err != nil {
		return nil, fmt.Errorf("calculate checksum: %w", err)
	}

	for _, task := range listTasksReply.Tasks {
		wrapper := &Wrapper{checksum: checksum}
		wrapper.Task = task
		wrapper.actions, err = createActionsForTask(wrapper.Task.Actions, opts.filePath)
		if err != nil {
			return nil, fmt.Errorf("parse actions of task '%s': %w", task.GetName(), err)
		}

		wrapper.actions = append(wrapper.actions, &PluginAction{provider: provider, taskName: task.GetName()})

		wrapper.filters, err = createFiltersForTask(wrapper.Task.Filters)
		if err != nil {
			return nil, fmt.Errorf("parse filters of task '%s': %w", task.GetName(), err)
		}

		wrapper.filters = append(wrapper.filters, &PluginFilter{provider: provider, taskName: task.GetName()})

		wrapper.onPrClosedFunc = func(repository host.Repository) error {
			reply, err := provider.OnPrClosed(&proto.OnPrClosedRequest{
				TaskName: task.GetName(),
				Context: &proto.Context{
					Repository: &proto.Repository{
						FullName:     repository.FullName(),
						CloneUrlHttp: repository.CloneUrlHttp(),
						CloneUrlSsh:  repository.CloneUrlSsh(),
						WebUrl:       repository.WebUrl(),
					},
				},
			})
			if err != nil {
				return fmt.Errorf("send 'PR Closed' event: %w", err)
			}

			if reply.GetError() != "" {
				return fmt.Errorf("'PR Closed' event failed: %s", reply.GetError())
			}

			return nil
		}

		wrapper.onPrCreatedFunc = func(repository host.Repository) error {
			reply, err := provider.OnPrCreated(&proto.OnPrCreatedRequest{
				TaskName: task.GetName(),
				Context: &proto.Context{
					Repository: &proto.Repository{
						FullName:     repository.FullName(),
						CloneUrlHttp: repository.CloneUrlHttp(),
						CloneUrlSsh:  repository.CloneUrlSsh(),
						WebUrl:       repository.WebUrl(),
					},
				},
			})
			if err != nil {
				return fmt.Errorf("send 'PR Created' event: %w", err)
			}

			if reply.GetError() != "" {
				return fmt.Errorf("'PR Created' event failed: %s", reply.GetError())
			}

			return nil
		}

		wrapper.onPrMergedFunc = func(repository host.Repository) error {
			reply, err := provider.OnPrMerged(&proto.OnPrMergedRequest{
				TaskName: task.GetName(),
				Context: &proto.Context{
					Repository: &proto.Repository{
						FullName:     repository.FullName(),
						CloneUrlHttp: repository.CloneUrlHttp(),
						CloneUrlSsh:  repository.CloneUrlSsh(),
						WebUrl:       repository.WebUrl(),
					},
				},
			})
			if err != nil {
				return fmt.Errorf("send 'PR Merged' event: %w", err)
			}

			if reply.GetError() != "" {
				return fmt.Errorf("'PR Merged' event failed: %s", reply.GetError())
			}

			return nil
		}

		wrapper.stopFunc = func() error {
			// It is safe to call Kill() multiple times.
			client.Kill()
			return nil
		}

		result = append(result, wrapper)
	}

	return result, nil
}

// PluginAction wraps remote actions provided by a plugin.
type PluginAction struct {
	provider plugin.Provider
	taskName string
}

// Apply implements action.Apply().
func (a *PluginAction) Apply(ctx context.Context) error {
	path := ctx.Value(gsContext.CheckoutPath{}).(string)
	repo := ctx.Value(gsContext.RepositoryKey{}).(host.Repository)
	reply, err := a.provider.ExecuteActions(&proto.ExecuteActionsRequest{
		TaskName: a.taskName,
		Path:     path,
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
		return fmt.Errorf("execute actions of plugin task: %w", err)
	}

	if reply.GetError() != "" {
		return fmt.Errorf("task plugin failed to execute actions: %s", reply.GetError())
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
	taskName string
}

// Do implements filter.Do().
func (f *PluginFilter) Do(ctx context.Context) (bool, error) {
	repo := ctx.Value(gsContext.RepositoryKey{}).(host.Repository)
	reply, err := f.provider.ExecuteFilters(&proto.ExecuteFiltersRequest{
		TaskName: f.taskName,
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
		return false, fmt.Errorf("execute filters of plugin task: %w", err)
	}

	if reply.GetError() != "" {
		return false, fmt.Errorf("task plugin failed to execute filters: %s", reply.GetError())
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
