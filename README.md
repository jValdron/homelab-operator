# homelab-operator
Provides a way to track devices and networks within a homelab as Kubernetes resources.

Automatically uses devices, ingresses and services (LoadBalancers) to create DNS and DHCP resources to be used by [dnsmasq-controller](https://github.com/kvaps/dnsmasq-controller).

Complete setup provides DHCP, DNS and RADIUS authentication (planned) to homelab.

## Notice
This comes as-is, with no support or no warranties.

## Background
I was tired of managing my deviecs through an Excel spreadsheet, just to have my various configuration being managed separately, and end up with drifts. After experimenting with [dnsmasq-controller](https://github.com/kvaps/dnsmasq-controller), and I needed RADIUS authentication configured for my devices too, I figured why not build a small operator to create the configuration I need and a way to track everything through code/custom resources.

## Options
Available options via environment variables (or dotenv):
* `LOG_LEVEL`: Optional, overrides the log level: ['trace', 'debug', 'info', 'warn', 'error']
* `DRY_RUN`: Whether or not to enable dry run mode, won't create any `dnsmasq` related resources.

* `DNSMASQ_RESOURCES_NAMESPACE`: Namespace to create the dnsmasq resources within
* `DNSMASQ_DOMAIN_NAME`: Adds a `domain-name` option under the created `DhcpOptions` resources.
* `DNSMASQ_DOMAIN_SEARCH`: Adds a `domain-seach` option under the created `DhcpOptions` resources, with one or more values.

* `ENABLE_DEVICE_BASED_HOSTS`: Whether or not to enable the _Devies based hosts_ controller (see Controllers).
* `ENABLE_DEVICE_SHORT_NAME`: Whether to enable the short name form of hostnames on the generated `DnsHosts` resources. Short name is simply the name of the device.
* `DEVICE_SUFFIX`: Optional, if provided, creates another hostname under `DnsHosts` resources using the device name and appending the given suffix.

* `ENABLE_INGRESS_BASED_DNS_HOSTS`: Whether or not to enable the _Ingress based DNS hosts_ controller (see Controllers).

* `ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS`: Whether or not to enable the _Load balancer based DNS hosts_ controller (see Controllers).
* `LOAD_BALANCER_BASED_SUFFIX`: Optionally, a suffix to add to the DNS hosts created as part of the load balancer based DNS hosts.

## Controllers
### Devices based hosts
The devices based hosts is the main controller which watches over `Devices` and `Networks` resources in order to create matching `DhcpOptions`, `DhcpHosts` and `DnsHosts`.

There will be one `DhcpHosts` and `DnsHosts` resource created per each `Networks` resource you create. Additionally, there will be a single `DhcpOptions` resource created: `dhcpoptions/device-based-hosts`

In order to generate `DhcpOptions` and `DhcpHosts`, you must create a `Networks` resource (with `serveDhcp`):
```
apiVersion: homelab.valdron.ca/v1
kind: Network
metadata:
  name: my-devices
spec:
  serveDhcp: true
  router: 10.1.0.1
```
_More options are available for `Networks`, see Resources._

This will create a matching `router` option under the `dhcpoptions/device-based-hosts` resource for the given device.

Simply need to create a device resource and a matching section under the given network's `DnsHosts` and `DhcpHosts` resource:
```
apiVersion: homelab.valdron.ca/v1
kind: Device
metadata:
  name: my-device
spec:
  network: my-devices
  ip: 10.1.0.4
  macAddresses:
    - 48:4d:7e:de:19:8f
  dhcp: true
```

### Ingress based DNS hosts
The Ingress based DNS hosts watches over all ingresses across all namespaces and adds a DNS host resource for their respective hostnames.

The generated resource will end up being: `dnshosts/ingress-based`

You must first create an `IngressRouter` resource for the ingress classes you wish to create host entries for:
```
apiVersion: homelab.valdron.ca/v1
kind: IngressRouter
metadata:
  name: internal
spec:
  ingressClass: internal
  canonicalName: router-internal.valdron.ca
```

This will create a host entry, under the `dnshosts/ingress-based` resource, for each ingresses with the `internal` ingress class.

### Load balancer based DNS hosts
Similar to the inress based DNS hosts but with with services with the `LoadBalancer` type.

The Ingress based DNS hosts watches over all services across all namespaces, with the `LoadBalancer` type, and adds a DNS host resource for their respective hostnames.

The generated resource will end up being: `dnshosts/load-balancer-based`

Optionally, you can specify `LOAD_BALANCER_BASED_SUFFIX` to add a suffix to all host entries that will be created.

## Resources
### Devices
```
apiVersion: homelab.valdron.ca/v1
kind: Device
metadata:
  name: jpc0
spec:
  network: devices
  ip: 10.1.30.2
  macAddresses:
    - 48:4d:7e:de:19:8f
  dhcp: true

  additionalHostnames:
    - jason-pc0
    - my-other-hostname

  trunked: false
  wired: true
  wireless: false
  usb: false 

  location: Rack
  position: 16
  pdu: '1/3'
  switch: '1/1/7'
  side: Front

  manufacturer: Dell
  model: Precision T5810
  serialNumber: 6CGHXG2
```

### Network
```
apiVersion: homelab.valdron.ca/v1
kind: Network
metadata:
  name: devices
spec:
  role: Devices

  vlan: 2030
  cidrBlock: 10.1.30.0/24

  serveDhcp: true
  router: 10.1.30.1
```

### Ingress Router
```
apiVersion: homelab.valdron.ca/v1
kind: IngressRouter
metadata:
  name: internal
spec:
  ingressClass: internal
  canonicalName: router-internal.valdron.ca
```

## TODO
* Add FreeRADIUS integration
* `ownerRef` implementation to reconcile when owned resources are modified
* Add leader election