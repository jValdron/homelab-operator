import * as Path from 'path';

import * as k8s from '@kubernetes/client-node';
import Operator from '@dot-i/k8s-operator';

import config from '../Config';
import { logger } from '../utils/Logger';
import { KubeHelpers } from '../utils/KubeHelpers';

import * as dnsmasq from '../models/Dnsmasq';

export default class LoadBalancerBasedDnsHosts extends Operator {
  private customObjectsClient: k8s.CustomObjectsApi;
  private coreClient: k8s.CoreV1Api;

  protected suffix: string;
  protected crds: {
    [key: string]: {
      group: string,
      versions: k8s.V1CustomResourceDefinitionVersion[],
      plural: string
    }
  };

  private currentReconciliation: Promise<void> = null;

  constructor(suffix: string) {
    super(logger);
    this.suffix = suffix;
  }

  protected async init(): Promise<void> {
    const kc = new k8s.KubeConfig();
    kc.loadFromDefault();

    this.customObjectsClient = kc.makeApiClient(k8s.CustomObjectsApi);
    this.coreClient = kc.makeApiClient(k8s.CoreV1Api);

    logger.trace({}, 'Setting up load balancer based DNS hosts watcher')

    const crdsBasePath = Path.resolve(__dirname, '..', '..', 'deploy', 'crds');

    this.crds = {
      dnsHost: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'dnsmasq', 'dns-hosts.yaml'))
    };

    await this.watchResource('', 'v1', 'services', async (e) => {
      logger.debug({ e }, 'Service resource was modified');
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

    logger.debug({}, 'Reconciliating load balancer services; will create DnsHosts based off services');

    this.currentReconciliation = new Promise((resolve, reject) => {
      this.coreClient.listServiceForAllNamespaces().then((services) => {
        const dnsHosts = new dnsmasq.DnsHosts('load-balancer-based');
        dnsHosts.spec = new dnsmasq.DnsHostsSpec();
        dnsHosts.spec.hosts = [];

        const hostsByIp: { [key: string]: dnsmasq.DnsHostsSpecHost } = {};

        services.body.items.filter((service) => {
          return service.spec.type === 'LoadBalancer';
        }).forEach((service) => {
          logger.trace({ service }, 'Found service');
          
          let dnsHost = hostsByIp[service.spec.loadBalancerIP];

          if (!hostsByIp[service.spec.loadBalancerIP])
          {
            dnsHost = new dnsmasq.DnsHostsSpecHost();
            dnsHost.ip = service.spec.loadBalancerIP;
            dnsHost.hostnames = [];
            hostsByIp[service.spec.loadBalancerIP] = dnsHost;
            dnsHosts.spec.hosts.push(dnsHost);
          }

          dnsHost.hostnames.push(service.metadata.name + this.suffix);
        });

        logger.info({ dnsHosts }, 'Generated load balancer based DnsHost; updating matching resource');

        KubeHelpers.upsertNamespacedCustomResource<dnsmasq.DnsHosts>(
          this.customObjectsClient,
          this.crds.dnsHost,
          config.DnsmasqResourcesNamespace,
          dnsHosts
        ).then((response) => {
          logger.debug({
            response: response,
            dnsHosts: dnsHosts
          }, 'Successfully replaced load balancer based DnsHost custom resource');

          return resolve();
        }).catch((err) => {
          logger.error({
            error: err,
            dnsHosts: dnsHosts
          }, 'Failed to replace load balancer based DnsHost custom resource');

          return reject(err);
        }).finally(() => {
          this.currentReconciliation = null;
        });
      });
    });

    return this.currentReconciliation;
  }
}
