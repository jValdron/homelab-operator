import Logger from 'pino';

import config from '../Config';

export const logger = Logger({
  name: 'homelab-operator',
  level: config.LogLevel
});