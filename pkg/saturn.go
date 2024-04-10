package pkg

import (
	"fmt"

	"github.com/wndhydrnt/saturn-sync/pkg/config"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
)

func createHostsFromConfig(cfg config.Config) ([]host.Host, error) {
	var hosts []host.Host
	if cfg.GitLabToken != "" {
		gitlab, err := host.NewGitLabHost(cfg.GitLabToken)
		if err != nil {
			return nil, fmt.Errorf("create gitlab host: %w", err)
		}

		hosts = append(hosts, gitlab)
	}

	if cfg.GitHubToken != "" {
		github, err := host.NewGitHubHost(cfg.GitHubAddress, cfg.GitHubToken, cfg.GitHubCacheDisabled)
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
