package config

import "testing"

const testDomain = "valdron.ca"

func TestDeriveDomainSearchFromSuffixes(t *testing.T) {
	cfg := Config{
		DnsmasqDomainName:       testDomain,
		LoadBalancerBasedSuffix: ".int." + testDomain,
		DeviceSuffix:            ".dev." + testDomain,
	}
	finalize(&cfg)
	want := []string{testDomain, "int." + testDomain, "dev." + testDomain}
	if len(cfg.DnsmasqDomainSearch) != len(want) {
		t.Fatalf("got %v want %v", cfg.DnsmasqDomainSearch, want)
	}
	for i := range want {
		if cfg.DnsmasqDomainSearch[i] != want[i] {
			t.Fatalf("got %v want %v", cfg.DnsmasqDomainSearch, want)
		}
	}
}

func TestDeriveDomainSearchExplicitWins(t *testing.T) {
	cfg := Config{
		DnsmasqDomainName:       testDomain,
		DnsmasqDomainSearch:     []string{"example.com"},
		LoadBalancerBasedSuffix: ".int." + testDomain,
		DeviceSuffix:            ".dev." + testDomain,
	}
	finalize(&cfg)
	if len(cfg.DnsmasqDomainSearch) != 1 || cfg.DnsmasqDomainSearch[0] != "example.com" {
		t.Fatalf("unexpected search: %v", cfg.DnsmasqDomainSearch)
	}
}
