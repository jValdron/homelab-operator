package dnsmasq

import "testing"

const intHostA = "a.int.example"

func TestNormalizeDnsHosts(t *testing.T) {
	spec := DnsHostsSpec{Hosts: []DnsHostEntry{
		{IP: "10.1.21.31", Hostnames: []string{"z.int.example", intHostA, intHostA}},
		{IP: "10.1.21.30", Hostnames: []string{"b.example", "a.example"}},
	}}
	got := NormalizeDnsHosts(spec)
	if len(got.Hosts) != 2 || got.Hosts[0].IP != "10.1.21.30" || got.Hosts[1].IP != "10.1.21.31" {
		t.Fatalf("host order: %+v", got.Hosts)
	}
	want30 := []string{"a.example", "b.example"}
	if len(got.Hosts[0].Hostnames) != len(want30) {
		t.Fatalf("hostnames .30: %v", got.Hosts[0].Hostnames)
	}
	for i := range want30 {
		if got.Hosts[0].Hostnames[i] != want30[i] {
			t.Fatalf("hostnames .30: got %v want %v", got.Hosts[0].Hostnames, want30)
		}
	}
	if len(got.Hosts[1].Hostnames) != 2 || got.Hosts[1].Hostnames[0] != intHostA {
		t.Fatalf("hostnames .31: %v", got.Hosts[1].Hostnames)
	}
}

func TestNormalizeDnsHostsIPOrder(t *testing.T) {
	spec := DnsHostsSpec{Hosts: []DnsHostEntry{
		{IP: "10.1.21.41", Hostnames: []string{"a"}},
		{IP: "10.1.21.5", Hostnames: []string{"b"}},
		{IP: "10.1.21.15", Hostnames: []string{"c"}},
	}}
	got := NormalizeDnsHosts(spec)
	wantIPs := []string{"10.1.21.5", "10.1.21.15", "10.1.21.41"}
	for i, ip := range wantIPs {
		if got.Hosts[i].IP != ip {
			t.Fatalf("index %d: got IP %q want %q (all: %v)", i, got.Hosts[i].IP, ip, hostsIPs(got))
		}
	}
}

func hostsIPs(spec DnsHostsSpec) []string {
	out := make([]string, len(spec.Hosts))
	for i, h := range spec.Hosts {
		out[i] = h.IP
	}
	return out
}
