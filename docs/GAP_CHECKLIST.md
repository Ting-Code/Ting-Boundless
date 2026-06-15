# GAP_CHECKLIST — Documentation vs Code vs Industry Best Practice

> **Purpose:** Track gaps between architecture docs, implemented code, and industry
> best practices. Use item IDs (e.g. `S-01`) in issues and PRs.
>
> **Scope:** Full repo scan (77 files, ~2900 lines Go, 0 test files).
> **Last updated:** 2026-06-08 (Node/TS consolidated under `node/` pnpm monorepo)
> **Rule:** This document is the single source for gap tracking; fix items in code
> separately — do not duplicate gap lists across README/ARCHITECTURE.

## How to read

| Symbol | Meaning |
|--------|---------|
| ✅ | Aligned with docs and industry baseline |
| 🟡 | Partial — skeleton, doc-only, or inconsistent |
| 🔴 | Missing or conflicts with docs / security baseline |
| 🟢 | P2 — acceptable to defer per evolution roadmap |

| Priority | When |
|----------|------|
| **P0** | Blocks “usable” or creates security hole; fix before any demo/ship |
| **P1** | Required for production MVP or doc/code honesty |
| **P2** | V2/V3 or nice-to-have per `ARCHITECTURE.md` evolution |

---

## Executive summary

```text
Aligned (✅)     23 / 65 items  (~35%)
Partial (🟡)    20 / 65 items  (~31%)
Gap (🔴)        22 / 65 items  (~34%)
```

**Shape:** Governance and docs are production-grade; runtime capabilities (auth,
audit, observability, tests) are early skeleton (~30–35% of industry MVP bar).

**Critical path to “first real loop”:** `S-01` → `S-02` → `S-08` → `P-07`/`P-08`
→ `D-02`/`D-04` — not adding more services.

---

## 1. Security & identity

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| S-01 | P0 | Gateway verifies JWT via cached JWKS; injects trusted identity | Bearer + BFF cookie via `auth.Verifier` + Redis session | ✅ | `gateway/internal/auth/middleware.go` |
| S-02 | P0 | `OIDC_ISSUER`, `OIDC_JWKS_URL`, `OIDC_AUDIENCE` in env | `auth.ConfigFromEnv` supports Logto + `AUTH_OIDC_ISSUER` / `AUTH_JWKS_URL` | ✅ | `go/pkg/auth`, `go/pkg/oidc` |
| S-03 | P1 | Web: BFF Token Handler, OIDC code exchange, HttpOnly cookie | `/sign-in`, `/callback`, `/sign-out`, `/sign-in/dev` | ✅ | `gateway/internal/bff/`, `docs/BFF_LOGTO.md` |
| S-04 | P1 | Mobile: OIDC + PKCE direct to Logto, Bearer to API | Go Gateway Bearer 已支持；`docs/MOBILE_AUTH.md` 文档 only | 🟢 V1 无 Node/App 客户端 | `docs/MOBILE_AUTH.md` |
| S-05 | P1 | Mini-program: code2session → standard JWT | `POST /v1/auth/miniprogram/login` + `GET /v1/auth/jwks`; Gateway dual JWKS | ✅ | `auth-service/`, `go/pkg/auth/issuer.go` |
| S-06 | P1 | Redis revocation/session blocklist for sensitive paths | Subject + session blocklist in Redis; Gateway checks on `GATEWAY_SENSITIVE_PREFIXES`; Logto `User.Deleted` revokes | ✅ | `go/pkg/revocation/`, `gateway/internal/auth/` |
| S-07 | P0 | Strip client `X-User-*` before trust | `StripUntrusted` first in `authenticate()` | ✅ | `go/pkg/identity/identity.go` |
| S-08 | P0 | Client `X-Request-Id` not trusted at edge; gateway regenerates | Gateway `authenticate()` generates `request_id`; no `RequestID` middleware on gateway | ✅ | `gateway/internal/auth/middleware.go` |
| S-09 | P0 | Business services do not parse end-user JWTs | Nest `IdentityMiddleware`; Go `identity.Middleware` | ✅ | `business-service`, `go/services/*/main.go` |
| S-10 | P0 | Service-to-service trust: internal token / network isolation | Gateway injects `X-Internal-Token`; `GatewayTrust` on Go + Nest; prod fails startup if unset | ✅ | `go/pkg/httpx/internal_token.go`, `gatewaytrust.go` |
| S-16 | P0 | Gateway anonymous path whitelist; non-whitelist 401 at edge | Exact+prefix rules; `/sign-in/dev` only when `GATEWAY_BFF_DEV_LOGIN=true` | ✅ | `gateway/internal/auth/anon.go` |
| S-11 | P1 | Logto webhook: signature verify, idempotency, audit mapping | HMAC verify + `webhook_deliveries` + audit emit | ✅ | `auth-service/internal/logto/` |
| S-12 | P1 | Only standard JWT with known issuer + JWKS | auth-service RS256 issuer + JWKS; Gateway verifies via `AUTH_JWKS_URL` | ✅ | `auth-service`, `go/pkg/auth/issuer.go` |
| S-13 | P1 | Gateway must not hard-fail readyz on Logto JWKS | No JWKS probe on gateway readyz | ✅ | `gateway/main.go` |
| S-14 | P2 | Casbin domain AuthZ in business services | Not in code | 🟢 Deferred OK for V1 | `ARCHITECTURE.md` |
| S-15 | P1 | Auth endpoints: stricter rate limit than general API | nginx `zone=auth` 5r/s | ✅ In compose path only | `deploy/nginx/nginx.conf` |

