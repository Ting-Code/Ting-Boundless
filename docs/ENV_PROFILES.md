# Environment profiles

Ting Boundless uses **one `.env` file** at the repo root. Values differ by how you run
the stack. Pick a profile below and adjust the listed variables.

| Profile | When | Compose / run |
|---------|------|----------------|
| **native-local** | Default dev: PG/Redis on host, `go run` + Nest | No app containers |
| **docker-infra** | Data stores in Docker, apps on host | `make up-infra` |
| **docker-full** | Everything containerized | `make up` |
| **cloud** | Managed RDS/Redis, apps in Docker or K8s | `make up-apps` + real endpoints |

`.env.example` defaults to **native-local**. `deploy/docker-compose.yml` overrides
service URLs for containers where needed (e.g. `USER_SERVICE_URL=http://user-service:8081`
on `gateway`).

---

## native-local (default)

Services listen on localhost; databases on the same machine.

| Variable | Value |
|----------|-------|
| `POSTGRES_HOST` | `127.0.0.1` |
| `POSTGRES_PORT` | `5432` |
| `REDIS_ADDR` | `localhost:6379` |
| `RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` |
| `S3_ENDPOINT` | `http://localhost:9000` (optional MinIO) |
| `USER_SERVICE_URL` | `http://127.0.0.1:8081` |
| `BUSINESS_SERVICE_URL` | `http://127.0.0.1:3005` |
| `FILE_SERVICE_URL` | `http://127.0.0.1:8083` |
| `AUTH_SERVICE_URL` | `http://127.0.0.1:8084` |
| `AUDIT_SERVICE_URL` | `http://127.0.0.1:8085` |
| `OIDC_ISSUER` / `OIDC_JWKS_URL` | `http://127.0.0.1:3001/oidc` … (Logto on host) |

**Run:** `make migrate`, `make run-gateway`, `make run-user-service`, `make run-business`, etc.

---

## docker-infra

Postgres/Redis/RabbitMQ in Docker; Go/Nest processes still on the host.

| Variable | Value |
|----------|-------|
| `POSTGRES_HOST` | `127.0.0.1` (published port from compose) |
| `REDIS_ADDR` | `localhost:6379` |
| `RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` |
| Service URLs | Same as **native-local** |

**Run:** `make up-infra` then `go run` / `pnpm dev` on the host.

---

## docker-full

All app images use Docker DNS names for east-west traffic. Keep **browser-facing**
OIDC URLs on published host ports unless you terminate TLS at nginx with a public hostname.

| Variable | In-container value | Notes |
|----------|-------------------|--------|
| `POSTGRES_HOST` | `postgres` | Set in compose infra |
| `REDIS_ADDR` | `redis:6379` | Gateway BFF sessions |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | Worker (optional) |
| `USER_SERVICE_URL` | `http://user-service:8081` | Overridden on `gateway` service |
| `AUTH_SERVICE_URL` | `http://auth-service:8084` | Overridden on `gateway` |
| `AUDIT_SERVICE_URL` | `http://audit-service:8085` | Worker outbox dispatcher |
| `BUSINESS_SERVICE_URL` | `http://host.docker.internal:3005` | Nest often on host in dev compose |
| `OIDC_JWKS_URL` (gateway) | `http://logto:3001/oidc/jwks` | Internal JWKS fetch |
| `OIDC_ISSUER` (tokens) | Still Logto public issuer URL | Must match token `iss` claim |

**Run:** `make up` (infra + apps). Nginx on `:80` → gateway.

**Logto:** compose publishes `3001`/`3002` for admin and OIDC from the host.

---

## cloud / production

Replace placeholders in `.env` with Tencent Cloud (or other) endpoints. See
[`DEPLOY_TENCENT.md`](DEPLOY_TENCENT.md).

- `POSTGRES_SSLMODE=require`
- `INTERNAL_API_TOKEN` **required** (non-empty); `APP_ENV=production` fails startup if missing
- `GATEWAY_BFF_DEV_LOGIN=false`
- `GATEWAY_DEV_JWT_SECRET` unset
- `LOGTO_WEBHOOK_SIGNING_KEY` set; `LOGTO_WEBHOOK_SKIP_VERIFY` unset

---

## Quick checks

| Check | Command |
|-------|---------|
| Postgres | `psql "postgresql://ting:change-me@127.0.0.1:5432/app_db" -c 'select 1'` |
| Gateway ready | `curl -s http://127.0.0.1:8080/readyz` |
| Bearer → user | `curl -H "Authorization: Bearer $(make dev-jwt)" http://127.0.0.1:8080/v1/users/me` |
