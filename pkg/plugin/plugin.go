package plugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	goPlugin "github.com/hashicorp/go-plugin"
	"github.com/wndhydrnt/saturn-bot-go/plugin"
	protoV1 "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
)

type StartOptions struct {
	Address     string
	Config      map[string]string
	Exec        ExecOptions
	OnMsgStderr func(pluginName string, msg string)
	OnMsgStdout func(pluginName string, msg string)
}

type ExecOptions struct {
	Args       []string
	Executable string
}

type Plugin struct {
	Name     string
	Provider plugin.Provider

	client        *goPlugin.Client
	stderrAdapter *stdioAdapter
	stdoutAdapter *stdioAdapter
}

// Apply implements action.Action
func (p *Plugin) Apply(ctx context.Context) error {
	path := ctx.Value(sbcontext.CheckoutPath{}).(string)
	runData := sbcontext.RunData(ctx)
	reply, err := p.ExecuteActions(&protoV1.ExecuteActionsRequest{
		Path:    path,
		Context: NewContext(ctx),
	})
	if err != nil {
		return fmt.Errorf("execute action: %w", err)
	}

	updateRunData(runData, reply.RunData)
	return nil
}

// Do implements filter.Filter.
func (p *Plugin) Do(ctx context.Context) (bool, error) {
	runData := sbcontext.RunData(ctx)
	reply, err := p.ExecuteFilters(&protoV1.ExecuteFiltersRequest{
		Context: NewContext(ctx),
	})
	if err != nil {
		return false, fmt.Errorf("execute filter: %w", err)
	}

	updateRunData(runData, reply.RunData)
	return reply.GetMatch(), nil
}

func (p *Plugin) ExecuteActions(req *protoV1.ExecuteActionsRequest) (*protoV1.ExecuteActionsResponse, error) {
	reply, err := p.Provider.ExecuteActions(req)
	if err != nil {
		return nil, fmt.Errorf("rpc call ExecuteActions: %w", err)
	}

	if reply.GetError() != "" {
		return nil, errors.New(reply.GetError())
	}

	return reply, nil
}

func (p *Plugin) ExecuteFilters(req *protoV1.ExecuteFiltersRequest) (*protoV1.ExecuteFiltersResponse, error) {
	reply, err := p.Provider.ExecuteFilters(req)
	if err != nil {
		return nil, fmt.Errorf("rpc call ExecuteFilters: %w", err)
	}

	if reply.GetError() != "" {
		return nil, errors.New(reply.GetError())
	}

	return reply, nil
}

func (p *Plugin) OnPrClosed(req *protoV1.OnPrClosedRequest) (*protoV1.OnPrClosedResponse, error) {
	reply, err := p.Provider.OnPrClosed(req)
	if err != nil {
		return nil, fmt.Errorf("rpc call OnPrClosed: %w", err)
	}

	if reply.GetError() != "" {
		return nil, fmt.Errorf("on pull request closed: %s", reply.GetError())
	}

	return reply, nil
}

func (p *Plugin) OnPrCreated(req *protoV1.OnPrCreatedRequest) (*protoV1.OnPrCreatedResponse, error) {
	reply, err := p.Provider.OnPrCreated(req)
	if err != nil {
		return nil, fmt.Errorf("rpc call OnPrCreated: %w", err)
	}

	if reply.GetError() != "" {
		return nil, fmt.Errorf("on pull request created: %s", reply.GetError())
	}

	return reply, nil
}

func (p *Plugin) OnPrMerged(req *protoV1.OnPrMergedRequest) (*protoV1.OnPrMergedResponse, error) {
	reply, err := p.Provider.OnPrMerged(req)
	if err != nil {
		return nil, fmt.Errorf("rpc call OnPrMerged: %w", err)
	}

	if reply.GetError() != "" {
		return nil, fmt.Errorf("on pull request merged: %s", reply.GetError())
	}

	return reply, nil
}

func (p *Plugin) Start(opts StartOptions) error {
	if p.Provider == nil {
		err := p.init(opts)
		if err != nil {
			return err
		}
	}

	getPluginResp, err := p.Provider.GetPlugin(&protoV1.GetPluginRequest{Config: opts.Config})
	if err != nil {
		return fmt.Errorf("get plugin: %w", err)
	}

	log.Log().Debugf("Registered plugin %s", getPluginResp.GetName())
	p.Name = getPluginResp.GetName()
	if p.stderrAdapter != nil {
		p.stderrAdapter.name = getPluginResp.GetName()
	}
	if p.stdoutAdapter != nil {
		p.stdoutAdapter.name = getPluginResp.GetName()
	}
	return nil
}