### Gateway edge auth (S-08, S-16)

Gateway chain (outer → inner): `TraceContext` → `AccessLog` → `Recover` → `RateLimit` → `Authenticate` → mux.

---

## 2. Observability

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| O-01 | P1 | `/healthz` liveness only | Implemented | ✅ | `go/pkg/httpx/health.go` |
| O-02 | P1 | `/readyz` probes real dependencies | PG/Redis/MQ/S3 per service | ✅ | `services/*/main.go`, `go/pkg/*/connect.go` |
| O-03 | P1 | Every service exposes `/metrics` (Prometheus) | `httpx.RegisterMetrics` via `health.Handler` on all Go services; Nest `business-service` | ✅ | `go/pkg/httpx/metrics.go`, services `main.go` |
| O-04 | P1 | JSON stdout, ECS-style fields | `go/pkg/logger` with `@timestamp`, `log.level`, `service.name` | ✅ | `go/pkg/logger/logger.go` |
| O-05 | P1 | `request_id` in access logs | `AccessLog` adds `request_id` | ✅ | `go/pkg/httpx/middleware.go` |
| O-06 | P1 | `trace_id` in log lines | `logger.RequestAttrs` extracts from `traceparent`; `TraceContext` on all Go services | ✅ | `go/pkg/logger/request.go`, `go/pkg/httpx/middleware.go` |
| O-07 | P1 | Propagate `traceparent` (W3C) on every hop | Go `TraceContext` on all services; Nest `TraceContextMiddleware`; proxy forwards headers | ✅ | `go/pkg/httpx`, `@ting/logger`, `gateway/internal/proxy` |
| O-08 | P1 | OpenTelemetry SDK → OTLP | Traces + logs fan-out via `pkg/otel` + `httpx.RunService` | ✅ | `go/pkg/otel/`, `go/pkg/httpx/server.go` |
| O-09 | P2 | OTel Collector fans out to Loki/Prom/Tempo | Prom :8889 + Tempo + Loki :3100 + Grafana :3003 | ✅ | `deploy/otel/` |
| O-10 | P2 | Bounded metric label cardinality | N/A until O-03 | 🟢 | `platform-contracts/docs/metrics.md` |

---

