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
| S-02 | P0 | `OIDC_ISSUER`, `OIDC_JWKS_URL`, `OIDC_AUDIENCE` in env | `auth.ConfigFromEnv`, `oidc.ConfigFromEnv` | ✅ | `go/pkg/auth`, `go/pkg/oidc` |
| S-03 | P1 | Web: BFF Token Handler, OIDC code exchange, HttpOnly cookie | `/sign-in`, `/callback`, `/sign-out`, `/sign-in/dev` | ✅ | `gateway/internal/bff/` |
| S-04 | P1 | Mobile: OIDC + PKCE direct to Logto, Bearer to API | No app-side integration in repo (expected client-side) | 🟡 Documented only | `AI_CONTEXT.md` |
| S-05 | P1 | Mini-program: code2session → standard JWT | `POST /v1/auth/miniprogram/login` returns 501 | 🔴 Not implemented | `go/services/auth-service/main.go` |
| S-06 | P1 | Redis revocation/session blocklist for sensitive paths | Gateway connects Redis; no revocation lookup | 🟡 Wired, unused | `gateway/main.go`, `go/pkg/cache/` |
| S-07 | P0 | Strip client `X-User-*` before trust | `StripUntrusted` first in `authenticate()` | ✅ | `go/pkg/identity/identity.go` |
| S-08 | P0 | Client `X-Request-Id` not trusted at edge; gateway regenerates | Gateway `authenticate()` generates `request_id`; no `RequestID` middleware on gateway | ✅ | `gateway/internal/auth/middleware.go` |
| S-09 | P0 | Business services do not parse end-user JWTs | Nest `IdentityMiddleware`; Go `identity.Middleware` | ✅ | `business-service`, `go/services/*/main.go` |
| S-10 | P0 | Service-to-service trust: internal token / network isolation | Gateway injects `X-Internal-Token`; `httpx.GatewayTrust` on Go + Nest services | 🟡 Dev skips when token unset | `go/pkg/httpx/gatewaytrust.go`, `proxy.go` |
| S-16 | P0 | Gateway anonymous path whitelist; non-whitelist 401 at edge | `GATEWAY_ANON_PREFIXES`, `anon.Allows()` | ✅ | `gateway/internal/auth/anon.go` |
| S-11 | P1 | Logto webhook: signature verify, idempotency, audit mapping | `handleLogtoWebhook` returns 202 with no logic | 🔴 Stub only | `auth-service/main.go` |
| S-12 | P1 | Only standard JWT with known issuer + JWKS | No token minting yet; rule documented | 🟡 Enforce when implementing S-05 | `auth-service`, `ARCHITECTURE.md` Token Issuance |
| S-13 | P1 | Gateway must not hard-fail readyz on Logto JWKS | No JWKS probe on gateway readyz | ✅ | `gateway/main.go` |
| S-14 | P2 | Casbin domain AuthZ in business services | Not in code | 🟢 Deferred OK for V1 | `ARCHITECTURE.md` |
| S-15 | P1 | Auth endpoints: stricter rate limit than general API | nginx `zone=auth` 5r/s | ✅ In compose path only | `deploy/nginx/nginx.conf` |

### Gateway edge auth (S-08, S-16)

Gateway chain (outer → inner): `authenticate` → `Recover` → `AccessLog` → mux.
`authenticate` strips untrusted headers, generates `request_id`, enforces anon whitelist.

---

