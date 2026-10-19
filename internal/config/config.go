package config

import (
	"os"
	"strings"
)

// Config holds operator settings (env parity with legacy Node operator).
type Config struct {
	DryRun bool

	DnsmasqResourcesNamespace string
	DnsmasqDNSServer          string
	DnsmasqDomainName         string
	DnsmasqDomainSearch       []string

	EnableDeviceBasedHosts        bool
	EnableDeviceShortName         bool
	DeviceSuffix                  string
	EnableFreeRadiusDevices       bool
	FreeRadiusDevicesNamespace    string
	FreeRadiusDevicesWirelessOnly bool

	EnableIngressBasedDNSHosts bool
	UseHTTPRoutesForIngress    bool

	EnableLoadBalancerBasedDNSHosts bool
	LoadBalancerBasedSuffix         string

	WatchNamespaces []string // empty = all namespaces
}

func envBool(key string, defaultVal bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return v != "false" && v != "0"
}

func Load() Config {
	watchNS := os.Getenv("HOMELAB_WATCH_NAMESPACES")
	var watch []string
	if watchNS != "" {
		for ns := range strings.SplitSeq(watchNS, ",") {
			if ns = strings.TrimSpace(ns); ns != "" {
				watch = append(watch, ns)
			}
		}
	}

	cfg := Config{
		DryRun: envBool("DRY_RUN", false),

		DnsmasqResourcesNamespace: getenv("DNSMASQ_RESOURCES_NAMESPACE", "default"),
		DnsmasqDNSServer:          os.Getenv("DNSMASQ_DNS_SERVER"),
		DnsmasqDomainName:         os.Getenv("DNSMASQ_DOMAIN_NAME"),
		DnsmasqDomainSearch:       splitCommaList(os.Getenv("DNSMASQ_DOMAIN_SEARCH")),

		EnableDeviceBasedHosts:        envBool("ENABLE_DEVICE_BASED_HOSTS", true),
		EnableDeviceShortName:         envBool("ENABLE_DEVICE_SHORT_NAME", true),
		DeviceSuffix:                  os.Getenv("DEVICE_SUFFIX"),
		EnableFreeRadiusDevices:       envBool("ENABLE_FREERADIUS_DEVICES", false),
		FreeRadiusDevicesNamespace:    os.Getenv("FREERADIUS_DEVICES_NAMESPACE"),
		FreeRadiusDevicesWirelessOnly: envBool("FREERADIUS_DEVICES_WIRELESS_ONLY", true),

		EnableIngressBasedDNSHosts:      envBool("ENABLE_INGRESS_BASED_DNS_HOSTS", true),
		UseHTTPRoutesForIngress:         envBool("USE_HTTP_ROUTES_FOR_INGRESS", true),
		EnableLoadBalancerBasedDNSHosts: envBool("ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS", true),
		LoadBalancerBasedSuffix:         getenv("LOAD_BALANCER_BASED_SUFFIX", ""),

		WatchNamespaces: watch,
	}

	finalize(&cfg)
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
