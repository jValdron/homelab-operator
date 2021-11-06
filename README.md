# homelab-operator
Provides a way to track devices and networks within a homelab as Kubernetes resources.

Automatically uses devices, ingresses and services (LoadBalancers) to create DNS and DHCP resources to be used by [dnsmasq-controller](https://github.com/kvaps/dnsmasq-controller).

Complete setup provides DHCP, DNS and RADIUS authentication to homelab.

## Options
Available options via environment variables (or dotenv):
* `LOG_LEVEL`: Optional, overrides the log level: ['trace', 'debug', 'info', 'warn', 'error']
* `DNSMASQ_RESOURCES_NAMESPACE`: Namespace to create the dnsmasq resources within
* `ENABLE_INGRESS_BASED_DNS_HOSTS`: Whether or not to enable the _Ingress based DNS hosts_ controller (see Controllers).
* `ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS`: Whether or not to enable the _Load balancer based DNS hosts_ controller (see Controllers).
* `LOAD_BALANCER_BASED_SUFFIX`: Optionally, a suffix to add to the DNS hosts created as part of the load balancer based DNS hosts.

## Controllers
### Ingress based DNS hosts
TODO

### Load balancer based DNS hosts
TODO

### Devices
TODO

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
* `ownerRef` implementation to reconcile when owned resources are modified