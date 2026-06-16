DO $$
BEGIN
  IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'audit_writer') THEN
    REVOKE SELECT, INSERT ON TABLE audit_events FROM audit_writer;
    REVOKE USAGE ON SCHEMA public FROM audit_writer;
  END IF;
END
$$;
