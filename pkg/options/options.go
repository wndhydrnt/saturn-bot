package options

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	sLog "github.com/wndhydrnt/saturn-bot/pkg/log"
)

type ActionFactories []action.Factory

func (af ActionFactories) Find(name string) action.Factory {
	for _, f := range af {
		if f.Name() == name {
			return f
		}
	}

	return nil
}

type FilterFactories []filter.Factory

func (ff FilterFactories) Find(name string) filter.Factory {
	for _, f := range ff {
		if f.Name() == name {
			return f
		}
	}

	return nil
}

type Opts struct {
	ActionFactories ActionFactories
	Config          config.Configuration
	FilterFactories FilterFactories
	Hosts           []host.Host
}

// ToOptions takes a configuration struct, initializes global state
// and returns an Options struct that can be modified further, if needed.
func ToOptions(c config.Configuration) (Opts, error) {
	opts := Opts{
		ActionFactories: action.BuiltInFactories,
		Config:          c,
		FilterFactories: filter.BuiltInFactories,
	}

	hosts, err := createHostsFromConfig(c)
	if err != nil {
		return opts, fmt.Errorf("create hosts from configuration: %w", err)
	}

	opts.Hosts = hosts
	return opts, nil
}

func createHostsFromConfig(cfg config.Configuration) ([]host.Host, error) {
	var hosts []host.Host
	if cfg.GitlabToken != nil {
		gitlab, err := host.NewGitLabHost(cfg.GitlabAddress, *cfg.GitlabToken)
		if err != nil {
			return nil, fmt.Errorf("create gitlab host: %w", err)
		}

		hosts = append(hosts, gitlab)
	}

	if cfg.GithubToken != nil {
		var addr string
		if cfg.GithubAddress != nil {
			addr = *cfg.GithubAddress
		}

		github, err := host.NewGitHubHost(addr, *cfg.GithubToken, cfg.GithubCacheDisabled)
		if err != nil {
			return nil, fmt.Errorf("create github host: %w", err)
		}

		hosts = append(hosts, github)
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no hosts configured")
	}

	return hosts, nil
}

// Initialize ensures that outside dependencies needed on every execution of saturn-bot are set up.
// Such dependencies can be logging or directories.
func Initialize(opts Opts) error {
	sLog.InitLog(opts.Config.LogFormat, opts.Config.LogLevel, opts.Config.GitLogLevel)

	if opts.Config.DataDir == nil {
		return fmt.Errorf("missing dataDir configuration setting")
	}

	info, err := os.Stat(*opts.Config.DataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Info("Creating data directory", "path", *opts.Config.DataDir)
			mkdirErr := os.MkdirAll(*opts.Config.DataDir, 0700)
			if mkdirErr != nil {
				return fmt.Errorf("create data directory: %w", err)
			}
		} else {
			return fmt.Errorf("check if data directory exists: %w", err)
		}
	}

	if info != nil && !info.IsDir() {
		return fmt.Errorf("data directory %s is not a directory", *opts.Config.DataDir)
	}

	return nil
}
