import * as k8s from '@kubernetes/client-node';

export class IngressRouterSpec {
  ingressClass?: string;
  ipAddress: string;
}

export class IngressRouterStatus {
}

export class IngressRouter implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: IngressRouterSpec;
  status?: IngressRouterStatus;
}