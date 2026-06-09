import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import type { Pool } from 'pg';

const MIGRATION_FILES = ['0000_business_outbox.sql', '0001_business_items.sql'];

export async function applyMigrations(pool: Pool): Promise<void> {
  const dir = join(__dirname, '..', '..', 'drizzle');
  for (const file of MIGRATION_FILES) {
    const sql = readFileSync(join(dir, file), 'utf8');
    await pool.query(sql);
  }
}
