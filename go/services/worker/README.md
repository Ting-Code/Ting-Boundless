# worker

Background jobs: **business outbox → audit-service** (V1 HTTP poll) and
**RabbitMQ platform jobs** with DLQ. Exposes `/healthz`, `/readyz`, `/metrics`.

## Outbox dispatcher (V1)

Polls `business_outbox` in `app_db` every 2s (written by Nest `business-service`
in the same transaction as domain writes) and POSTs to `audit-service`.

| Variable | Default | Purpose |
|----------|---------|---------|
| `AUDIT_SERVICE_URL` | `http://127.0.0.1:8085` | Audit ingest base |
| `INTERNAL_API_TOKEN` | (empty) | `X-Internal-Token` for audit ingest |
| `POSTGRES_*` | — | `app_db` connection |

Requires PostgreSQL and a running `audit-service`.

## RabbitMQ consumer (D-07)

Processes platform jobs from `ting.jobs.platform`. `business.item.*` jobs record
idempotent side-effects in `worker_job_effects` (audit still flows via outbox →
audit-service). Invalid payloads retry then DLQ.

| Variable | Default | Purpose |
|----------|---------|---------|
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | Broker |
| `RABBITMQ_QUEUE_PREFIX` | `ting.jobs` | Exchange/queue name prefix |
| `RABBITMQ_PREFETCH` | `10` | Consumer prefetch |
| `RABBITMQ_MAX_RETRIES` | `3` | Requeue attempts before DLQ |

Job envelope (JSON):

```json
{"id":"uuid","type":"ping","time":"2026-06-15T12:00:00Z","payload":{}}
```

## Other env

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8086` | Health/metrics listen |
| `INTERNAL_API_TOKEN` | — | Required for `GET /internal/job-effects` |

## Internal job-effects API

`GET /internal/job-effects` lists recent rows from `worker_job_effects` (idempotent MQ side-effects).

Query: `limit` (default 50), `tenant_id`, `job_type`.

```bash
curl -H "X-Internal-Token: $INTERNAL_API_TOKEN" \
  'http://127.0.0.1:8086/internal/job-effects?limit=20'
```

## Local

```bash
make up-infra          # includes RabbitMQ :5672
make run-worker
make enqueue-ping      # publishes {"type":"ping"} to ting.jobs.platform
```
