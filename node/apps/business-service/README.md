# @ting/business-service

NestJS + Drizzle domain service. Primary API under `/v1/business/*`.

- Trusts **Gateway-injected identity headers** only (no end-user JWT parsing).
- Exposes `/healthz`, `/readyz`, `/metrics` per platform baseline.
- Uses `@ting/api-types` for shared response shapes.

## Endpoints (V1 scaffold)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Postgres probe |
| GET | `/metrics` | Prometheus |
| GET | `/v1/business/ping` | Smoke test |
| GET | `/v1/business/me` | Echo trusted identity (requires `X-User-Id`) |
| GET | `/v1/business/items` | List items (tenant-scoped) |
| POST | `/v1/business/items` | Create item + outbox event |

## Run

From repo root:

```bash
make node-install
cd node && pnpm --filter @ting/api-types build
pnpm dev:business
```

Or from this app:

```bash
pnpm dev
```

Default listen: `:3005` (`BUSINESS_HTTP_ADDR`; avoids Logto on `:3001`).

## Database

```bash
# apply drizzle SQL manually for now, or:
pnpm db:migrate
```

Schema: `src/db/schema.ts` (`business_outbox` table for future audit outbox).

## Gateway

Gateway proxies `/v1/business/*` → `BUSINESS_SERVICE_URL` (default `http://127.0.0.1:3005`).
