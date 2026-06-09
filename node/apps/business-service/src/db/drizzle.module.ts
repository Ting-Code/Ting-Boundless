import { Global, Inject, Logger, Module, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import { drizzle, type NodePgDatabase } from 'drizzle-orm/node-postgres';
import { Pool } from 'pg';
import { isPlaceholder, postgresConfig } from '../config/env';
import * as schema from './schema';
import { applyMigrations } from './apply-migrations';
import { DRIZZLE, PG_POOL } from './drizzle.tokens';

export type AppDatabase = NodePgDatabase<typeof schema>;

@Global()
@Module({
  providers: [
    {
      provide: PG_POOL,
      useFactory: (): Pool | null => {
        const cfg = postgresConfig();
        if (isPlaceholder(cfg.host) || isPlaceholder(cfg.password)) {
          return null;
        }
        return new Pool({
          host: cfg.host,
          port: cfg.port,
          user: cfg.user,
          password: cfg.password,
          database: cfg.database,
          ssl: cfg.ssl ? { rejectUnauthorized: false } : undefined,
          max: 10,
        });
      },
    },
    {
      provide: DRIZZLE,
      inject: [PG_POOL],
      useFactory: (pool: Pool | null): AppDatabase | null => {
        if (!pool) {
          return null;
        }
        return drizzle(pool, { schema });
      },
    },
  ],
  exports: [DRIZZLE, PG_POOL],
})
export class DrizzleModule implements OnModuleInit, OnModuleDestroy {
  private readonly log = new Logger(DrizzleModule.name);

  constructor(@Inject(PG_POOL) private readonly pool: Pool | null) {}

  async onModuleInit(): Promise<void> {
    if (!this.pool) {
      this.log.warn('postgres not configured; skipping migrations');
      return;
    }
    try {
      await applyMigrations(this.pool);
      this.log.log('sql migrations applied');
    } catch (err) {
      this.log.error(`migrations failed: ${String(err)}`);
      throw err;
    }
  }

  async onModuleDestroy(): Promise<void> {
    if (this.pool) {
      await this.pool.end();
    }
  }
}