## 2. Observability

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| O-01 | P1 | `/healthz` liveness only | Implemented | ✅ | `go/pkg/httpx/health.go` |
| O-02 | P1 | `/readyz` probes real dependencies | PG/Redis/MQ/S3 per service | ✅ | `services/*/main.go`, `go/pkg/*/connect.go` |
| O-03 | P1 | Every service exposes `/metrics` (Prometheus) | **No `/metrics` handler anywhere** | 🔴 Docs contradict code | `AGENTS.md`, `AI_CONTEXT.md`, all services |
| O-04 | P1 | JSON stdout, ECS-style fields | `go/pkg/logger` with `@timestamp`, `log.level`, `service.name` | ✅ | `go/pkg/logger/logger.go` |
| O-05 | P1 | `request_id` in access logs | `AccessLog` adds `request_id` | ✅ | `go/pkg/httpx/middleware.go` |
| O-06 | P1 | `trace_id` in log lines | Not populated | 🟡 | `schemas/logging.schema.json` |
| O-07 | P1 | Propagate `traceparent` (W3C) on every hop | Not implemented; proxy does not forward trace headers explicitly | 🔴 | All services, `internal/proxy` |
| O-08 | P1 | OpenTelemetry SDK → OTLP | No SDK imports or init | 🔴 | All `main.go` |
| O-09 | P2 | OTel Collector fans out to Loki/Prom/Tempo | Collector config exists; **debug exporter only** | 🟡 Infra only | `deploy/otel/collector.yaml` |
| O-10 | P2 | Bounded metric label cardinality | N/A until O-03 | 🟢 | `platform-contracts/docs/metrics.md` |

---

## 3. Audit

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| A-01 | P1 | Audit separate from application logs | `go/pkg/audit` distinct from `go/pkg/logger` | ✅ | `go/pkg/audit/audit.go` |
| A-02 | P1 | CloudEvents-style envelope | `audit.Event` + JSON schema | ✅ | `go/pkg/audit/`, `schemas/audit-event.schema.json` |
| A-03 | P1 | Identity events: Logto webhook → auth-service → audit | Webhook stub | 🔴 | `auth-service/main.go` |
| A-04 | P2 | Entry events: gateway async (`api.access.denied`, etc.) | No emitter | 🔴 | `gateway/main.go` |
| A-05 | P0 | Domain events: Transactional Outbox same DB tx | No `outbox` table, writer, or dispatcher | 🔴 | Business services, `AI_CONTEXT.md` Audit Rules |
| A-06 | P1 | Audit Service append-only → `audit_db` | `POST /internal/audit/events` decodes nothing, no INSERT | 🔴 | `audit-service/main.go` |
| A-07 | P1 | Idempotency by event `id` | Not implemented | 🟡 | `audit-service` |
| A-08 | P2 | V2: RabbitMQ path for audit dispatch | Worker has no consumers | 🟢 V2 per roadmap | `worker/main.go` |
| A-09 | P1 | Three sources use different delivery (outbox vs async) | Documented; code treats all as HTTP stub | 🔴 | `ARCHITECTURE.md` Audit Sources |

---

## 4. Data, messaging, storage

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| D-01 | P1 | One PG instance; `app_db`, `logto_db`, `audit_db` | `setup-local.sql`, `init/01-databases.sql` | ✅ | `deploy/postgres/` |
| D-02 | P0 | golang-migrate / Atlas; migrations in Git + CI | `go/migrations/` present | 🟡 CI not wired | `go/migrations/`, `go/cmd/migrate` |
| D-03 | P1 | pgx pool, DSN from env | `go/pkg/db` with ping on open | ✅ | `go/pkg/db/postgres.go` |
| D-04 | P0 | Domain tables (users, business entities) | No SQL schema, no queries | 🔴 | `user-service`, `business-service` |
| D-05 | P1 | `tenant_id` on tenant-scoped tables | Field in `Identity` only | 🟡 Model without storage | `go/pkg/identity/identity.go` |
| D-06 | P2 | Redis: cache, session, revocation | Connect + readyz only | 🟡 | `go/pkg/cache/` |
| D-07 | P2 | RabbitMQ + DLQ for async | Connect + readyz; worker stub | 🟡 | `go/pkg/mq/`, `worker/` |
| D-08 | P2 | S3-compatible file storage | TCP probe to endpoint; upload 501 | 🟡 | `go/pkg/storage/`, `file-service` |
| D-09 | P1 | Cloud placeholder hosts skip connect | `config.IsPlaceholder` | ✅ | `go/pkg/config/placeholder.go` |
| D-10 | P1 | `audit_db` restricted credentials in production | Dev shares `ting` role on all DBs | 🟡 Acceptable V1 dev only | `setup-local.sql`, `ARCHITECTURE.md` |
| D-11 | P1 | `Postgres.Pool()` exposed but unused | Pool opened, no repositories | 🟡 Connection without data layer | `go/pkg/db/postgres.go` |