func (p *Plugin) Stop() {
	if p.client != nil {
		// It is safe to call Kill() multiple times.
		p.client.Kill()
	}
}

// String implements action.Action
// String implements filter.Filter
func (p *Plugin) String() string {
	return fmt.Sprintf("plugin(name=%s)", p.Name)
}

func (p *Plugin) init(opts StartOptions) error {
	p.stderrAdapter = &stdioAdapter{onMessage: opts.OnMsgStderr}
	p.stdoutAdapter = &stdioAdapter{onMessage: opts.OnMsgStdout}
	clientCfg := &goPlugin.ClientConfig{
		HandshakeConfig: plugin.Handshake,
		Logger:          log.DefaultHclogAdapter(),
		Plugins: map[string]goPlugin.Plugin{
			plugin.ID: &plugin.ProviderPlugin{},
		},
		AllowedProtocols: []goPlugin.Protocol{goPlugin.ProtocolGRPC},
		SyncStdout:       p.stdoutAdapter,
		SyncStderr:       p.stderrAdapter,
	}
	if opts.Address == "" {
		clientCfg.Cmd = exec.Command(opts.Exec.Executable, opts.Exec.Args...) // #nosec G204 -- arguments are controlled by saturn-bot
	} else {
		reattachCfg, err := parseAddr(opts.Address)
		if err != nil {
			return fmt.Errorf("parse plugin address: %w", err)
		}

		clientCfg.Reattach = reattachCfg
	}

	p.client = goPlugin.NewClient(clientCfg)

	rpcClient, err := p.client.Client()
	if err != nil {
		return fmt.Errorf("get client of plugin: %w", err)
	}

	raw, err := rpcClient.Dispense(plugin.ID)
	if err != nil {
		return fmt.Errorf("dispense plugin: %w", err)
	}

	p.Provider = raw.(plugin.Provider)
	return nil
}

func GetPluginExec(path, javaExec, pythonExec string) ExecOptions {
	var opts ExecOptions
	ext := filepath.Ext(path)
	switch ext {
	case ".jar":
		opts.Executable = javaExec
		if opts.Executable == "" {
			opts.Executable = "java"
		}
		opts.Args = append(opts.Args, "-jar", path)
	case ".py":
		opts.Executable = pythonExec
		if opts.Executable == "" {
			opts.Executable = "python"
		}
		opts.Args = append(opts.Args, path)
	default:
		opts.Executable = path
	}

	return opts
}

func NewContext(ctx context.Context) *protoV1.Context {
	repo := ctx.Value(sbcontext.RepositoryKey{}).(host.Repository)
	pluginCtx := &protoV1.Context{
		Repository: &protoV1.Repository{
			FullName:     repo.FullName(),
			CloneUrlHttp: repo.CloneUrlHttp(),
			CloneUrlSsh:  repo.CloneUrlSsh(),
			WebUrl:       repo.WebUrl(),
		},
		RunData: sbcontext.RunData(ctx),
	}

	pr, ok := ctx.Value(sbcontext.PullRequestKey{}).(host.PullRequest)
	if ok {
		pluginCtx.PullRequest = &protoV1.PullRequest{
			Number: pr.Number,
			WebUrl: pr.WebURL,
		}
	}

	return pluginCtx
}

func updateRunData(current, received map[string]string) {
	for k, v := range received {
		current[k] = v
	}
}

// parseAddr parses a go-plugin connection string.
// The format is: <go-plugin protocol version>|<plugin protocol version>|<net protocol>|<address>|<encoding>
// Example: 1|1|tcp|127.0.0.1:11049|grpc
// See https://github.com/hashicorp/go-plugin/blob/06afb6d7ae99fca0002c4eecfd7b973073982672/docs/internals.md
func parseAddr(addr string) (*goPlugin.ReattachConfig, error) {
	parts := strings.Split(addr, "|")
	protocolVersion, _ := strconv.Atoi(parts[1])
	var address net.Addr
	switch parts[2] {
	case "tcp":
		var err error
		address, err = net.ResolveTCPAddr("tcp", parts[3])
		if err != nil {
			return nil, fmt.Errorf("resolve tcp address: %w", err)
		}
	case "unix":
		var err error
		address, err = net.ResolveUnixAddr("unix", parts[3])
		if err != nil {
			return nil, fmt.Errorf("resolve unix address: %w", err)
		}
	}
	cfg := &goPlugin.ReattachConfig{
		Protocol:        goPlugin.Protocol(parts[4]),
		ProtocolVersion: protocolVersion,
		Addr:            address,
	}
	return cfg, nil
}
