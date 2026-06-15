CREATE TABLE IF NOT EXISTS user_identities (
    provider      TEXT NOT NULL,
    provider_uid  TEXT NOT NULL,
    user_id       TEXT NOT NULL,
    tenant_id     TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, provider_uid)
);

CREATE INDEX IF NOT EXISTS idx_user_identities_user_id ON user_identities (user_id);
