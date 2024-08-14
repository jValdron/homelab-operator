import * as k8s from '@kubernetes/client-node';

export class HttpRouteSpec {
  hostnames: string[];
  parentRefs: HttpRouteParentRef[];
}

export class HttpRouteParentRef {
  group: string;
  kind: string;
  name: string;
  namespace: string;
}

export class HttpRouteStatus {
}

export class HttpRoute implements k8s.KubernetesObject {
  apiVersion: string;
  kind: string;
  metadata: k8s.V1ObjectMeta;
  spec?: HttpRouteSpec;
  status?: HttpRouteStatus;
}

export class HttpRouteList implements k8s.KubernetesListObject<HttpRoute> {
  apiVersion: string;
  items: HttpRoute[];
  kind: string;
  metadata: k8s.V1ObjectMeta;
}