---

## 5. API & contracts (`platform-contracts`)

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| C-01 | P1 | External APIs under `/v1` | All routes use `/v1/` prefix | ✅ | Service handlers |
| C-02 | P1 | buf lint + generate; `gen/` from proto | proto×3, buf config; **no `gen/`** (gitignored, never generated in repo) | 🟡 Contract doc-only | `platform-contracts/` |
| C-03 | P1 | JSON schemas ↔ Go types in sync | Hand-written `go/pkg/errs`, `go/pkg/audit`, `go/pkg/identity` | 🟡 Manual sync risk | `schemas/*.json`, `go/pkg/` |
| C-04 | P1 | Unified errors via `errs.Write` | `go/pkg/errs` exists; handlers use ad-hoc `httpx.JSON` | 🟡 | e.g. `user-service/main.go` `handleMe` |
| C-05 | P1 | buf breaking checks in CI | `make proto-breaking` only; no CI workflow | 🟡 | `Makefile` |
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
| P-07 | P0 | Service ports consistent across nginx ↔ process | nginx `auth_service:8080`; auth-service default **`:8084`** | 🔴 Docker auth routes fail | `nginx.conf`, `auth-service/main.go` |
| P-08 | P0 | Gateway upstream URLs use Docker DNS in compose | Defaults `http://127.0.0.1:8081`…; compose does not set `USER_SERVICE_URL` etc. | 🔴 Gateway cannot reach peers in containers | `gateway/main.go`, `docker-compose.yml` |
| P-09 | P1 | `.env` for compose uses service names not localhost | `.env.example` uses `127.0.0.1` / `localhost` for PG, Redis, URLs | 🟡 Two env profiles needed; undocumented split | `.env.example`, `README.md` |
| P-10 | P1 | Dockerfile Go version matches `go.mod` | Dockerfile `golang:1.25-alpine`; `go/go.mod` `go 1.25.0` | ✅ | `deploy/Dockerfile`, `go/go.mod` |
| P-11 | P1 | Logto in app compose with DB | Service present; **no published port** for 3001 in compose snippet | 🟡 Host access to Logto unclear | `docker-compose.yml` |
| P-12 | P2 | V1 backups (PG, audit, config) | Documented only | 🟢 | `ARCHITECTURE.md` Backups |
| P-13 | P1 | `depends_on` without health condition | compose `depends_on` only; no `condition: service_healthy` | 🟡 Race on startup | `docker-compose.yml` |
| P-14 | P2 | Image scan + ACR push in CI | Documented in AI_CONTEXT; no workflow | 🟢 | `AI_CONTEXT.md` Minimal CI |

### Environment profile matrix (continued analysis)

| Variable | `go run` on host (`.env.example`) | Docker full stack (expected) | Current gap |
|----------|-----------------------------------|------------------------------|-------------|
| `POSTGRES_HOST` | `127.0.0.1` | `postgres` | P-09: same file used for both |
| `REDIS_ADDR` | `localhost:6379` | `redis:6379` | P-09 |
| `USER_SERVICE_URL` | `http://127.0.0.1:8081` | `http://user-service:8081` | P-08 |
| `OIDC_JWKS_URL` | `http://127.0.0.1:3001/...` | `http://logto:3001/...` (internal) | P-09, P-11 |
| `HTTP_ADDR` (auth) | `:8084` | should match nginx `:8080` OR nginx fix | P-07 |

Recommend documenting two profiles: **native-local** vs **docker-full** (no code change required in this file).

---

## 7. Engineering quality

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| Q-01 | P0 | Tests for auth and identity boundary | **0** `*_test.go` files | 🔴 | Entire repo |
| Q-02 | P1 | Integration test: gateway → user with token | None | 🔴 | — |
| Q-03 | P1 | CI: tidy, vet, build, test, image scan | `make ci` local only; no `.github/workflows` | 🟡 | `Makefile` |
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
| user-service | :8081 | ✅ | PG | `/me` echoes headers | ✅ Accurate |
| business-service (Nest) | :3005 | ✅ | PG probe on /readyz | items CRUD + outbox | 🟡 W-01: more domains pending |
| file-service | :8083 | ✅ | PG + S3 probe | Upload 501 | ✅ Accurate |
| audit-service | :8085 | ✅ | `audit_db` | Ingest stub | ✅ Accurate |
| worker | :8086 | ✅ | RabbitMQ probe | No consumers | ✅ Accurate |

