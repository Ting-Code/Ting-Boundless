-- V1 uses one PostgreSQL instance with separate databases per concern.
-- app_db is created via POSTGRES_DB; create the others here.
-- Note: audit_db should use restricted, append-only-style credentials in
-- production (see docs/ARCHITECTURE.md). This dev init keeps it simple.

CREATE DATABASE logto_db;
CREATE DATABASE audit_db;
