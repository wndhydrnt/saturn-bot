package pkg

import (
	"fmt"

	"github.com/wndhydrnt/saturn-sync/pkg/config"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
)

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
