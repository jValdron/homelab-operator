package dnsmasq

import (
	"cmp"
	"net/netip"
	"slices"
)

type DnsHostEntry struct {
	IP        string   `json:"ip"`
	Hostnames []string `json:"hostnames,omitempty"`
}

type DnsHostsSpec struct {
	Controller string         `json:"controller,omitempty"`
	Hosts      []DnsHostEntry `json:"hosts,omitempty"`
}

// NormalizeDnsHosts sorts host entries by IP and hostnames within each entry for stable specs.
func NormalizeDnsHosts(spec DnsHostsSpec) DnsHostsSpec {
	for i := range spec.Hosts {
		slices.Sort(spec.Hosts[i].Hostnames)
		spec.Hosts[i].Hostnames = slices.Compact(spec.Hosts[i].Hostnames)
	}
	slices.SortFunc(spec.Hosts, func(a, b DnsHostEntry) int { return compareIP(a.IP, b.IP) })
	return spec
}

func compareIP(a, b string) int {
	pa, errA := netip.ParseAddr(a)
	pb, errB := netip.ParseAddr(b)
	if errA == nil && errB == nil {
		return pa.Compare(pb)
	}
	return cmp.Compare(a, b)
}

const DhcpLeaseTimeInfinite = "infinite"

type DhcpHostEntry struct {
	MACs      []string `json:"macs,omitempty"`
	IP        string   `json:"ip,omitempty"`
	Hostname  string   `json:"hostname,omitempty"`
	SetTags   []string `json:"setTags,omitempty"`
	LeaseTime string   `json:"leaseTime,omitempty"`
}

type DhcpHostsSpec struct {
	Controller string          `json:"controller,omitempty"`
	Hosts      []DhcpHostEntry `json:"hosts,omitempty"`
}

type DhcpOption struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
	Tags   []string `json:"tags,omitempty"`
}

type DhcpOptionsSpec struct {
	Controller string       `json:"controller,omitempty"`
	Options    []DhcpOption `json:"options,omitempty"`
}
