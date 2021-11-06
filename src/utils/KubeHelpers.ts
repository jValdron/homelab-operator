import * as http from 'http';
import * as k8s from '@kubernetes/client-node';

export class CustomResourceDefinition {
  group: string;
  versions: k8s.V1CustomResourceDefinitionVersion[];
  plural: string;
}

export class KubeHelpers {
  /**
   * Upserts a custom resource to the given namespace.
   * @param client An instance of CustomObjectsApi.
   * @param crd The custom resource definition of the object to upsert.
   * @param namespace The namespace to operate within.
   * @param object The custom resource.
   * @returns A promise; resolves when upserted.
   */
  static async upsertNamespacedCustomResource<T extends k8s.KubernetesObject>(client: k8s.CustomObjectsApi, crd: CustomResourceDefinition, namespace: string, object: T): Promise<{
    response: http.IncomingMessage;
    body: object;
  }> {
    return new Promise((resolve, reject) => {
      client.getNamespacedCustomObject(
        crd.group,
        crd.versions[0].name,
        namespace,
        crd.plural,
        object.metadata.name
      ).then((existingObject) => {
        object.metadata.resourceVersion = (<T>existingObject.body).metadata.resourceVersion;

        client.replaceNamespacedCustomObject(
          crd.group,
          crd.versions[0].name,
          namespace,
          crd.plural,
          object.metadata.name,
          object
        ).then(resolve).catch(reject);
      }).catch(() => {
        client.createNamespacedCustomObject(
          crd.group,
          crd.versions[0].name,
          namespace,
          crd.plural,
          object
        ).then(resolve).catch(reject);
      });
    });
  }
}

