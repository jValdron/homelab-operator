import * as k8s from '@kubernetes/client-node';

// v1beta.dnsmasq.kvaps.cf/DhcpHosts
// @see https://github.com/kvaps/dnsmasq-controller/blob/7dd2859be514465ab004466aca9ca99962daca99/api/v1beta1/dhcphosts_types.go
export class DhcpHostsSpecHost {
  macs?: string[];
  clientId?: string;
  setTags?: string[];
  tags?: string[];
  ip?: string;
  hostname?: string;
  leaseTime?: string = 'infinite';
  ignore?: boolean;
}

export class DhcpHostsSpec {
  controller?: string;
  hosts?: DhcpHostsSpecHost[]
}

export class DhcpHostsStatus {
}

export class DhcpHosts implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: DhcpHostsSpec;
  status?: DhcpHostsStatus;

  constructor(name: string) {
    this.apiVersion = 'dnsmasq.kvaps.cf/v1beta1';
    this.kind = 'DhcpHosts';
    this.metadata = new k8s.V1ObjectMeta();
    this.metadata.name = name;
  }
}

// v1beta.dnsmasq.kvaps.cf/DhcpOptions
// @see https://github.com/kvaps/dnsmasq-controller/blob/7dd2859be514465ab004466aca9ca99962daca99/api/v1beta1/dhcpoptions_types.go
export class DhcpOptionsSpecDhcpOption {
  key: string;
  values: string[];
  tags?: string[];
  encap?: string;
  viEncap?: string;
  vendor?: string;

  constructor(key: string, values: string[], tag?: string) {
    this.key = key;
    this.values = values;
    this.tags = tag ? [tag] : undefined;
  }
}

export class DhcpOptionsSpec {
  controller?: string;
  options?: DhcpOptionsSpecDhcpOption[]
}

export class DhcpOptionsStatus {
}

export class DhcpOptions implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: DhcpOptionsSpec;
  status?: DhcpOptionsStatus;

  constructor(name: string) {
    this.apiVersion = 'dnsmasq.kvaps.cf/v1beta1';
    this.kind = 'DhcpOptions';
    this.metadata = new k8s.V1ObjectMeta();
    this.metadata.name = name;
  }
}

// v1beta.dnsmasq.kvaps.cf/DnsHosts
// @see https://github.com/kvaps/dnsmasq-controller/blob/7dd2859be514465ab004466aca9ca99962daca99/api/v1beta1/dnshosts_types.go
export class DnsHostsSpecHost {
  ip: string;
  hostnames: string[];
}

export class DnsHostsSpec {
  controller?: string;
  hosts?: DnsHostsSpecHost[];
}

export class DnsHostsStatus {
}

export class DnsHosts implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: DnsHostsSpec;
  status?: DnsHostsStatus;

  constructor(name: string) {
    this.apiVersion = 'dnsmasq.kvaps.cf/v1beta1';
    this.kind = 'DnsHosts';
    this.metadata = new k8s.V1ObjectMeta();
    this.metadata.name = name;
  }
}
