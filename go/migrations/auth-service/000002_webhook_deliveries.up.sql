CREATE TABLE IF NOT EXISTS webhook_deliveries (
    delivery_key  TEXT PRIMARY KEY,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