## 3. Audit

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| A-01 | P1 | Audit separate from application logs | `go/pkg/audit` distinct from `go/pkg/logger` | ✅ | `go/pkg/audit/audit.go` |
| A-02 | P1 | CloudEvents-style envelope | `audit.Event` + JSON schema | ✅ | `go/pkg/audit/`, `schemas/audit-event.schema.json` |
| A-03 | P1 | Identity events: Logto webhook → auth-service → audit | `POST /internal/webhooks/logto` → `user.login.*` audit | ✅ | `auth-service/internal/logto/` |
| A-04 | P2 | Entry events: gateway async (`api.access.denied`, etc.) | `api.access.denied`, `api.token.invalid`, `api.rate_limited` | ✅ | `gateway/internal/auth`, `gateway/internal/ratelimit` |
| A-05 | P0 | Domain events: Transactional Outbox same DB tx | Nest writes `business_outbox` in tx; worker polls → `audit-service` | ✅ | `business-service`, `worker/internal/outbox` |
| A-06 | P1 | Audit Service append-only → `audit_db` | `POST /internal/audit/events` persists with idempotent `id` | ✅ | `audit-service/internal/store/` |
| A-07 | P1 | Idempotency by event `id` | `ON CONFLICT (id) DO NOTHING` on insert | ✅ | `audit-service/internal/store/events.go` |
| A-08 | P2 | V2: RabbitMQ path for audit dispatch | Worker has no consumers | 🟢 V2 per roadmap | `worker/main.go` |
| A-09 | P1 | Three sources use different delivery (outbox vs async) | Documented V1 paths in ARCHITECTURE § Audit Sources | ✅ | `docs/ARCHITECTURE.md` |
| A-10 | P1 | Admin read API for audit events | `GET /v1/audit/events` (admin role, tenant scope) + admin UI | ✅ | `audit-service/internal/query/`, `admin/AuditPage` |

---

## 4. Data, messaging, storage

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| D-01 | P1 | One PG instance; `app_db`, `logto_db`, `audit_db` | `setup-local.sql`, `init/01-databases.sql` | ✅ | `deploy/postgres/` |
| D-02 | P0 | golang-migrate / Atlas; migrations in Git + CI | `go/migrations/` + Drizzle; `make migrate` + CI `migrate` job | ✅ | `go/cmd/migrate`, `.github/workflows/ci.yml` |
| D-03 | P1 | pgx pool, DSN from env | `go/pkg/db` with ping on open | ✅ | `go/pkg/db/postgres.go` |
| D-04 | P0 | Domain tables (users, business entities) | `users`, `business_items`, `files` + CRUD/upload | ✅ | `user-service`, `business-service`, `file-service` |
| D-05 | P1 | `tenant_id` on tenant-scoped tables | `tenant_id` on users, files, business_items, audit, identities | ✅ | migrations + Drizzle schema |
| D-06 | P2 | Redis: cache, session, revocation | Sessions, rate limits, revocation blocklist | ✅ | `go/pkg/cache/`, `go/pkg/revocation/` |
| D-07 | P2 | RabbitMQ + DLQ for async | Work queue + DLQ + consumer in worker; `pkg/mq` publish/consume | ✅ | `go/pkg/mq/`, `worker/internal/jobs/` |
| D-08 | P2 | S3-compatible file storage | SigV4 PUT/GET + presigned download URL | ✅ | `pkg/storage/`, `file-service/internal/` |
| D-09 | P1 | Cloud placeholder hosts skip connect | `config.IsPlaceholder` | ✅ | `go/pkg/config/placeholder.go` |
| D-10 | P1 | `audit_db` restricted credentials in production | Dev shares `ting` role on all DBs | 🟡 Acceptable V1 dev only | `setup-local.sql`, `ARCHITECTURE.md` |
| D-11 | P1 | `Postgres.Pool()` exposed but unused | All PG services use `Pool()` via `internal/store` layers | ✅ | `go/pkg/db/postgres.go`, `services/*/internal/store/` |

---

