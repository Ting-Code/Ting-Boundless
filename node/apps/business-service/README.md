# @ting/business-service

NestJS + Drizzle domain service. Primary API under `/v1/business/*`.

- Trusts **Gateway-injected identity headers** only (no end-user JWT parsing).
- `RequireAuthenticatedMiddleware` on all routes except health/metrics and `/v1/business/ping` (mirrors Go `httpx.TrustedAuth`).
- Exposes `/healthz`, `/readyz`, `/metrics` per platform baseline.
- Uses `@ting/api` for shared API types and paths.

## Endpoints (V1 scaffold)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Postgres probe |
| GET | `/metrics` | Prometheus |
| GET | `/v1/business/ping` | Smoke test |
| GET | `/v1/business/me` | Echo trusted identity (requires auth via middleware) |
| GET | `/v1/business/items` | List items (tenant-scoped) |
| POST | `/v1/business/items` | Create item + outbox event + optional MQ job |
| GET | `/v1/business/items/{id}` | Get item by id (tenant-scoped) |
| PATCH | `/v1/business/items/{id}` | Update title/body + outbox event |
| DELETE | `/v1/business/items/{id}` | Delete item + outbox event |

## Async jobs (RabbitMQ)

After create/update/delete, may publish `business.item.*` jobs to the platform work queue
(same topology as Go worker: `ting.jobs` / routing key `platform`). Skipped when
`RABBITMQ_URL` is unset or a placeholder.

```env
RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/
# RABBITMQ_QUEUE_PREFIX=ting.jobs
```

Run `make run-worker` to consume jobs alongside the outbox → audit dispatcher.

## Run

From repo root:

```bash
make node-install
cd node && pnpm --filter @ting/api build
pnpm dev:business
```

Or from this app:

```bash
pnpm dev
```

Default listen: `:3005` (`BUSINESS_HTTP_ADDR`; avoids Logto on `:3001`).

**Auth:** `RequireAuthenticatedMiddleware` rejects unauthenticated requests on all routes except `/healthz`, `/readyz`, `/metrics`, and `GET /v1/business/ping`.

## Observability

When `OTEL_EXPORTER_OTLP_ENDPOINT` is set, `src/instrument.ts` enables the OpenTelemetry
Node SDK (auto-instrumentation) before Nest boots. Traces and **logs** (JSON stdout + OTLP)
flow: business-service → Collector → Tempo / Loki.

```env
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
# OTEL_LOGS_EXPORTER=otlp   # set to "none" to skip OTLP log fan-out
```

## Database

```bash
# apply drizzle SQL manually for now, or:
pnpm db:migrate
```

Schema: `src/db/schema.ts` (`business_outbox` table for future audit outbox).

## Gateway

Gateway proxies `/v1/business/*` → `BUSINESS_SERVICE_URL` (default `http://127.0.0.1:3005`).
