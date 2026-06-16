-- Restricted append-only credentials for audit-service runtime (production-style).
-- Migrations still run as POSTGRES_USER (ting). Set AUDIT_POSTGRES_* in audit-service env.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'audit_writer') THEN
    CREATE ROLE audit_writer LOGIN PASSWORD 'change-me-audit';
  END IF;
END
$$;

GRANT CONNECT ON DATABASE audit_db TO audit_writer;
