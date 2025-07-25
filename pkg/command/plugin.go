package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	protoV1 "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/plugin"
)

var (
	defaultContext = &protoV1.Context{
		Repository: &protoV1.Repository{
			FullName:     "git.localhost/plugin/debug",
			CloneUrlHttp: "https://git.localhost/plugin/debug.git",
			CloneUrlSsh:  "git@git.localhost/plugin/debug.git",
			WebUrl:       "https://git.localhost/plugin/debug",
		},
		RunData: make(map[string]string),
	}
	PluginFuncs = map[string]func(ExecPluginOptions, *plugin.Plugin) (any, error){
		"apply":       applyPlugin,
		"filter":      filterPlugin,
		"onPrClosed":  onPrClosedPlugin,
		"onPrCreated": onPrCreatedPlugin,
		"onPrMerged":  onPrMergedPlugin,
		"shutdown":    shutdown,
	}
)

func applyPlugin(opts ExecPluginOptions, p *plugin.Plugin) (any, error) {
	return p.ExecuteActions(&protoV1.ExecuteActionsRequest{
		Context: opts.Context,
		Path:    opts.WorkDir,
	})
}

func filterPlugin(opts ExecPluginOptions, p *plugin.Plugin) (any, error) {
	return p.ExecuteFilters(&protoV1.ExecuteFiltersRequest{
		Context: opts.Context,
	})
}

func onPrClosedPlugin(opts ExecPluginOptions, p *plugin.Plugin) (any, error) {
	return p.OnPrClosed(&protoV1.OnPrClosedRequest{
		Context: opts.Context,
	})
}

func onPrCreatedPlugin(opts ExecPluginOptions, p *plugin.Plugin) (any, error) {
	return p.OnPrCreated(&protoV1.OnPrCreatedRequest{
		Context: opts.Context,
	})
}

func onPrMergedPlugin(opts ExecPluginOptions, p *plugin.Plugin) (any, error) {
	return p.OnPrMerged(&protoV1.OnPrMergedRequest{
		Context: opts.Context,
	})
}

func shutdown(_ ExecPluginOptions, p *plugin.Plugin) (any, error) {
	return p.Shutdown(&protoV1.ShutdownRequest{})
}

// ExecPluginOptions defines all options expected by function ExecPlugin.
type ExecPluginOptions struct {
	Addr      string
	Config    map[string]string
	Context   *protoV1.Context
	LogFormat string
	LogLevel  string
	Out       io.Writer
	Path      string
	WorkDir   string
}

// ExecPlugin executes the function of a plugin specified by `funcName`.
// It creates a temporary directory if `opts` doesn't specify one.
// It decodes the context content to send to the plugin from JSON.
func ExecPlugin(funcName string, opts ExecPluginOptions) error {
	f, ok := PluginFuncs[funcName]
	if !ok {
		return fmt.Errorf("unknown plugin function %s", funcName)
	}

	log.InitLog(
		config.ConfigurationLogFormat(opts.LogFormat),
		config.ConfigurationLogLevel(opts.LogLevel),
		config.ConfigurationGitLogLevelError,
	)

	if opts.Context == nil {
		opts.Context = defaultContext
	}

	if opts.WorkDir == "" {
		var err error
		opts.WorkDir, err = os.MkdirTemp("", "")
		if err != nil {
			return fmt.Errorf("create temporary working directory: %w", err)
		}
	}

	startOpts := plugin.StartOptions{
		Config: opts.Config,
	}
	if opts.Addr != "" {
		startOpts.Address = opts.Addr
	} else {
		startOpts.Exec = plugin.GetPluginExec(opts.Path, "", "")
	}

	p := &plugin.Plugin{}
	err := p.Start(startOpts)
	if err != nil {
		return fmt.Errorf("start plugin: %w", err)
	}

	// Only stop if plugin was started by saturn-bot
	if opts.Addr == "" {
		defer p.Stop()
	}

	reply, err := f(opts, p)
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	enc := json.NewEncoder(opts.Out)
	enc.SetIndent("", "  ")
	err = enc.Encode(map[string]any{
		"error": errStr,
		"reply": reply,
	})
	if err != nil {
		return fmt.Errorf("encode reply to JSON: %w", err)
	}

	return nil
}
