package dnsmasq

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	Group   = "dnsmasq.kvaps.cf"
	Version = "v1beta1"
)

var (
	DnsHostsGVK    = schema.GroupVersionKind{Group: Group, Version: Version, Kind: "DnsHosts"}
	DhcpHostsGVK   = schema.GroupVersionKind{Group: Group, Version: Version, Kind: "DhcpHosts"}
	DhcpOptionsGVK = schema.GroupVersionKind{Group: Group, Version: Version, Kind: "DhcpOptions"}
)

const (
	FreeRadiusGroup   = "freeradius.valdron.ca"
	FreeRadiusVersion = "v1"
)

var FreeRadiusDeviceGVK = schema.GroupVersionKind{Group: FreeRadiusGroup, Version: FreeRadiusVersion, Kind: "Device"}
