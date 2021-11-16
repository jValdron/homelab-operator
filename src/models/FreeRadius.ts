import * as k8s from '@kubernetes/client-node';

export class DeviceSpec {
  macAddresses: string[];
  vlan: number;
}

export class DeviceStatus {
}

export class Device implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: DeviceSpec;
  status?: DeviceStatus;

  constructor(name: string) {
    this.apiVersion = 'freeradius.valdron.ca/v1';
    this.kind = 'Device';
    this.metadata = new k8s.V1ObjectMeta();
    this.metadata.name = name;
    this.spec = new DeviceSpec();
  }
}

export class DeviceList implements k8s.KubernetesListObject<Device> {
  apiVersion: string;
  items: Device[];
  kind: string;
  metadata: k8s.V1ObjectMeta;
}
