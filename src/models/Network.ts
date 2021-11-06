import * as k8s from '@kubernetes/client-node';

export class NetworkSpec {
  role?: string;

  vlan: number;
  cidrBlock: string;

  serveDhcp: boolean = false;
  router?: string;
}

export class NetworkStatus {
}

export class Network implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: NetworkSpec;
  status?: NetworkStatus;
}

export class NetworkList implements k8s.KubernetesListObject<Network> {
  apiVersion: string;
  items: Network[];
  kind: string;
  metadata: k8s.V1ObjectMeta;
}
