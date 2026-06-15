import './instrument';
import { Logger } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { configureLogger, createNestLoggerAdapter, logger } from '@ting/logger';
import { AppModule } from './app.module';
import { loadEnvFiles, listenPort, assertInternalTokenConfigured } from './config/env';

async function bootstrap(): Promise<void> {
  loadEnvFiles();
  assertInternalTokenConfigured();
  configureLogger({
    service: 'business-service',
    level: process.env.LOG_LEVEL ?? 'info',
  });

  const app = await NestFactory.create(AppModule, {
    logger: false,
  });

  Logger.overrideLogger(createNestLoggerAdapter());

  const port = listenPort();
  await app.listen(port);

  logger.info('listening', { port, addr: `:${port}` });
}

bootstrap().catch((err: unknown) => {
  logger.error('bootstrap failed', { error: String(err) });
  process.exit(1);
});
