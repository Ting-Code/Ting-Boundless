# audit-service

Platform service. Consumes audit events and persists them append-only to
`audit_db`. Owns no business logic.

Endpoints: `POST /internal/audit/events`.

Env: `HTTP_ADDR`, `POSTGRES_*` (audit_db).

V1 ingests over HTTP (outbox dispatcher / auth-service); V2 moves to RabbitMQ.
TODO: schema validation, idempotency by event id, append-only storage + retention.
