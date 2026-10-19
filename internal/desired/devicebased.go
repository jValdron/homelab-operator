package desired

import (
	"cmp"
	"fmt"
	"slices"

	homelabv1 "github.com/jValdron/homelab-operator/api/v1"
	"github.com/jValdron/homelab-operator/internal/config"
	"github.com/jValdron/homelab-operator/internal/dnsmasq"
)

const DhcpOptionsName = "device-based-hosts"

func DeviceBasedNetworkName(network string) string {
	return fmt.Sprintf("device-based-%s", network)
}

type DeviceBasedOutput struct {
	DhcpOptions dnsmasq.DhcpOptionsSpec
	DhcpHosts   map[string]dnsmasq.DhcpHostsSpec
	DnsHosts    map[string]dnsmasq.DnsHostsSpec
	FreeRadius  map[string]map[string]any // name -> spec map
}

func BuildDeviceBased(devices []homelabv1.Device, networks []homelabv1.Network, cfg config.Config) DeviceBasedOutput {
	slices.SortFunc(devices, func(a, b homelabv1.Device) int { return cmp.Compare(a.Name, b.Name) })
	slices.SortFunc(networks, func(a, b homelabv1.Network) int { return cmp.Compare(a.Name, b.Name) })

	netByName := map[string]homelabv1.Network{}
	for _, n := range networks {
		netByName[n.Name] = n
	}

	out := DeviceBasedOutput{
		DhcpHosts:  map[string]dnsmasq.DhcpHostsSpec{},
		DnsHosts:   map[string]dnsmasq.DnsHostsSpec{},
		FreeRadius: map[string]map[string]any{},
	}

	opts := dnsmasq.DhcpOptionsSpec{Options: []dnsmasq.DhcpOption{}}
	if cfg.DnsmasqDomainName != "" {
		opts.Options = append(opts.Options, dnsmasq.DhcpOption{
			Key: "option:domain-name", Values: []string{cfg.DnsmasqDomainName},
		})
	}
	if len(cfg.DnsmasqDomainSearch) > 0 {
		opts.Options = append(opts.Options, dnsmasq.DhcpOption{
			Key: "option:domain-search", Values: cfg.DnsmasqDomainSearch,
		})
	}

	for _, network := range networks {
		if !network.Spec.ServeDHCP || network.Spec.Router == "" {
			continue
		}
		tag := network.Name
		if network.Spec.DNSServer != "" {
			opts.Options = append(opts.Options, dnsmasq.DhcpOption{
				Key: "option:dns-server", Values: []string{network.Spec.DNSServer}, Tags: []string{tag},
			})
		} else if cfg.DnsmasqDNSServer != "" {
			opts.Options = append(opts.Options, dnsmasq.DhcpOption{
				Key: "option:dns-server", Values: []string{cfg.DnsmasqDNSServer}, Tags: []string{tag},
			})
		}
		opts.Options = append(opts.Options, dnsmasq.DhcpOption{
			Key: "option:router", Values: []string{network.Spec.Router}, Tags: []string{tag},
		})
	}
	out.DhcpOptions = opts

	for _, device := range devices {
		if device.Spec.Network == "" {
			continue
		}
		network, ok := netByName[device.Spec.Network]
		if !ok {
			continue
		}
		netName := device.Spec.Network
		resName := DeviceBasedNetworkName(netName)

		if device.Spec.IP != "" {
			dnsSpec, ok := out.DnsHosts[resName]
			if !ok {
				dnsSpec = dnsmasq.DnsHostsSpec{Hosts: []dnsmasq.DnsHostEntry{}}
			}
			dnsSpec.Hosts = append(dnsSpec.Hosts, dnsmasq.DnsHostEntry{
				IP: device.Spec.IP, Hostnames: hostnamesForDevice(device, cfg),
			})
			out.DnsHosts[resName] = dnsSpec
		}

		if device.Spec.DHCP {
			dhcp, ok := out.DhcpHosts[resName]
			if !ok {
				dhcp = dnsmasq.DhcpHostsSpec{Hosts: []dnsmasq.DhcpHostEntry{}}
			}
			dhcp.Hosts = append(dhcp.Hosts, dnsmasq.DhcpHostEntry{
				Hostname:  device.Name,
				IP:        device.Spec.IP,
				MACs:      append([]string(nil), device.Spec.MACAddresses...),
				SetTags:   []string{netName},
				LeaseTime: dnsmasq.DhcpLeaseTimeInfinite,
			})
			out.DhcpHosts[resName] = dhcp
		}

		if cfg.EnableFreeRadiusDevices {
			if cfg.FreeRadiusDevicesWirelessOnly && !device.Spec.Wireless {
				continue
			}
			out.FreeRadius[device.Name] = freeRadiusDeviceSpec(device.Spec.MACAddresses, network.Spec.VLAN)
		}
	}

	for name, dnsSpec := range out.DnsHosts {
		out.DnsHosts[name] = dnsmasq.NormalizeDnsHosts(dnsSpec)
	}

	return out
}

// freeRadiusDeviceSpec builds JSON-compatible spec values for unstructured/API writes.
func freeRadiusDeviceSpec(macAddresses []string, vlan int) map[string]any {
	macs := make([]any, len(macAddresses))
	for i, m := range macAddresses {
		macs[i] = m
	}
	return map[string]any{
		"macAddresses": macs,
		"vlan":         int64(vlan),
	}
}

func hostnamesForDevice(device homelabv1.Device, cfg config.Config) []string {
	var names []string
	add := func(h string) {
		if cfg.EnableDeviceShortName {
			names = append(names, h)
		}
		if cfg.DeviceSuffix != "" {
			names = append(names, h+cfg.DeviceSuffix)
		}
	}
	add(device.Name)
	for _, h := range device.Spec.AdditionalHostnames {
		add(h)
	}
	return names
}
