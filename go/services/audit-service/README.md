# audit-service

Platform service. Consumes audit events and persists them append-only to
`audit_db`. Owns no business logic.

| Method | Path | Auth | Notes |
|--------|------|------|-------|
| POST | `/internal/audit/events` | `X-Internal-Token` | Ingest; validates `data` via `pkg/contracts` proto bridge |
| GET | `/v1/audit/events` | Gateway identity + `admin` role | List recent events (`limit`, `type`, `source`) |

Env: `HTTP_ADDR`, `POSTGRES_*` (audit_db), `INTERNAL_API_TOKEN`.

V1 ingests over HTTP; V2 moves ingest to RabbitMQ. Query API is tenant-scoped when
`X-Tenant-Id` is set (includes global gateway events with empty tenant).

```bash
make run-audit-service
curl -H "Authorization: Bearer $(make dev-jwt)" \
  'http://127.0.0.1:8080/v1/audit/events?limit=20'
```