## 5. API & contracts (`platform-contracts`)

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| C-01 | P1 | External APIs under `/v1` | All routes use `/v1/` prefix | ✅ | Service handlers |
| C-02 | P1 | buf lint + generate; `gen/` from proto | `make proto` → `go/gen/`; CI verifies stubs + `pkg/contracts` test | ✅ | `platform-contracts/`, `go/gen/` |
| C-03 | P1 | JSON schemas ↔ Go types in sync | `pkg/contracts` proto bridges + round-trip tests for identity/errs/audit | ✅ | `go/pkg/contracts/`, `schemas/*.json` |
| C-04 | P1 | Unified errors via `errs.Write` | Go handlers + Nest `HttpExceptionFilter` use `ErrorEnvelope` | ✅ | `go/pkg/errs`, `business-service` filter |
| C-05 | P1 | buf breaking checks in CI | `buf lint` + `buf generate` + PR `buf breaking` in CI | ✅ | `.github/workflows/ci.yml`, `Makefile` |
| C-08 | P1 | OpenAPI lint in CI | Redocly `lint openapi/*.yaml`; `make lint-openapi` | ✅ | `platform-contracts/redocly.yaml`, CI `contracts` job |
| C-09 | P2 | OpenAPI breaking checks in CI | `oasdiff breaking` vs `main` on PR; `make openapi-breaking` | ✅ | `scripts/oasdiff-breaking.sh`, CI `contracts` job |
| C-06 | P2 | Identity field ↔ header mapping table | Comments in proto/schema only | 🟢 | `identity.proto`, `identity-context.schema.json` |
| C-07 | P2 | gRPC internal APIs (V2) | Not started | 🟢 V2 | `ARCHITECTURE.md` V2 |

---

## 6. Deployment & configuration

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| P-01 | P1 | infra / apps compose split | `docker-compose.infra.yml` + `.yml` | ✅ | `deploy/` |
| P-02 | P1 | 12-Factor env config | `httpx.Env`, `config.LoadEnvFile` | ✅ | `go/pkg/config/`, services |
| P-03 | P1 | Non-root distroless images | `gcr.io/distroless/static-debian12:nonroot` | ✅ | `deploy/Dockerfile` |
| P-04 | P1 | Graceful shutdown on SIGTERM | `httpx.Server.Run` | ✅ | `go/pkg/httpx/server.go` |
| P-05 | P1 | nginx coarse + auth rate limits | `zone=general`, `zone=auth` | ✅ | `deploy/nginx/nginx.conf` |
| P-06 | P2 | HTTPS at edge (certbot / cloud cert) | nginx listens **80 only** | 🟢 Local dev OK | `nginx.conf` |
| P-07 | P0 | Service ports consistent across nginx ↔ process | `/v1/auth/` proxied via Gateway → `auth-service:8084`; no direct nginx→auth port mismatch | ✅ | `deploy/nginx/nginx.conf`, `docker-compose.yml` |
| P-08 | P0 | Gateway upstream URLs use Docker DNS in compose | `USER_SERVICE_URL` etc. set in `docker-compose.yml` | ✅ | `gateway/main.go`, `docker-compose.yml` |
| P-09 | P1 | `.env` for compose uses service names not localhost | `docs/ENV_PROFILES.md` + docker overrides in compose; `.env.example` docker block | ✅ | `docs/ENV_PROFILES.md`, `deploy/docker-compose.yml` |
| P-10 | P1 | Dockerfile Go version matches `go.mod` | Dockerfile `golang:1.25-alpine`; `go/go.mod` `go 1.25.0` | ✅ | `deploy/Dockerfile`, `go/go.mod` |
| P-11 | P1 | Logto in app compose with DB | Service on :3001/:3002 | ✅ | `docker-compose.yml` |
| P-12 | P2 | V1 backups (PG, audit, config) | Documented only | 🟢 | `ARCHITECTURE.md` Backups |
| P-13 | P1 | `depends_on` without health condition | `docker-compose.local.yml` + infra healthchecks for local `make up` | ✅ | `deploy/docker-compose.local.yml`, `docker-compose.infra.yml` |
| P-14 | P2 | Image scan + ACR push in CI | Trivy in `ci.yml`; ACR push in `deploy-tencent.yml` | ✅ | `.github/workflows/` |

