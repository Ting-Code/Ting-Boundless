import { config as loadDotenv } from 'dotenv';
import { existsSync } from 'node:fs';
import { resolve } from 'node:path';

const candidates = [
  resolve(process.cwd(), '.env'),
  resolve(process.cwd(), '../.env'),
  resolve(process.cwd(), '../../.env'),
  resolve(process.cwd(), '../../../.env'),
];

export function loadEnvFiles(): void {
  for (const path of candidates) {
    if (existsSync(path)) {
      loadDotenv({ path });
      return;
    }
  }
}

export function env(name: string, fallback = ''): string {
  return process.env[name] ?? fallback;
}

/** True when APP_ENV=production or REQUIRE_INTERNAL_TOKEN=true. */
export function requireInternalToken(): boolean {
  if (env('REQUIRE_INTERNAL_TOKEN') === 'true') {
    return true;
  }
  return env('APP_ENV').trim().toLowerCase() === 'production';
}

/** Fails fast when production requires INTERNAL_API_TOKEN but it is unset. */
export function assertInternalTokenConfigured(): void {
  const token = env('INTERNAL_API_TOKEN').trim();
  if (requireInternalToken() && token === '') {
    throw new Error(
      'INTERNAL_API_TOKEN is required when APP_ENV=production or REQUIRE_INTERNAL_TOKEN=true',
    );
  }
}

/** Align with go/pkg/config/placeholder.go — local dev values are not placeholders. */
export function isPlaceholder(value: string): boolean {
  const v = value.trim().toLowerCase();
  if (v === '') {
    return true;
  }
  const markers = [
    'placeholder',
    'rm-xxx',
    'redis-placeholder',
    'oss-placeholder',
    'mq-placeholder',
    'example.invalid',
  ];
  return markers.some((m) => v.includes(m));
}

export function listenPort(): number {
  const raw = env('BUSINESS_HTTP_ADDR', env('HTTP_ADDR', ':3005'));
  const match = raw.match(/:(\d+)$/);
  if (match) {
    return Number(match[1]);
  }
  const asNum = Number(raw);
  return Number.isFinite(asNum) && asNum > 0 ? asNum : 3005;
}

export function postgresConfig() {
  return {
    host: env('POSTGRES_HOST', '127.0.0.1'),
    port: Number(env('POSTGRES_PORT', '5432')),
    user: env('POSTGRES_USER', 'ting'),
    password: env('POSTGRES_PASSWORD', ''),
    database: env('POSTGRES_DB', 'app_db'),
    ssl: env('POSTGRES_SSLMODE', 'disable') === 'require',
  };
}
