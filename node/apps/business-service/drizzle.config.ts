import { defineConfig } from 'drizzle-kit';

function env(name: string, fallback = ''): string {
  return process.env[name] ?? fallback;
}

export default defineConfig({
  schema: './src/db/schema.ts',
  out: './drizzle',
  dialect: 'postgresql',
  dbCredentials: {
    host: env('POSTGRES_HOST', '127.0.0.1'),
    port: Number(env('POSTGRES_PORT', '5432')),
    user: env('POSTGRES_USER', 'ting'),
    password: env('POSTGRES_PASSWORD', ''),
    database: env('POSTGRES_DB', 'app_db'),
    ssl: env('POSTGRES_SSLMODE', 'disable') === 'require',
  },
});
