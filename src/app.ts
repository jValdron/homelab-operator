require('dotenv').config();

import DeviceBasedHosts from './controllers/DeviceBasedHosts';
import IngressBasedDnsHosts from './controllers/IngressBasedDnsHosts';
import LoadBalancerBasedDnsHosts from './controllers/LoadBalancerBasedDnsHosts';

import config from './Config';
import { logger } from './utils/Logger';

(async () => {
  const controllers = [];

  if (config.EnableDeviceBasedHosts)
  {
    logger.debug({}, 'Creating devices based hosts controller');
    const controller = new DeviceBasedHosts();
    logger.info({ controller }, 'Starting devices based hosts controller');
    await controller.start();
    controllers.push(controller);
  }
  else
  {
    logger.warn({ config }, 'Load balancers based DNS hosts controlled is disabled, not setting up');
  }

  if (config.EnableIngressBasedDnsHosts)
  {
    logger.debug({}, 'Creating ingress based DNS hosts controller');
    const controller = new IngressBasedDnsHosts();
    logger.info({ controller }, 'Starting ingress based DNS hosts controller');
    await controller.start();
    controllers.push(controller);
  }
  else
  {
    logger.warn({ config }, 'Ingress based DNS hosts controlled is disabled, not setting up');
  }

  if (config.EnableLoadBalancerBasedDnsHosts)
  {
    logger.debug({}, 'Creating load balancers based DNS hosts controller');
    const controller = new LoadBalancerBasedDnsHosts(config.LoadBalancerBasedSuffix);
    logger.info({ controller }, 'Starting load balancers based DNS hosts controller');
    await controller.start();
    controllers.push(controller);
  }
  else
  {
    logger.warn({ config }, 'Load balancers based DNS hosts controlled is disabled, not setting up');
  }

  const exit = (reason: string) => {
    logger.info({
      reason: reason
    }, 'Stopping all controllers');

    controllers.forEach((controller) => {
      logger.debug({ controller }, 'Stopping controller');
      controller.stop();
    });

    process.exit(0);
  };
  
  process.on('unhandledRejection', (reason, promise) => {
    logger.error({
      reason: reason,
      promise: promise
    }, 'Unhandled rejection within promise');
  });
  
  process.on('SIGTERM', () => exit('SIGTERM'))
         .on('SIGINT', () => exit('SIGINT'));
})();