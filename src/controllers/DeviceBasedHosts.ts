import * as Path from 'path';

import * as k8s from '@kubernetes/client-node';
import Operator from '@dot-i/k8s-operator';

import config from '../Config';
import { logger } from '../utils/Logger';
import { KubeHelpers } from '../utils/KubeHelpers';

import * as dnsmasq from '../models/Dnsmasq';
import { Device, DeviceList } from '../models/Device';
import { Network, NetworkList } from '../models/Network';

export default class DeviceBasedHosts extends Operator {
  private customObjectsClient: k8s.CustomObjectsApi;

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

    logger.trace({}, 'Setting up devices watcher');

    const crdsBasePath = Path.resolve(__dirname, '..', '..', 'deploy', 'crds');

    this.crds = {
      device: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'device.yaml')),
      dnsHosts: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'dnsmasq', 'dns-hosts.yaml')),
      dhcpHosts: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'dnsmasq', 'dhcp-hosts.yaml')),
      dhcpOptions: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'dnsmasq', 'dhcp-options.yaml')),
      network: await this.registerCustomResourceDefinition(Path.resolve(crdsBasePath, 'network.yaml'))
    };

    await this.watchResource(this.crds.device.group, this.crds.device.versions[0].name, this.crds.device.plural, async (e) => {
      logger.debug({ e }, 'Device resource was modified');
      await this.reconcile(e.meta.namespace);
    });

    await this.watchResource(this.crds.network.group, this.crds.network.versions[0].name, this.crds.network.plural, async (e) => {
      logger.debug({ e }, 'Network resource was modified');
      await this.reconcile(e.meta.namespace);
    });
  }

  protected async reconcile(namespace: string): Promise<void> {
    if (this.currentReconciliation)
    {
      logger.trace({
        promise: this.currentReconciliation
      }, 'Reconciliation already in progress; skipping reconciliation');
      return this.currentReconciliation;
    }

    logger.debug({}, 'Reconciliating devices and networks; will create DhcpHosts, DhcpOptions and DnsHosts based off devices and networks');

    this.currentReconciliation = new Promise((resolve, reject) => {
      this.reconciliateNetworks(namespace).then((networks) => {
        this.reconciliateDevices(namespace, networks).then(() => {
        
          // TODO null current recon
          return resolve();
        }).catch(reject);
      }).catch(reject);
    });

    return this.currentReconciliation;
  }

  private async reconciliateDevices(namespace: string, networks: Network[]): Promise<Device[]> {
    logger.debug({}, 'Reconciliating devices; will create DhcpHosts and DnsHosts based off devices and networks');

    const dnsHosts: { [key: string]: dnsmasq.DnsHosts } = {},
          dhcpHosts: { [key: string]: dnsmasq.DhcpHosts } = {};

    return new Promise((resolve, reject) => {
      this.customObjectsClient.listNamespacedCustomObject(
        this.crds.device.group,
        this.crds.device.versions[0].name,
        namespace,
        this.crds.device.plural
      ).then((response) => {
        const devices = (<DeviceList>response.body).items,
              mappedNetworks = networks.reduce(function(networks, network) {
                networks[network.metadata.name] = network;
                return networks;
              }, {});

        devices.forEach((device) => {
          logger.trace({ device }, 'Found device');

          if (device.spec && device.spec.network)
          {
            const network = device.spec.network;
            logger.debug({ device }, 'Found device with network; setting up a DnsHost entry');

            if (mappedNetworks[network])
            {
              if (device.spec.ip)
              {
                if (!dnsHosts[network])
                {
                  logger.trace({
                    device: device,
                    network: network
                  }, 'Missing DnsHost resource for network');
  
                  dnsHosts[network] = new dnsmasq.DnsHosts(`device-based-${network}`);
                  dnsHosts[network].spec = new dnsmasq.DnsHostsSpec();
                  dnsHosts[network].spec.hosts = [];
                }
  
                const host = new dnsmasq.DnsHostsSpecHost();
                host.ip = device.spec.ip;
                host.hostnames = [];
                
                const addHostnames = hostname => {
                  if (config.EnableDeviceShortName)
                  {
                    host.hostnames.push(hostname);
                  }
  
                  if (config.DeviceSuffix)
                  {
                    host.hostnames.push(`${hostname}${config.DeviceSuffix}`);
                  }
                }

                addHostnames(device.metadata.name);

                if (device.spec.additionalHostnames)
                {
                  device.spec.additionalHostnames.forEach(addHostnames);
                }
                
                dnsHosts[network].spec.hosts.push(host);
              }
              else
              {
                logger.debug({ device }, 'Device has no IP configured');
              }

              if (device.spec.dhcp)
              {
                logger.debug({ device }, 'Found device with DHCP enabled; setting up a DhcpHost entry');

                if (!dhcpHosts[network])
                {
                  logger.trace({
                    device: device,
                    network: network
                  }, 'Missing DhcpHost resource for network');
  
                  dhcpHosts[network] = new dnsmasq.DhcpHosts(`device-based-${network}`);
                  dhcpHosts[network].spec = new dnsmasq.DhcpHostsSpec();
                  dhcpHosts[network].spec.hosts = [];
                }
  
                const host = new dnsmasq.DhcpHostsSpecHost();
                host.hostname = device.metadata.name;
                host.ip = device.spec.ip;
                host.macs = device.spec.macAddresses;
                host.setTags = [network];

                dhcpHosts[network].spec.hosts.push(host);
              }
            }
            else
            {
              logger.warn({ device }, 'Invalid `Device` configuration, network does not exist')
            }
          }
        });

        logger.info({
          dhcpHosts: dhcpHosts,
          dnsHosts: dnsHosts
        }, 'Generated device based DhcpHosts and DnsHosts; updating matching resources');

        Promise.all(
          Object.values(dhcpHosts).map(
            dhcpHosts => KubeHelpers.upsertNamespacedCustomResource<dnsmasq.DhcpHosts>(
              this.customObjectsClient,
              this.crds.dhcpHosts,
              config.DnsmasqResourcesNamespace,
              dhcpHosts
            ).then((response) => {
              logger.debug({
                dhcpHosts: dhcpHosts,
                response: response
              }, 'Successfully updated DhcpHosts custom resource');
            }).catch((err) => {
              logger.error({
                dhcpHosts: dhcpHosts,
                error: err
              }, 'Failed to update DhcpHosts custom resource');
            })
          ).concat(Object.values(dnsHosts).map(
            dnsHosts => KubeHelpers.upsertNamespacedCustomResource<dnsmasq.DnsHosts>(
              this.customObjectsClient,
              this.crds.dnsHosts,
              config.DnsmasqResourcesNamespace,
              dnsHosts
            ).then((response) => {
              logger.debug({
                dnsHosts: dnsHosts,
                response: response
              }, 'Successfully updated DnsHosts custom resource');
            }).catch((err) => {
              logger.error({
                dnsHosts: dnsHosts,
                error: err
              }, 'Failed to update DnsHosts custom resource');
            })
          ))
        ).then((response) => {
          logger.debug({
            response: response,
            dhcpHosts: dhcpHosts,
            dnsHosts: dnsHosts
          }, 'Successfully replaced all device based hosts DhcpHosts and DnsHosts custom resources');

          return resolve(devices);
        }).catch((err) => {
          logger.error({
            error: err,
            dhcpHosts: dhcpHosts,
            dnsHosts: dnsHosts
          }, 'Failed to replace all device based hosts DhcpHosts and DnsHosts custom resources');

          return reject(err);
        });
      }).catch((err) => {
        logger.error({ 
          crd: this.crds.device,
          namespace: namespace,
          error: err
        }, 'Failed to list devices');
        return reject(err);
      });
    });
  }

  private async reconciliateNetworks(namespace: string): Promise<Network[]> {
    logger.debug({}, 'Reconciliating networks; will create DhcpOptions based off networks');

    const dhcpOptions = new dnsmasq.DhcpOptions('device-based-hosts');
    dhcpOptions.spec = new dnsmasq.DhcpOptionsSpec();
    dhcpOptions.spec.options = [];

    if (config.DnsmasqDomainName)
    {
      dhcpOptions.spec.options.push(new dnsmasq.DhcpOptionsSpecDhcpOption('option:domain-name', [config.DnsmasqDomainName]));
    }

    if (config.DnsmasqDomainSearch)
    {
      dhcpOptions.spec.options.push(new dnsmasq.DhcpOptionsSpecDhcpOption('option:domain-search', config.DnsmasqDomainSearch));
    }

    return new Promise((resolve, reject) => {
      this.customObjectsClient.listNamespacedCustomObject(
        this.crds.network.group,
        this.crds.network.versions[0].name,
        namespace,
        this.crds.network.plural
      ).then((response) => {
        const networks = (<NetworkList>response.body).items.filter((network) => {
          return network.spec?.serveDhcp && network.spec?.router;
        });

        networks.forEach((network) => {
          logger.trace({ network }, 'Found network');

          dhcpOptions.spec.options.push(new dnsmasq.DhcpOptionsSpecDhcpOption('option:router', [network.spec.router], network.metadata.name));
        });

        logger.info({ dhcpOptions }, 'Generated device based DhcpOptions; updating matching resource');

        KubeHelpers.upsertNamespacedCustomResource<dnsmasq.DhcpOptions>(
          this.customObjectsClient,
          this.crds.dhcpOptions,
          config.DnsmasqResourcesNamespace,
          dhcpOptions
        ).then((response) => {
          logger.debug({
            response: response,
            dhcpOptions: dhcpOptions
          }, 'Successfully replaced device based hosts DhcpOptions custom resource');

          return resolve(networks);
        }).catch((err) => {
          logger.error({
            error: err,
            dhcpOptions: dhcpOptions
          }, 'Failed to replace device based hosts DhcpOptions custom resource');

          return reject(err);
        });
      }).catch((err) => {
        logger.error({ 
          crd: this.crds.network,
          namespace: namespace,
          error: err
        }, 'Failed to list networks');
        return reject(err);
      });
    });
  }
}
