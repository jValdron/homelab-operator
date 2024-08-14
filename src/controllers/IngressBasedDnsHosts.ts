import * as Path from 'path';

import * as k8s from '@kubernetes/client-node';
import Operator, { ResourceEventType } from '@dot-i/k8s-operator';

import config from '../Config';
import { logger } from '../utils/Logger';
import { KubeHelpers } from '../utils/KubeHelpers';

import { IngressRouter } from '../models/IngressRouter';
import * as dnsmasq from '../models/Dnsmasq';
import * as gatewayApi from '../models/GatewayApi';

export default class IngressBasedDnsHosts extends Operator {
  private customObjectsClient: k8s.CustomObjectsApi;
  private networkingClient: k8s.NetworkingV1Api;

  protected ingressRouters: { [key: string]: IngressRouter } = {};
  protected crds: {
    [key: string]: {
      group: string,
      versions: k8s.V1CustomResourceDefinitionVersion[],
      plural: string
    }
  };

  private currentReconciliation: Promise<void> = null;

  constructor() {
    super(logger);
  }

  protected async init(): Promise<void> {
    const kc = new k8s.KubeConfig();
    kc.loadFromDefault();

    this.customObjectsClient = kc.makeApiClient(k8s.CustomObjectsApi);
    this.networkingClient = kc.makeApiClient(k8s.NetworkingV1Api);

    logger.trace({}, 'Setting up ingress based DNS hosts watcher')

    const crdsBasePath = Path.resolve(__dirname, '..', '..', 'deploy', 'crds');

    this.crds = {
      dnsHost: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'dnsmasq', 'dns-hosts.yaml')),
      httpRoute: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'gataway-api', 'http-routes.yaml')),
      ingressRouter: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'ingress-router.yaml')),
    };

    await this.watchResource(this.crds.ingressRouter.group, this.crds.ingressRouter.versions[0].name, this.crds.ingressRouter.plural, async (e) => {
      const ir = <IngressRouter>e.object;

      if (e.type == ResourceEventType.Added || ResourceEventType.Modified)
      {
        logger.debug({ e }, 'Ingress router updated');
        this.ingressRouters[ir.metadata.name] = ir;
      }
      else
      {
        logger.debug({ e }, 'Ingress router deleted');
        delete this.ingressRouters[ir.metadata.name];
      }

      await this.reconcile();
    });

    await this.watchResource('networking.k8s.io', 'v1', 'ingresses', async (e) => {
      logger.debug({ e }, 'Ingress resource was modified');
      await this.reconcile();
    });

    if (config.UseHttpRoutesForIngress)
    {
      await this.watchResource('gateway.networking.k8s.io', 'v1', 'httproutes', async (e) => {
        logger.debug({ e }, 'HTTP route was modified');
        await this.reconcile();
      });
    }
  }

  protected async reconcile(): Promise<void> {
    if (this.currentReconciliation)
    {
      logger.trace({
        promise: this.currentReconciliation
      }, 'Reconciliation already in progress; skipping reconciliation');
      return this.currentReconciliation;
    }

    if (!Object.keys(this.ingressRouters).length)
    {
      logger.debug({
        ingressRouters: this.ingressRouters
      }, 'No ingress routers; not reconciliating ingress based routers');
      return Promise.resolve();
    }

    logger.debug({}, 'Reconciliating ingresses; will create DnsHosts based off ingresses');

    this.currentReconciliation = new Promise(async (resolve, reject) => {
      const ingressBasedDnsHosts = new dnsmasq.DnsHosts('ingress-based');
      ingressBasedDnsHosts.metadata.namespace = config.DnsmasqResourcesNamespace;
      ingressBasedDnsHosts.spec = new dnsmasq.DnsHostsSpec();
      ingressBasedDnsHosts.spec.hosts = [];

      const dnsHosts: { [key: string]: dnsmasq.DnsHostsSpecHost } = {};

      const pushIngressRouterIfNotExists = (ingressRouterName, ipAddress) => {
        if (!dnsHosts[ingressRouterName])
        {
          logger.info({
            name: ingressRouterName
          }, 'Created new DNS host for matching ingress router');

          const dnsHost = new dnsmasq.DnsHostsSpecHost();
          dnsHost.ip = ipAddress;
          dnsHost.hostnames = [];

          dnsHosts[ingressRouterName] = dnsHost;
          ingressBasedDnsHosts.spec.hosts.push(dnsHost);
        }

        return dnsHosts[ingressRouterName];
      }

      const ingresses = await this.networkingClient.listIngressForAllNamespaces();
      ingresses.body.items.forEach((ingress) => {
        logger.trace({ ingress }, 'Found ingress');

        const ingressClass = ingress.spec.ingressClassName || ingress.metadata.annotations['kubernetes.io/ingress.class'];
        const matchingIngressRouters = Object.values(this.ingressRouters).filter((ir) => {
          return ir.spec.ingressClass == ingressClass;
        });

        if (matchingIngressRouters.length)
        {
          logger.debug({
            ingress: ingress,
            ingressRouter: matchingIngressRouters[0]
          }, 'Found matching ingress router; processing ingress');

          const ingressRouterName = matchingIngressRouters[0].metadata.name;
          const dnsHost = pushIngressRouterIfNotExists(ingressRouterName, matchingIngressRouters[0].spec.ipAddress)

          ingress.spec.rules.forEach((rule) => {
            logger.trace({
              host: rule.host,
              ingress: ingress,
              dnsHost: dnsHost
            }, 'Adding ingress hostname to DNS host');

            dnsHost.hostnames.push(rule.host);
          });
        }
        else
        {
          logger.debug({ ingress }, 'No matching ingress router; not handling ingress');
        }
      });

      if (config.UseHttpRoutesForIngress)
      {
        const httpRoutes = await this.customObjectsClient.listClusterCustomObject(
          this.crds.httpRoute.group,
          this.crds.httpRoute.versions[0].name,
          this.crds.httpRoute.plural
        );

        (<gatewayApi.HttpRouteList>httpRoutes.body).items.forEach((httpRoute) => {
          logger.trace({ httpRoute }, 'Found HTTP route');

          const gateway = httpRoute.spec.parentRefs[0].name;
          const matchingIngressRouters = Object.values(this.ingressRouters).filter((ir) => {
            return ir.spec.ingressClass == gateway;
          });

          if (matchingIngressRouters.length)
          {
            logger.debug({
              httpRoute: httpRoute,
              ingressRouter: matchingIngressRouters[0]
            }, 'Found matching ingress router; processing HTTP route');

            const ingressRouterName = matchingIngressRouters[0].metadata.name;
            const dnsHost = pushIngressRouterIfNotExists(ingressRouterName, matchingIngressRouters[0].spec.ipAddress)

            httpRoute.spec.hostnames.forEach((hostname) => {
              logger.trace({
                hostname: hostname,
                httpRoute: httpRoute,
                dnsHost: dnsHost
              }, 'Adding HTTP route hostname to DNS host');

              dnsHost.hostnames.push(hostname);
            });
          }
          else
          {
            logger.debug({ httpRoute }, 'No matching ingress router (gateway); not handling HTTP route');
          }
        });
      }

      logger.info({ ingressBasedDnsHosts }, 'Generated ingress-based DnsHost; updating matching resource');

      KubeHelpers.upsertNamespacedCustomResource<dnsmasq.DnsHosts>(
        this.customObjectsClient,
        this.crds.dnsHost,
        config.DnsmasqResourcesNamespace,
        ingressBasedDnsHosts
      ).then((response) => {
        logger.debug({
          response: response,
          dnsHosts: dnsHosts
        }, 'Successfully replaced ingress based DnsHost custom resource');

        return resolve();
      }).catch((err) => {
        logger.error({
          error: err,
          dnsHosts: dnsHosts
        }, 'Failed to replace ingress based DnsHost custom resource');

        return reject(err);
      }).finally(() => {
        this.currentReconciliation = null;
      });
    });

    return this.currentReconciliation;
  }
}
