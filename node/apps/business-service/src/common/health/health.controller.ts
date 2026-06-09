import { Controller, Get, Header, HttpException, HttpStatus, Inject } from '@nestjs/common';
import { register } from 'prom-client';
import type { Pool } from 'pg';
import { PG_POOL } from '../../db/drizzle.tokens';

@Controller()
export class HealthController {
  constructor(@Inject(PG_POOL) private readonly pool: Pool | null) {}

  @Get('healthz')
  healthz(): { status: string } {
    return { status: 'ok' };
  }

  @Get('readyz')
  async readyz(): Promise<{ status: string; checks: Record<string, string> }> {
    const checks: Record<string, string> = {
      process: 'ok',
    };

    if (!this.pool) {
      checks.postgres = 'not_configured';
      throw new HttpException(
        { code: 'ready.not_configured', message: 'postgres not configured', checks },
        HttpStatus.SERVICE_UNAVAILABLE,
      );
    }

    try {
      await this.pool.query('SELECT 1');
      checks.postgres = 'ok';
    } catch {
      checks.postgres = 'unavailable';
      throw new HttpException(
        { code: 'ready.postgres_unavailable', message: 'postgres unavailable', checks },
        HttpStatus.SERVICE_UNAVAILABLE,
      );
    }

    return { status: 'ok', checks };
  }

  @Get('metrics')
  @Header('Content-Type', 'text/plain; version=0.0.4; charset=utf-8')
  async metrics(): Promise<string> {
    register.setDefaultLabels({ service_name: 'business-service' });
    return register.metrics(); // Prometheus text; Nest serializes as text/plain body
  }
}
