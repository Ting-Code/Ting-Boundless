CREATE TABLE IF NOT EXISTS files (
    id            TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL DEFAULT '',
    owner_id      TEXT NOT NULL,
    bucket        TEXT NOT NULL,
    object_key    TEXT NOT NULL,
    content_type  TEXT NOT NULL DEFAULT '',
    size_bytes    BIGINT NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_files_bucket_object_key ON files (bucket, object_key);
CREATE INDEX IF NOT EXISTS idx_files_tenant_id ON files (tenant_id);
CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files (owner_id);