### Environment profile matrix (continued analysis)

| Variable | `go run` on host (`.env.example`) | Docker full stack (expected) | Current gap |
|----------|-----------------------------------|------------------------------|-------------|
| `POSTGRES_HOST` | `127.0.0.1` | `postgres` | ✅ documented in ENV_PROFILES |
| `REDIS_ADDR` | `localhost:6379` | `redis:6379` | ✅ |
| `USER_SERVICE_URL` | `http://127.0.0.1:8081` | `http://user-service:8081` | ✅ set in compose |
| `OIDC_JWKS_URL` | `http://127.0.0.1:3001/...` | `http://logto:3001/...` (internal) | ✅ ENV_PROFILES; P-11 host port |
| `HTTP_ADDR` (auth) | `:8084` | Gateway routes to `:8084` in compose | ✅ |

Recommend documenting two profiles: **native-local** vs **docker-full** (no code change required in this file).

---

## 7. Engineering quality

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| Q-01 | P0 | Tests for auth and identity boundary | Gateway auth + edge integration (JWT, 401, rate limit, revocation, request_id) | ✅ | `go/services/gateway/` |
| Q-02 | P1 | Integration test: gateway → user with token | `integration_user_test.go` (Bearer → proxy → upstream headers) | ✅ | `go/services/gateway/` |
| Q-03 | P1 | CI: tidy, vet, build, test, image scan | `ci.yml`: go mod tidy + buf + migrate + node build + Trivy gateway image | ✅ | `.github/workflows/ci.yml` |
| Q-04 | P2 | golangci-lint in pipeline | Optional in Makefile | 🟢 | `Makefile` |
| Q-05 | P1 | `AGENTS.md` entry for AI agents | Present | ✅ | `AGENTS.md` |
| Q-06 | P1 | Cursor rules enforce architecture | 6× `.mdc` rules | ✅ | `.cursor/rules/` |
| Q-07 | P1 | `new-go-service` skill (Golden Path) | Present | ✅ | `.cursor/skills/new-go-service/` |
| Q-08 | P2 | Conventional commits (commitizen) | `.czrc` present | ✅ | `.czrc` |
| Q-09 | P0 | Secrets not in Git | `.env` gitignored; not tracked | ✅ | `.gitignore` |
| Q-10 | P1 | Local Windows bootstrap scripts | `setup-local.ps1`, `verify-local.ps1`, etc. | ✅ | `scripts/` |

---

## 8. Per-service completion matrix

| Service | Default port | Health | Dependencies connected | Business logic | README TODO match |
|---------|-------------|--------|------------------------|----------------|-------------------|
| gateway | :8080 | ✅ | Redis probe | Proxy ✅; auth ❌ | ✅ Accurate |
| auth-service | :8084 | ✅ | PG + Redis | Webhook/mini stub | ✅ Accurate |
| user-service | :8081 | ✅ | PG | `GET/PATCH /v1/users/me` profile | ✅ |
| business-service (Nest) | :3005 | ✅ | PG probe on /readyz | items CRUD + outbox | ✅ |
| file-service | :8083 | ✅ | PG + S3 probe | Upload + metadata + download + presigned URL | ✅ |
| audit-service | :8085 | ✅ | `audit_db` | Ingest + `GET /v1/audit/events` | ✅ |
| worker | :8086 | ✅ | PG + RabbitMQ probe | Outbox → audit; MQ `business.item.*` → `worker_job_effects` | ✅ |

---

## 9. Per-package (`go/pkg/`) maturity

