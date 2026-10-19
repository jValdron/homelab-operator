package desired

import (
	"testing"

	homelabv1 "github.com/jValdron/homelab-operator/api/v1"
	"github.com/jValdron/homelab-operator/internal/config"
	"github.com/jValdron/homelab-operator/internal/dnsmasq"
)

const (
	testNetworkName = "net"
	testRouterIP    = "10.0.0.1"
)

func TestBuildDeviceBasedDeterministic(t *testing.T) {
	devices := []homelabv1.Device{{
		Name: "b",
		Spec: homelabv1.DeviceSpec{
			Network: testNetworkName, IP: "10.0.0.2", DHCP: true,
			MACAddresses: []string{"aa:bb:cc:dd:ee:01"},
		},
	}}
	devices2 := []homelabv1.Device{{
		Name: "a",
		Spec: homelabv1.DeviceSpec{
			Network: testNetworkName, IP: testRouterIP, DHCP: true,
			MACAddresses: []string{"aa:bb:cc:dd:ee:02"},
		},
	}}

	networks := []homelabv1.Network{{
		Name: testNetworkName,
		Spec: homelabv1.NetworkSpec{
			VLAN: 10, CIDRBlock: "10.0.0.0/24", ServeDHCP: true, Router: testRouterIP,
		},
	}}

	cfg := config.Load()
	cfg.EnableDeviceShortName = true

	out1 := BuildDeviceBased([]homelabv1.Device{devices[0], devices2[0]}, networks, cfg)
	out2 := BuildDeviceBased([]homelabv1.Device{devices2[0], devices[0]}, networks, cfg)

	if len(out1.DnsHosts) != len(out2.DnsHosts) {
		t.Fatalf("dns host map size differs")
	}
	d1 := out1.DnsHosts[DeviceBasedNetworkName(testNetworkName)]
	d2 := out2.DnsHosts[DeviceBasedNetworkName(testNetworkName)]
	if len(d1.Hosts) != 2 || len(d2.Hosts) != 2 {
		t.Fatalf("expected 2 dns hosts, got %d and %d", len(d1.Hosts), len(d2.Hosts))
	}
	if d1.Hosts[0].IP != testRouterIP || d1.Hosts[1].IP != "10.0.0.2" {
		t.Fatalf("expected sorted device order by name, got IPs %s and %s", d1.Hosts[0].IP, d1.Hosts[1].IP)
	}
}

func TestBuildDeviceBasedDomainSearch(t *testing.T) {
	cfg := config.Config{
		DnsmasqDomainName:       "valdron.ca",
		LoadBalancerBasedSuffix: ".int.valdron.ca",
		DeviceSuffix:            ".dev.valdron.ca",
	}
	config.ApplyCSVFlags(&cfg, config.CSVFlags{}) // runs finalize → deriveDomainSearch

	out := BuildDeviceBased(nil, nil, cfg)
	var search []string
	for _, o := range out.DhcpOptions.Options {
		if o.Key == "option:domain-search" {
			search = o.Values
		}
	}
	want := []string{"valdron.ca", "int.valdron.ca", "dev.valdron.ca"}
	if len(search) != len(want) {
		t.Fatalf("domain-search values: got %v want %v", search, want)
	}
	for i := range want {
		if search[i] != want[i] {
			t.Fatalf("domain-search values: got %v want %v", search, want)
		}
	}
}

func TestBuildDeviceBasedDhcpLeaseTime(t *testing.T) {
	devices := []homelabv1.Device{{
		Name: "bpc0",
		Spec: homelabv1.DeviceSpec{
			Network: testNetworkName, IP: "10.1.30.30", DHCP: true,
			MACAddresses: []string{"52:54:00:fe:c5:3f"},
		},
	}}
	networks := []homelabv1.Network{{
		Name: testNetworkName,
		Spec: homelabv1.NetworkSpec{
			VLAN: 1, CIDRBlock: "10.1.30.0/24", ServeDHCP: true, Router: "10.1.30.1",
		},
	}}
	out := BuildDeviceBased(devices, networks, config.Load())
	dhcp := out.DhcpHosts[DeviceBasedNetworkName(testNetworkName)]
	if len(dhcp.Hosts) != 1 || dhcp.Hosts[0].LeaseTime != dnsmasq.DhcpLeaseTimeInfinite {
		t.Fatalf("expected leaseTime infinite, got %+v", dhcp.Hosts)
	}
}

func TestBuildDeviceBasedRouterOption(t *testing.T) {
	networks := []homelabv1.Network{{
		Name: "devices",
		Spec: homelabv1.NetworkSpec{
			VLAN: 1, CIDRBlock: "10.1.0.0/24", ServeDHCP: true, Router: "10.1.0.1",
		},
	}}

	out := BuildDeviceBased(nil, networks, config.Load())
	found := false
	for _, o := range out.DhcpOptions.Options {
		if o.Key == "option:router" && len(o.Values) == 1 && o.Values[0] == "10.1.0.1" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected router dhcp option")
	}
}
