-- Local PostgreSQL bootstrap (run once as superuser).
--   psql -U postgres -f deploy/postgres/setup-local.sql
-- Safe to re-run: resets ting password; CREATE DATABASE may warn if already exists.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'ting') THEN
    CREATE ROLE ting LOGIN CREATEROLE PASSWORD 'change-me';
  ELSE
    ALTER ROLE ting WITH LOGIN CREATEROLE PASSWORD 'change-me';
  END IF;
END
$$;

-- CREATE DATABASE cannot run inside PL/pgSQL; keep as standalone statements.
CREATE DATABASE app_db OWNER ting;
CREATE DATABASE logto_db OWNER ting;
CREATE DATABASE audit_db OWNER ting;

GRANT ALL PRIVILEGES ON DATABASE app_db TO ting;
GRANT ALL PRIVILEGES ON DATABASE logto_db TO ting;
GRANT ALL PRIVILEGES ON DATABASE audit_db TO ting;

-- Restricted runtime user for audit-service (see AUDIT_POSTGRES_* env).
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'audit_writer') THEN
    CREATE ROLE audit_writer LOGIN PASSWORD 'change-me-audit';
  ELSE
    ALTER ROLE audit_writer WITH LOGIN PASSWORD 'change-me-audit';
  END IF;
END
$$;

GRANT CONNECT ON DATABASE audit_db TO audit_writer;
