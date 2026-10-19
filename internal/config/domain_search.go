package config

import (
	"slices"
	"strings"
)

// deriveDomainSearch fills DnsmasqDomainSearch when unset, using domain-name plus DNS suffix flags.
// Matches typical homelab setups (e.g. valdron.ca + .int.valdron.ca + .dev.valdron.ca) without a
// separate DNSMASQ_DOMAIN_SEARCH value.
func deriveDomainSearch(cfg *Config) {
	if len(cfg.DnsmasqDomainSearch) > 0 || cfg.DnsmasqDomainName == "" {
		return
	}
	if cfg.DeviceSuffix == "" && cfg.LoadBalancerBasedSuffix == "" {
		return
	}

	var out []string
	add := func(raw string) {
		d := strings.TrimPrefix(strings.TrimSpace(raw), ".")
		if d == "" {
			return
		}
		if slices.Contains(out, d) {
			return
		}
		out = append(out, d)
	}
	add(cfg.DnsmasqDomainName)
	add(cfg.LoadBalancerBasedSuffix)
	add(cfg.DeviceSuffix)
	cfg.DnsmasqDomainSearch = out
}
