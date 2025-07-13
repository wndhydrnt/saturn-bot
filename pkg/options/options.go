package options

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/metrics"
)

var (
	// ErrNoHosts indicates that no hosts have been set up.
	ErrNoHosts = errors.New("no hosts configured")
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
	// Clock interfaces to a clock.
	// Its purpose is to fake time in unit tests.
	// Defaults to an object that proxies to the [time] package.
	Clock                clock.Clock
	Config               config.Configuration
	DataDir              string
	FilterFactories      FilterFactories
	Hosts                []host.Host
	IsCi                 bool
	SkipPlugins          bool
	PushGateway          *push.Pusher
	PrometheusGatherer   prometheus.Gatherer
	PrometheusRegisterer prometheus.Registerer
	RepositoryCacheTtl   time.Duration
	// ServerShutdownCheckInterval is the interval at which the API server checks if all conditions
	// have been met before shutting down gracefully.
	// This option isn't exposed as a configuration item because it's used by tests only.
	ServerShutdownCheckInterval time.Duration
	// ServerShutdownTimeout is the maximum duration the API server waits before
	// it abandons a graceful shutdown and exits.
	ServerShutdownTimeout time.Duration
	WorkerLoopInterval    time.Duration
}

func (o *Opts) SetPrometheusRegistry(reg *prometheus.Registry) {
	o.PrometheusGatherer = reg
	o.PrometheusRegisterer = reg
}

// ToOptions takes a configuration struct, initializes global state
// and returns an Options struct that can be modified further, if needed.
func ToOptions(c config.Configuration) (Opts, error) {
	opts := Opts{
		ActionFactories:      action.BuiltInFactories,
		Config:               c,
		FilterFactories:      filter.BuiltInFactories,
		PrometheusGatherer:   prometheus.DefaultGatherer,
		PrometheusRegisterer: prometheus.DefaultRegisterer,
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
		return nil, ErrNoHosts
	}

	return hosts, nil
}

// Initialize ensures that outside dependencies needed on every execution of saturn-bot are set up.
// Such dependencies can be logging or directories.
func Initialize(opts *Opts) error {
	log.InitLog(opts.Config.LogFormat, opts.Config.LogLevel, opts.Config.GitLogLevel)

	var dataDir string
	if opts.Config.DataDir == nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get user home dir to set default data directory: %w", err)
		}

		dataDir = filepath.Join(homeDir, ".saturn-bot", "data")
	} else {
		dataDir = *opts.Config.DataDir
	}

	log.Log().Infof("Using data directory %s", dataDir)
	info, err := os.Stat(dataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Log().Infof("Creating data directory %s", dataDir)
			mkdirErr := os.MkdirAll(dataDir, 0700)
			if mkdirErr != nil {
				return fmt.Errorf("create data directory: %w", err)
			}
		} else {
			return fmt.Errorf("check if data directory exists: %w", err)
		}
	}

	if info != nil && !info.IsDir() {
		return fmt.Errorf("data directory %s is not a directory", dataDir)
	}

	opts.DataDir = dataDir

	loop, err := time.ParseDuration(opts.Config.WorkerLoopInterval)
	if err != nil {
		return fmt.Errorf("setting workerLoopInterval '%s' is not a Go duration: %w", opts.Config.WorkerLoopInterval, err)
	}
	opts.WorkerLoopInterval = loop

	repositoryCacheTtl, err := time.ParseDuration(opts.Config.RepositoryCacheTtl)
	if err != nil {
		return fmt.Errorf("setting repositoryCacheTtl '%s' is not a Go duration: %w", opts.Config.RepositoryCacheTtl, err)
	}
	opts.RepositoryCacheTtl = repositoryCacheTtl

	if opts.Config.PrometheusPushgatewayUrl != nil && opts.PrometheusRegisterer != nil && opts.PrometheusGatherer != nil {
		metrics.Register(opts.PrometheusRegisterer)
		opts.PushGateway = push.New(*opts.Config.PrometheusPushgatewayUrl, "saturn-bot").
			Gatherer(opts.PrometheusGatherer)
	}

	if opts.Clock == nil {
		opts.Clock = clock.Default
	}

	shutdownTimeout, err := time.ParseDuration(opts.Config.ServerShutdownTimeout)
	if err != nil {
		return fmt.Errorf("setting serverShutdownTimeout '%s' is not a Go duration: %w", opts.Config.ServerShutdownTimeout, err)
	}
	opts.ServerShutdownTimeout = shutdownTimeout

	return nil
}
