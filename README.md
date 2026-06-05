# Ting Boundless

A low-cost-to-start, standards-driven, smoothly-scalable backend platform.

Goals: **centralized shared logic, zero-intrusion auth for business services,
multi-language extensibility, observability and auditability** — without early
choices that block future evolution.

## Architecture in one sentence

Nginx guards the edge, a Go Gateway/BFF centralizes shared request logic, Logto owns
identity (OIDC/JWKS), Go services own core business, Python is reserved for AI/data,
`platform-contracts` define cross-language behavior, OpenTelemetry provides
observability, CloudEvents audit events flow into the Audit Service, and
PostgreSQL/Redis/RabbitMQ/S3 form the infrastructure base.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full design and
[`docs/AI_CONTEXT.md`](docs/AI_CONTEXT.md) for the durable rule set.

## Repository layout

| Path | Purpose |
|------|---------|
| `platform-contracts/` | Cross-language source of truth: Protobuf + buf, JSON schemas |
| `pkg/` | Shared Go libraries (thin SDK over the contracts) |
| `services/` | Deployable services (gateway, auth-service, user-service, ...) |
| `deploy/` | docker-compose (apps + infra split), nginx, OpenTelemetry collector |
| `docs/` | Architecture docs |
| `.cursor/` | Cursor rules and skills |

## Services (V1)

| Service | Role |
|---------|------|
| `gateway` | Edge API entry: token verification, identity injection, routing, rate limiting |
| `auth-service` | IdP integration: Logto webhooks, mini-program code2session, token issuance |
| `user-service` | User domain |
| `business-service` | Core business domain |
| `file-service` | File upload/download over S3-compatible storage |
| `audit-service` | Consumes audit events, persists to `audit_db` |
| `worker` | Async jobs from RabbitMQ |

## Quick start (local dev)

**Default: locally installed PostgreSQL / Redis / RabbitMQ** (not Docker). Services
load `.env` automatically and connect to `localhost`.

### 1. Install data stores locally

| Component | Windows example | Default in `.env` |
|-----------|-----------------|-------------------|
| PostgreSQL 16+ | [postgresql.org/download](https://www.postgresql.org/download/windows/) | `localhost:5432` |
| Redis | Memurai / WSL / [Redis for Windows](https://github.com/redis-windows/redis-windows) | `localhost:6379` |
| RabbitMQ | [rabbitmq.com/install-windows](https://www.rabbitmq.com/docs/install-windows) | `localhost:5672` |
| MinIO (optional) | [min.io/download](https://min.io/download) | `localhost:9000` |

### 2. Bootstrap PostgreSQL

**Windows（推荐）：** 双击或在 repo 根目录运行：

```bash
scripts/setup-local.bat
# 或
powershell -ExecutionPolicy Bypass -File scripts/setup-local.ps1
```

按提示输入安装 PostgreSQL 时设置的 **postgres** 超级用户密码。

**手动方式：**

```bash
cp .env.example .env          # edit POSTGRES_PASSWORD to match your setup
"D:/app/PostgreSQL/bin/psql.exe" -U postgres -f deploy/postgres/setup-local.sql
```

Creates role `ting` (password `change-me`) and databases `app_db`, `logto_db`, `audit_db`.
Ensure `.env` has `POSTGRES_PASSWORD=change-me` (or your chosen password).

### 3. Run services

```bash
go build ./...
go run ./services/user-service   # :8081 — check GET /readyz
go run ./services/gateway        # :8080
```

Each service connects at startup and registers dependency checks on `/readyz`.
If a dependency is down, the process still starts but `/readyz` returns 503.

### Production / cloud

Use fake placeholders until real endpoints exist (see commented block in
`.env.example`). When `POSTGRES_HOST` / `REDIS_ADDR` / etc. contain
`placeholder` or `rm-xxx`, services **skip connecting** and `/readyz` reports
`not configured` until you replace them with real cloud values.

### Docker (optional)

Docker Compose is optional for local dev. Use it when you prefer containers
over native installs, or for deployment:

| Mode | When | Command |
|------|------|---------|
| **Native local (default)** | `go run` + locally installed PG/Redis | see above |
| **Docker infra only** | DB in Docker, apps on host | `make up-infra`, set `POSTGRES_HOST=localhost` |
| **Docker full stack** | everything containerized | `make up` |
| **Production apps** | managed RDS/Redis, apps in Docker | `make up-apps` + cloud `.env` |

```bash
make up-infra   # Postgres, Redis, RabbitMQ, MinIO (optional)
make up-apps    # gateway, services, nginx, logto, otel
make up         # both
make down       # stop all
```

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) § Data Infrastructure for rationale.

## Conventions

Every service: `/healthz`, `/readyz`, `/metrics`, JSON stdout logs (ECS-style),
`traceparent` propagation, unified error responses, 12-Factor config via env.
