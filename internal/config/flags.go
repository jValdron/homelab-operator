package config

import (
	"flag"
	"strings"
)

// CSVFlags holds comma-separated settings registered separately from Config fields.
type CSVFlags struct {
	DnsmasqDomainSearch string
	WatchNamespaces     string
}

// BindFlags registers homelab operator settings on fs. Defaults come from cfg (typically env via Load).
func BindFlags(fs *flag.FlagSet, cfg *Config) CSVFlags {
	cf := CSVFlags{
		DnsmasqDomainSearch: strings.Join(cfg.DnsmasqDomainSearch, ","),
		WatchNamespaces:     strings.Join(cfg.WatchNamespaces, ","),
	}

	fs.BoolVar(&cfg.DryRun, "dry-run", cfg.DryRun,
		"Compute desired state but do not write dnsmasq, FreeRADIUS, or status (env DRY_RUN)")
	fs.StringVar(&cfg.DnsmasqResourcesNamespace, "dnsmasq-resources-namespace", cfg.DnsmasqResourcesNamespace,
		"Namespace for dnsmasq CRs (env DNSMASQ_RESOURCES_NAMESPACE)")
	fs.StringVar(&cfg.DnsmasqDNSServer, "dnsmasq-dns-server", cfg.DnsmasqDNSServer,
		"Default DHCP DNS server option (env DNSMASQ_DNS_SERVER)")
	fs.StringVar(&cfg.DnsmasqDomainName, "dnsmasq-domain-name", cfg.DnsmasqDomainName,
		"DHCP domain-name option (env DNSMASQ_DOMAIN_NAME)")
	fs.StringVar(&cf.DnsmasqDomainSearch, "dnsmasq-domain-search", cf.DnsmasqDomainSearch,
		"Comma-separated DHCP domain-search values (env DNSMASQ_DOMAIN_SEARCH)")
	fs.BoolVar(&cfg.EnableDeviceBasedHosts, "enable-device-based-hosts", cfg.EnableDeviceBasedHosts,
		"Reconcile Device/Network into dnsmasq (env ENABLE_DEVICE_BASED_HOSTS)")
	fs.BoolVar(&cfg.EnableDeviceShortName, "enable-device-short-name", cfg.EnableDeviceShortName,
		"Use device name as DNS hostname (env ENABLE_DEVICE_SHORT_NAME)")
	fs.StringVar(&cfg.DeviceSuffix, "device-suffix", cfg.DeviceSuffix,
		"Extra hostname suffix per device (env DEVICE_SUFFIX)")
	fs.BoolVar(&cfg.EnableFreeRadiusDevices, "enable-freeradius-devices", cfg.EnableFreeRadiusDevices,
		"Sync FreeRADIUS Device CRs (env ENABLE_FREERADIUS_DEVICES)")
	fs.StringVar(&cfg.FreeRadiusDevicesNamespace, "freeradius-devices-namespace", cfg.FreeRadiusDevicesNamespace,
		"Target namespace for FreeRADIUS devices (env FREERADIUS_DEVICES_NAMESPACE)")
	fs.BoolVar(&cfg.FreeRadiusDevicesWirelessOnly, "freeradius-devices-wireless-only", cfg.FreeRadiusDevicesWirelessOnly,
		"Only wireless devices for FreeRADIUS (env FREERADIUS_DEVICES_WIRELESS_ONLY)")
	fs.BoolVar(&cfg.EnableIngressBasedDNSHosts, "enable-ingress-based-dns-hosts", cfg.EnableIngressBasedDNSHosts,
		"Ingress → dnshosts/ingress-based (env ENABLE_INGRESS_BASED_DNS_HOSTS)")
	fs.BoolVar(&cfg.UseHTTPRoutesForIngress, "use-http-routes-for-ingress", cfg.UseHTTPRoutesForIngress,
		"Include Gateway API HTTPRoutes (env USE_HTTP_ROUTES_FOR_INGRESS)")
	fs.BoolVar(&cfg.EnableLoadBalancerBasedDNSHosts, "enable-load-balancer-based-dns-hosts", cfg.EnableLoadBalancerBasedDNSHosts,
		"LoadBalancer Services → dnshosts/load-balancer-based (env ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS)")
	fs.StringVar(&cfg.LoadBalancerBasedSuffix, "load-balancer-based-suffix", cfg.LoadBalancerBasedSuffix,
		"Suffix on load balancer DNS names (env LOAD_BALANCER_BASED_SUFFIX)")
	fs.StringVar(&cf.WatchNamespaces, "homelab-watch-namespaces", cf.WatchNamespaces,
		"Comma-separated namespaces to watch; empty = all (env HOMELAB_WATCH_NAMESPACES)")

	return cf
}

// ApplyCSVFlags copies comma-separated flag values into cfg and applies derived defaults.
func ApplyCSVFlags(cfg *Config, cf CSVFlags) {
	cfg.DnsmasqDomainSearch = splitCommaList(cf.DnsmasqDomainSearch)
	cfg.WatchNamespaces = splitCommaList(cf.WatchNamespaces)
	finalize(cfg)
}

func splitCommaList(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func finalize(cfg *Config) {
	if cfg.EnableFreeRadiusDevices && cfg.FreeRadiusDevicesNamespace == "" {
		cfg.FreeRadiusDevicesNamespace = cfg.DnsmasqResourcesNamespace
	}
	deriveDomainSearch(cfg)
}
