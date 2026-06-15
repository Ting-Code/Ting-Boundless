import { initOtelFromEnv } from '@ting/logger';

const shutdown = initOtelFromEnv('business-service');

async function onExit(): Promise<void> {
  if (shutdown) {
    await shutdown();
  }
}

process.once('SIGTERM', () => {
  void onExit();
});
process.once('SIGINT', () => {
  void onExit();
});