| Package | Industry role | Status | Gap IDs |
|---------|---------------|--------|---------|
| `logger` | ECS JSON stdout | ✅ Production-ready for V1 | — |
| `identity` | Header strip/inject/context | ✅ Gateway strip + inject; pkg/contracts proto bridge | — |
| `errs` | Unified error envelope | ✅ Go `errs.Write` + Nest filter | — |
| `httpx` | Server, health, middleware | ✅ Metrics + trace middleware | S-08 |
| `audit` | Event model + HTTP/async emitters | ✅ HTTPEmitter + Async; worker outbox dispatch | — | `go/pkg/audit/` |
| `db` | pgx pool + readyz | ✅ Connect layer done | — |
| `cache` | Redis client | ✅ Sessions, rate limits, revocation | S-06 |
| `mq` | RabbitMQ client | ✅ Topology, publish, consume + DLQ | — |
| `storage` | S3 probe + SigV4 client | ✅ PUT/GET/presign | — |
| `config` | env + placeholder | ✅ | — |

---

## 10. Document cross-consistency (docs only)

| Check | Result | Notes |
|-------|--------|-------|
| `ARCHITECTURE.md` ↔ `AI_CONTEXT.md` core rules | ✅ | Identity, audit, client auth, language split, end-to-end chain aligned |
| `.cursor/rules/*.mdc` ↔ `AI_CONTEXT.md` | ✅ | Layout includes `node/` pnpm monorepo |
| `ARCHITECTURE.md` Nest `business-service` ↔ code | ✅ | items CRUD + outbox aligned |
| `README.md` claims `/metrics`, `traceparent` | ✅ | O-03, O-07 implemented |
| `AGENTS.md` claims same | ✅ | O-03, O-07 implemented |
| Service README TODO ↔ `main.go` | ✅ | |
| Architecture diagram audit flow (outbox) ↔ code | ✅ | Nest outbox + worker dispatch (A-05) |

---

## 11. Industry benchmark (qualitative)

| Dimension | vs industry MVP | vs industry production |
|-----------|-----------------|------------------------|
| Architecture documentation | Top ~10% (small teams) | Strong |
| Platform / AI governance | Top ~20% | Ahead of many startups |
| Edge security (JWT, trust) | Below MVP | Far |
| Observability (metrics/trace) | Below MVP | Far |
| Audit / compliance path | Designed, not built | Not ready |
| Testing & CI | Bottom quartile | Far |
| Container config correctness | Native dev OK; full compose weak | Needs P-07–P-09 |

---

## 12. Recommended fix order (planning only)

### Phase 1 — Security closed loop (blocks “usable”)

```text
S-08 → S-01 → S-02 → S-10 → P-07 → P-08
```

Deliverable: Bearer token from Logto → gateway verifies → `/v1/users/me` returns real `user_id`.

### Phase 2 — First business persistence

```text
D-02 → D-04 → C-04 → Q-02
```

Deliverable: `users` table + migration; unified errors; one integration test.

### Phase 3 — Audit & observability honesty

```text
A-05 → A-06 → O-03 → O-07 → O-08 → C-02 → Q-03
```

Deliverable: outbox + audit persist; metrics endpoint or doc downgrade; CI workflow.

### Phase 4 — Multi-client auth

```text
S-03 → S-05 → S-11 → S-06
```

Deliverable: Web cookie, mini-program JWT, webhook audit.

### Phase 5 — TypeScript business + **web admin** (V1 focus)

```text
W-01 → W-02 → W-03 → W-06
```

Deliverable: Nest `business-service` CRUD; `@ting/api`（business/users/files/audit OpenAPI）; admin SPA（items / files / account / audit）; Gateway `/v1/business/*` + cookie BFF.

**Deferred (Node):** W-04 `@ting/site` 打磨；小程序/App TS 客户端；`auth.v1` OpenAPI。Go 平台侧小程序登录（S-05）保留，不在此阶段扩展 Node。

---

