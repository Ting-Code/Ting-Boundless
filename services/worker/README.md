# worker

Consumes async jobs from RabbitMQ (notifications, file processing, audit
dispatch, retries). Exposes `/healthz` + `/readyz` only.

Env: `HTTP_ADDR` (default `:8081`), `RABBITMQ_URL`, `POSTGRES_*`, `REDIS_ADDR`.

TODO: RabbitMQ consumers with DLQ, idempotent handlers, identity/trace
propagation from the message that created the job.
