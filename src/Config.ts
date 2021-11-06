export default {
  LogLevel: process.env.LOG_LEVEL ?? 'info',

  DnsmasqResourcesNamespace: process.env.DNSMASQ_RESOURCES_NAMESPACE ?? 'default',
  DnsmasqDomainName: process.env.DNSMASQ_DOMAIN_NAME,
  DnsmasqDomainSearch: process.env.DNSMASQ_DOMAIN_SEARCH ? process.env.DNSMASQ_DOMAIN_SEARCH.split(',') : null,

  EnableDeviceBasedHosts: process.env.ENABLE_DEVICE_BASED_HOSTS ? !(process.env.ENABLE_DEVICE_BASED_HOSTS === 'false' || process.env.ENABLE_DEVICE_BASED_HOSTS === '0') : true,
  EnableDeviceShortName: process.env.ENABLE_DEVICE_SHORT_NAME ? !(process.env.ENABLE_DEVICE_SHORT_NAME === 'false' || process.env.ENABLE_DEVICE_SHORT_NAME === '0') : true,
  DeviceSuffix: process.env.DEVICE_SUFFIX,

  EnableIngressBasedDnsHosts: process.env.ENABLE_INGRESS_BASED_DNS_HOSTS ? !(process.env.ENABLE_INGRESS_BASED_DNS_HOSTS === 'false' || process.env.ENABLE_INGRESS_BASED_DNS_HOSTS === '0') : true,

  EnableLoadBalancerBasedDnsHosts: process.env.ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS ? !(process.env.ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS === 'false' || process.env.ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS === '0') : true,
  LoadBalancerBasedSuffix: process.env.LOAD_BALANCER_BASED_SUFFIX ?? '',
}