package pkg

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	sLog "github.com/wndhydrnt/saturn-bot/pkg/log"
)

// initialize ensures that outside dependencies needed on every execution of saturn-bot are set up.
// Such dependencies can be logging or directories.
func initialize(cfg config.Configuration) error {
	sLog.InitLog(cfg.LogFormat, cfg.LogLevel, cfg.GitLogLevel)

	if cfg.DataDir == nil {
		return fmt.Errorf("missing dataDir configuration setting")
	}

	info, err := os.Stat(*cfg.DataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Info("Creating data directory", "path", *cfg.DataDir)
			mkdirErr := os.MkdirAll(*cfg.DataDir, 0700)
			if mkdirErr != nil {
				return fmt.Errorf("create data directory: %w", err)
			}
		} else {
			return fmt.Errorf("check if data directory exists: %w", err)
		}
	}

	if info != nil && !info.IsDir() {
		return fmt.Errorf("data directory %s is not a directory", *cfg.DataDir)
	}

	return nil
}

func createHostsFromConfig(cfg config.Configuration) ([]host.Host, error) {
	var hosts []host.Host
	if cfg.GitlabToken != nil {
		gitlab, err := host.NewGitLabHost(*cfg.GitlabToken)
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
