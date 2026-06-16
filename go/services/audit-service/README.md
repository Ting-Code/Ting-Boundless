# audit-service

Platform service. Consumes audit events and persists them append-only to
`audit_db`. Owns no business logic.

| Method | Path | Auth | Notes |
|--------|------|------|-------|
| POST | `/internal/audit/events` | `X-Internal-Token` | Ingest; validates `data` via `pkg/contracts` proto bridge |
| GET | `/v1/audit/events` | Gateway identity + `admin` role | List recent events (`limit`, `type`, `source`) |

Env: `HTTP_ADDR`, `POSTGRES_*` (migrations / dev runtime), `INTERNAL_API_TOKEN`.

**Production DB split (D-10):** migrations run as `POSTGRES_USER` (schema owner).
Runtime may use a restricted role via `AUDIT_POSTGRES_USER` / `AUDIT_POSTGRES_PASSWORD`
(`audit_writer`: `SELECT` + `INSERT` on `audit_events` only). Bootstrap:
`deploy/postgres/setup-local.sql` or `init/02-audit-writer.sql`; grants in migration
`000003_audit_writer_grants`.

V1 ingests over HTTP; V2 moves ingest to RabbitMQ. Query API is tenant-scoped when
`X-Tenant-Id` is set (includes global gateway events with empty tenant).

```bash
make run-audit-service
curl -H "Authorization: Bearer $(make dev-jwt)" \
  'http://127.0.0.1:8080/v1/audit/events?limit=20'
```
