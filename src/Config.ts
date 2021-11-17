export default {
  LogLevel: process.env.LOG_LEVEL ?? 'info',
  DryRun: process.env.DRY_RUN ? !(process.env.DRY_RUN === 'false' || process.env.DRY_RUN === '0') : false,

  DnsmasqResourcesNamespace: process.env.DNSMASQ_RESOURCES_NAMESPACE ?? 'default',
  DnsmasqDnsServer: process.env.DNSMASQ_DNS_SERVER,
  DnsmasqDomainName: process.env.DNSMASQ_DOMAIN_NAME,
  DnsmasqDomainSearch: process.env.DNSMASQ_DOMAIN_SEARCH ? process.env.DNSMASQ_DOMAIN_SEARCH.split(',') : null,

  EnableDeviceBasedHosts: process.env.ENABLE_DEVICE_BASED_HOSTS ? !(process.env.ENABLE_DEVICE_BASED_HOSTS === 'false' || process.env.ENABLE_DEVICE_BASED_HOSTS === '0') : true,
  EnableDeviceShortName: process.env.ENABLE_DEVICE_SHORT_NAME ? !(process.env.ENABLE_DEVICE_SHORT_NAME === 'false' || process.env.ENABLE_DEVICE_SHORT_NAME === '0') : true,
  DeviceSuffix: process.env.DEVICE_SUFFIX,
  EnableFreeRadiusDevices: process.env.ENABLE_FREERADIUS_DEVICES ? !(process.env.ENABLE_FREERADIUS_DEVICES === 'false' || process.env.ENABLE_FREERADIUS_DEVICES === '0') : false,
  FreeRadiusDevicesNamespace: process.env.FREERADIUS_DEVICES_NAMESPACE,
  FreeRadiusDevicesWirelessOnly: process.env.FREERADIUS_DEVICES_WIRELESS_ONLY ? !(process.env.FREERADIUS_DEVICES_WIRELESS_ONLY === 'false' || process.env.FREERADIUS_DEVICES_WIRELESS_ONLY === '0') : true,

  EnableIngressBasedDnsHosts: process.env.ENABLE_INGRESS_BASED_DNS_HOSTS ? !(process.env.ENABLE_INGRESS_BASED_DNS_HOSTS === 'false' || process.env.ENABLE_INGRESS_BASED_DNS_HOSTS === '0') : true,

  EnableLoadBalancerBasedDnsHosts: process.env.ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS ? !(process.env.ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS === 'false' || process.env.ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS === '0') : true,
  LoadBalancerBasedSuffix: process.env.LOAD_BALANCER_BASED_SUFFIX ?? '',
}