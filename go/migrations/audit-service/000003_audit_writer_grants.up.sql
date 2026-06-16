-- Restricted runtime role for audit-service (append-only: SELECT + INSERT).
-- Role must exist (see deploy/postgres/setup-local.sql). Skipped when absent (CI without role).
DO $$
BEGIN
  IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'audit_writer') THEN
    GRANT USAGE ON SCHEMA public TO audit_writer;
    GRANT SELECT, INSERT ON TABLE audit_events TO audit_writer;
  END IF;
END
$$;