## 13. Web & TypeScript stack

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| W-01 | P1 | `node/apps/business-service` = NestJS + Drizzle under `/v1/business/*` | items full CRUD + outbox; OpenAPI `business.v1.yaml` | ✅ | `node/apps/business-service/` |
| W-02 | P1 | `@ting/api` generated from OpenAPI | 5 域 spec → openapi-typescript；paths + `apiFetch` | ✅ | `node/packages/api/` |
| W-03 | P1 | `@ting/admin` Vite + TanStack Query → Gateway `/v1` | items + files + account + audit 页 | ✅ | `node/apps/admin/` |
| W-04 | P2 | `@ting/site` Next.js SSR behind Gateway | 脚手架可用；V1 后台主线外 | 🟢 延后 | `node/apps/site/` |
| W-05 | P1 | Gateway/nginx route table for `/v1/business/*`, `/admin`, Next | Gateway site proxy (`SITE_SERVICE_URL`); anon `=/`; nginx `/` → gateway | ✅ | `gateway/main.go`, `deploy/nginx/nginx.conf` |
| W-06 | P2 | OpenAPI specs for `/v1` domains | `business`, `users`, `files`, `audit` + `common` ErrorEnvelope | ✅ | `platform-contracts/openapi/` |
| W-07 | P2 | Nest service template (health, identity guard, outbox) | Only Go `new-go-service` skill | 🟢 | `.cursor/skills/` |
| W-08 | P1 | `node/` pnpm monorepo (`apps/*`, `packages/*`) | ✅ `pnpm-workspace.yaml`, `@ting/*` workspaces | ✅ | `node/` |

---

## 14. References (industry)

| Topic | Reference |
|-------|-----------|
| Gateway / BFF auth | [microservices.io authn part 2](https://microservices.io/post/architecture/2025/05/28/microservices-authn-authz-part-2-authentication.html) |
| BFF token handler | [Auth0 BFF pattern](https://auth0.com/blog/the-backend-for-frontend-pattern-bff/) |
| Edge JWT + headers | [Ory Oathkeeper](https://www.ory.sh/oathkeeper/docs/) |
| Transactional outbox | [microservices.io outbox](https://microservices.io/patterns/data/transactional-outbox.html) |
| Go outbox + pgx | [nikolayk812/pgx-outbox](https://github.com/nikolayk812/pgx-outbox) |
| Contracts | [bufbuild/examples](https://github.com/bufbuild/examples) |
| Observability | [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go) |

---

## 14. Changelog

| Date | Change |
|------|--------|
| 2026-06-05 | Initial full-repo scan; 65 gap items; env profile matrix; package/service matrices |
| 2026-06-08 | Node/TS consolidated under `node/` pnpm monorepo; W-08 added; W-01–W-04 paths updated |
| 2026-06-15 | W-02/W-06: openapi-typescript → `@ting/api`; C-03: `pkg/contracts` proto bridges |
| 2026-06-15 | W-02/W-06: OpenAPI → `@ting/api`; Nest `business.item.created` → RabbitMQ |
| 2026-06-15 | A-09 audit delivery table; Nest OTLP logs fan-out; admin account page |
| 2026-06-15 | Q-03/P-14: CI go mod tidy + Trivy gateway image scan; admin files page |
| 2026-06-05 | W-04/W-05 Next site + Gateway siteproxy; W-01 items full CRUD + admin edit/delete |
| 2026-06-05 | V1 Node 范围收敛：Web 后台优先；`@ting/api` 去掉 authPaths；W-04/site 标延后 |
| 2026-06-05 | Web admin: dev `/sign-in/dev` 默认 `user,admin`；OpenAPI Redocly lint (C-08)；e2e audit step |
| 2026-06-05 | OpenAPI specs: Gateway `servers` + Redocly 规则；admin 审计页详情抽屉 |
| 2026-06-05 | `docs/BFF_LOGTO.md`：Logto BFF 生产路径、Admin 调用契约、联调 checklist |
| 2026-06-05 | C-09: OpenAPI breaking CI (`oasdiff`)；BFF `Secure` cookie；admin SessionBar 显示 roles |

---

*When an item is fixed, update its row (✅) and add a note under Changelog. Do not remove IDs — mark superseded instead.*
