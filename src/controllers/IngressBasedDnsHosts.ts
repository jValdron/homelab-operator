import * as Path from 'path';

import * as k8s from '@kubernetes/client-node';
import Operator, { ResourceEventType } from '@dot-i/k8s-operator';

import config from '../Config';
import { logger } from '../utils/Logger';
import { KubeHelpers } from '../utils/KubeHelpers';

import { IngressRouter } from '../models/IngressRouter';
import * as dnsmasq from '../models/Dnsmasq';

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
      ingressRouter: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'ingress-router.yaml'))
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

    this.currentReconciliation = new Promise((resolve, reject) => {
      this.networkingClient.listIngressForAllNamespaces().then((ingresses) => {
        const ingressBasedDnsHosts = new dnsmasq.DnsHosts('ingress-based');
        ingressBasedDnsHosts.metadata.namespace = config.DnsmasqResourcesNamespace;
        ingressBasedDnsHosts.spec = new dnsmasq.DnsHostsSpec();
        ingressBasedDnsHosts.spec.hosts = [];

        const dnsHosts: { [key: string]: dnsmasq.DnsHostsSpecHost } = {};
      
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

            if (!dnsHosts[ingressRouterName])
            {
              logger.info({
                name: ingressRouterName
              }, 'Created new DNS host for matching ingress router');

              const dnsHost = new dnsmasq.DnsHostsSpecHost();
              dnsHost.ip = matchingIngressRouters[0].spec.ipAddress;
              dnsHost.hostnames = [];

              dnsHosts[ingressRouterName] = dnsHost;
              ingressBasedDnsHosts.spec.hosts.push(dnsHost);
            }

            const dnsHost = dnsHosts[ingressRouterName];

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
    });

    return this.currentReconciliation;
  }
}
