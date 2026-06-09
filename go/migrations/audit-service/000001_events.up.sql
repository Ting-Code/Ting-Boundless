CREATE TABLE IF NOT EXISTS audit_events (
    id            TEXT PRIMARY KEY,
    source        TEXT NOT NULL,
    type          TEXT NOT NULL,
    subject       TEXT,
    event_time    TIMESTAMPTZ NOT NULL,
    tenant_id     TEXT,
    actor_user_id TEXT,
    data          JSONB,
    received_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_event_time ON audit_events (event_time DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_id ON audit_events (tenant_id);
