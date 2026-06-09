# Ting Boundless

A low-cost-to-start, standards-driven, smoothly-scalable backend platform.

Goals: **centralized shared logic, zero-intrusion auth for business services,
multi-language extensibility, observability and auditability** — without early
choices that block future evolution.

## Architecture in one sentence

Nginx guards the edge, a Go Gateway/BFF centralizes shared request logic (no duplicate
auth in Node or Next), Logto owns identity (OIDC/JWKS), Go services own platform
capabilities, NestJS + Drizzle owns domain CRUD, Next.js serves SSR and Vite serves
the admin SPA (shared `node/packages/api-types`), Python is reserved for heavy AI/data
pipelines, `platform-contracts` define cross-language behavior, OpenTelemetry
provides observability, CloudEvents audit events flow into the Audit Service, and
PostgreSQL/Redis/RabbitMQ/S3 form the infrastructure base.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full design and
[`docs/AI_CONTEXT.md`](docs/AI_CONTEXT.md) for the durable rule set.

**Tencent Cloud deploy:** [`docs/DEPLOY_TENCENT.md`](docs/DEPLOY_TENCENT.md) (GitHub Actions → TCR → CVM).

## Repository layout

| Path | Purpose |
|------|---------|
| `platform-contracts/` | Cross-language source of truth: OpenAPI, Protobuf + buf, JSON schemas |
| `node/` | **pnpm monorepo** — all Node/TypeScript (`apps/` + `packages/`) |
| `node/apps/business-service/` | NestJS + Drizzle domain API |
| `node/apps/site/` | Next.js SSR public site |
| `node/apps/admin/` | Vite + React admin SPA |
| `node/packages/api-types/` | Generated TS types from OpenAPI |
| `go/` | **Go monorepo** — see [`go/README.md`](go/README.md) |
| `go/pkg/` | Shared Go libraries (thin SDK over the contracts) |
| `go/services/` | Go platform services only |
| `deploy/` | docker-compose (apps + infra split), nginx, OpenTelemetry collector |
| `docs/` | Architecture docs; full request chain in `ARCHITECTURE.md` |
| `.cursor/` | Cursor rules and skills |

## Services (V1)

| Service | Stack | Role |
|---------|-------|------|
| `gateway` | Go | Edge API entry: token/session verification, identity injection, routing |
| `auth-service` | Go | IdP integration: Logto webhooks, mini-program code2session, token issuance |
| `user-service` | Go | User profile and membership baseline |
| `business-service` | NestJS + Drizzle | Domain CRUD and admin APIs under `/v1/business/*` |
| `file-service` | Go | File upload/download over S3-compatible storage |
| `audit-service` | Go | Consumes audit events, persists to `audit_db` |
| `worker` | Go | Async jobs from RabbitMQ |
| `node/apps/site` | Next.js | SSR public web (behind Gateway) |
| `node/apps/admin` | Vite + React | Admin SPA (static + TanStack Query → `/v1`) |

## Quick start (local dev)

**Default: locally installed PostgreSQL / Redis / RabbitMQ** (not Docker). Services
load `.env` automatically and connect to `localhost`.

### 1. Install data stores locally

| Component | Windows example | Default in `.env` |
|-----------|-----------------|-------------------|
| PostgreSQL 16+ | [postgresql.org/download](https://www.postgresql.org/download/windows/) | `localhost:5432` |
| Redis | Memurai / WSL / [Redis for Windows](https://github.com/redis-windows/redis-windows) | `localhost:6379` |
| RabbitMQ | [rabbitmq.com/install-windows](https://www.rabbitmq.com/docs/install-windows) | `localhost:5672` |

**RabbitMQ 4.3.x requires Erlang 27** (not Erlang 29). If RabbitMQ fails to start,
run `scripts/install-erlang27.ps1` then `scripts/start-rabbitmq.bat`.
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

**Go platform:**

```bash
make build
make migrate                     # users + audit_events tables
make run-user-service            # :8081 — check GET /readyz
make run-gateway                 # :8080
make run-business                # Nest :3005

# Bearer JWT test (mobile/app style)
make dev-jwt my-user
curl -H "Authorization: Bearer $(make dev-jwt my-user)" http://127.0.0.1:8080/v1/users/me

# Web BFF cookie test (no Logto): requires Redis + GATEWAY_BFF_DEV_LOGIN=true
# Open http://127.0.0.1:8080/sign-in/dev?return_to=/admin then:
curl -b cookies.txt -c cookies.txt http://127.0.0.1:8080/sign-in/dev?user_id=web-user
curl -b cookies.txt http://127.0.0.1:8080/v1/users/me

# Real OIDC via Logto (see docs/LOGTO_SETUP.md):
# powershell -File scripts/start-logto.ps1
# Configure Traditional Web App + API resource ting-boundless, set OIDC_CLIENT_ID/SECRET
# Build admin SPA, then open http://127.0.0.1:8080/sign-in?return_to=/admin/items
# (cd node && pnpm --filter @ting/admin build)
```

**Node / TypeScript (pnpm monorepo):**

```bash
make node-install                # or: cd node && pnpm install
make run-business                # Nest :3005
make run-admin                   # Vite admin → Gateway cookie flow
make e2e-admin                   # smoke script (see docs/E2E_ADMIN.md)
cd node && pnpm dev:site         # Next.js
```

See [`node/README.md`](node/README.md).

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
