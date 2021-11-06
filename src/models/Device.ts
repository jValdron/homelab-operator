import * as k8s from '@kubernetes/client-node';

export class DeviceSpec {
  network?: string;
  ip?: string;
  macAddresses?: string[];
  dhcp: boolean = false;

  trunked: boolean = false;
  wired: boolean = false;
  wireless: boolean = false;
  usb: boolean = false; 

  location?: string;
  position?: number;
  pdu?: string;
  switch?: string;
  side?: RackSide;

  manufacturer?: string;
  model?: string;
  serialNumber?: string;
}

export class DeviceStatus {
}

export class Device implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: DeviceSpec;
  status?: DeviceStatus;
}

export class DeviceList implements k8s.KubernetesListObject<Device> {
  apiVersion: string;
  items: Device[];
  kind: string;
  metadata: k8s.V1ObjectMeta;
}

export enum RackSide {
  Front,
  Back
}