---

## 9. Per-package (`go/pkg/`) maturity

| Package | Industry role | Status | Gap IDs |
|---------|---------------|--------|---------|
| `logger` | ECS JSON stdout | ✅ Production-ready for V1 | — |
| `identity` | Header strip/inject/context | 🟡 Core OK; edge order issue | S-08 |
| `errs` | Unified error envelope | 🟡 Package OK; adoption incomplete | C-04 |
| `httpx` | Server, health, middleware | 🟡 Missing metrics, trace | O-03, O-07, S-08 |
| `audit` | Event model + Emitter iface | 🟡 Types only; no outbox impl | A-05 |
| `db` | pgx pool + readyz | ✅ Connect layer done | D-04, D-11 |
| `cache` | Redis client | 🟡 Connect only | S-06 |
| `mq` | RabbitMQ client | 🟡 Connect only | D-07 |
| `storage` | S3 probe | 🟡 Probe only | D-08 |
| `config` | env + placeholder | ✅ | — |

---

## 10. Document cross-consistency (docs only)

| Check | Result | Notes |
|-------|--------|-------|
| `ARCHITECTURE.md` ↔ `AI_CONTEXT.md` core rules | ✅ | Identity, audit, client auth, language split, end-to-end chain aligned |
| `.cursor/rules/*.mdc` ↔ `AI_CONTEXT.md` | ✅ | Layout includes `node/` pnpm monorepo |
| `ARCHITECTURE.md` Nest `business-service` ↔ code | 🟡 | Nest scaffold live; full CRUD still W-01 |
| `README.md` claims `/metrics`, `traceparent` | 🔴 | See O-03, O-07 |
| `AGENTS.md` claims same | 🔴 | See O-03, O-07 |
| Service README TODO ↔ `main.go` | ✅ | |
| Architecture diagram audit flow (outbox) ↔ code | 🔴 | Diagram correct; code bypasses outbox (A-05) |

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

### Phase 5 — TypeScript business + web tier (per End-to-End Request Chain)

```text
W-01 → W-02 → W-03 → W-04 → W-05
```

Deliverable: Nest `node/apps/business-service` first CRUD; `@ting/api-types`; admin SPA one list page; Gateway route `/v1/business/*`.

---

## 13. Web & TypeScript stack

| ID | Priority | Docs / industry | Code reality | Gap | Locations |
|----|----------|-----------------|--------------|-----|-----------|
| W-01 | P1 | `node/apps/business-service` = NestJS + Drizzle under `/v1/business/*` | items list/create + outbox; OpenAPI `business.v1.yaml` | 🟡 more domains pending | `node/apps/business-service/` |
| W-02 | P1 | `@ting/api-types` generated from OpenAPI | Workspace placeholder only | 🟡 Scaffolded | `node/packages/api-types/`, `platform-contracts/` |
| W-03 | P1 | `@ting/admin` Vite + TanStack Query → Gateway `/v1` | items page list/create; Vite proxy to Gateway | 🟡 more pages pending | `node/apps/admin/` |
| W-04 | P2 | `@ting/site` Next.js SSR behind Gateway | Workspace placeholder only | 🟡 Scaffolded | `node/apps/site/` |
| W-05 | P1 | Gateway/nginx route table for `/v1/business/*`, `/admin`, Next | Gateway `/v1/auth/` + anon whitelist; nginx `/admin` static + auth via gateway | 🟡 Next SSR pending | `gateway/main.go`, `deploy/nginx/nginx.conf` |
| W-06 | P2 | OpenAPI spec for business domain (source for api-types) | `openapi/business.v1.yaml`; generate script pending | 🟡 | `platform-contracts/openapi/` |
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

---

*When an item is fixed, update its row (✅) and add a note under Changelog. Do not remove IDs — mark superseded instead.*
