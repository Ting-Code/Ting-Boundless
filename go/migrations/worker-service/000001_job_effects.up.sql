-- Idempotent record of platform async job side-effects (RabbitMQ consumer).
CREATE TABLE IF NOT EXISTS worker_job_effects (
    job_id         TEXT PRIMARY KEY,
    job_type       TEXT NOT NULL,
    tenant_id      TEXT NOT NULL DEFAULT '',
    actor_user_id  TEXT NOT NULL DEFAULT '',
    resource_id    TEXT NOT NULL DEFAULT '',
    payload        JSONB NOT NULL DEFAULT '{}',
    processed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_worker_job_effects_type ON worker_job_effects (job_type);
CREATE INDEX IF NOT EXISTS idx_worker_job_effects_tenant ON worker_job_effects (tenant_id, processed_at DESC);
