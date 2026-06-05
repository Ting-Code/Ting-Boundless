# business-service

A core business domain. Enforces domain authorization (ownership, tenant
isolation, business-state rules) locally.

Endpoints: `GET /v1/business/ping`.

Env: `HTTP_ADDR`, `POSTGRES_*`, `REDIS_ADDR`, `RABBITMQ_URL`.

TODO: domain model + migrations; emit domain audit events via outbox.
