# AI Context For Ting Boundless

## Purpose

This file gives AI coding agents a compact, durable understanding of the project architecture. Use it before proposing implementation plans, creating services, changing authentication, adding observability, or designing cross-language APIs.

## Architecture In One Sentence

Nginx guards the edge, Go Gateway/BFF centralizes shared request logic without duplication, Logto owns identity through OIDC/JWKS, Go services own platform capabilities, NestJS + Drizzle owns domain CRUD in node/apps/business-service; Next.js and Vite admin in node/apps share Gateway cookies and /v1 APIs via @ting/api-types, Python is reserved for heavy AI/data pipelines, platform-contracts define cross-language behavior, OpenTelemetry provides observability, CloudEvents audit events flow into Audit Service, and PostgreSQL/Redis/RabbitMQ/S3 form the infrastructure base.

## Non-Negotiable Rules

### Identity Boundary

```text
Gateway decides who the caller is and whether the request may enter.
Business services decide whether the caller may operate on a specific business resource.
```

Do not move domain authorization into the Gateway. Do not make business services parse end-user JWTs in the normal request path.

### Identity Context Propagation

Every request path must preserve identity and trace context:

```text
Client -> Gateway -> Service
Service A -> Service B
Worker -> Service
```

Required context fields:

```text
request_id
trace_id
actor_user_id
tenant_id
roles
scopes
auth_subject
```

Async jobs must store the actor context that created the job.

### Header Trust

External identity headers are untrusted.

The Gateway must strip these from incoming external requests before injecting trusted values:

```text
X-User-Id
X-Tenant-Id
X-Roles
X-Scopes
X-Auth-Subject
X-Request-Id
```

Business services should only trust identity headers from internal Gateway traffic.

A client-supplied `X-Request-Id` is not trusted at the edge: the Gateway generates a fresh `request_id`. `X-Request-Id` is only trusted and propagated inside the internal service chain as the correlation id.

### IdP Contract

Logto is the initial IdP implementation. The architecture depends on standards:

```text
OIDC
OAuth2
JWKS
JWT claims
```

Avoid Logto-specific assumptions in business services.

### Platform Contracts Before SDKs

This is a multi-language architecture. Define shared behavior in contracts first.

Core contracts:

```text
Protobuf + buf
ECS-style structured logging schema
OpenTelemetry propagation
CloudEvents audit event schema
Unified error response model
Identity context schema
```

Language SDKs are helpers, not the source of truth.

## Preferred Technology Choices

### Edge

```text
Nginx
```

Responsibilities:

```text
HTTPS
domain routing
static assets
basic security headers
request body limits
reverse proxy
```

### Identity

```text
Logto
```

Responsibilities:

```text
phone login
WeChat login
third-party login
password login if needed
OIDC/OAuth2
access and refresh tokens
identity user management
```

Login and verification-code endpoints are public and high-risk. Protect them with stricter controls than ordinary APIs:

```text
Nginx IP throttling on auth paths
Gateway per-IP / per-device / per-phone limits
SMS quotas and cooldown windows
CAPTCHA or challenge for suspicious patterns
optional WAF for production auth endpoints
```

### Auth Service / IdP Integration

`auth-service` is the integration layer between external identity and the system. It is a platform service, not a domain business service.

```text
verify Logto webhooks (signature, normalization, idempotency)
convert identity events into CloudEvents audit events
mini-program code2session, then issue a standard JWT
optionally act as an internal OIDC issuer (own issuer + JWKS)
no domain business logic
```

Tokens minted here must be standard JWTs verifiable by the Gateway via a known issuer + JWKS, never ad-hoc tokens.

### Gateway

```text
Go Gateway/BFF
```

Responsibilities:

```text
JWT verification through Logto JWKS, with cached JWKS
rate limiting
structured access logging
request_id generation
trace context propagation
trusted identity header injection
route forwarding
lightweight aggregation
unified error response
API entry-level audit events
```

Gateway readiness must not hard-depend on Logto being reachable. If cached JWKS is available, a temporary Logto outage should put Gateway in degraded mode, not make it not-ready.

### Client Authentication

Whether login passes through the Gateway depends on the client. The Gateway accepts multiple credential forms and normalizes them into the same identity headers, so business services stay credential-agnostic.

```text
Web / Admin (SPA):
  BFF Token Handler, Gateway does OIDC code exchange
  token in HttpOnly + Secure cookie
  login through Gateway: yes
  API credential: cookie

WeChat Mini Program:
  server-side code2session, then issue a standard JWT (not an ad-hoc token)
  token in mini-program storage (no HttpOnly cookie)
  login through Gateway: yes
  API credential: Bearer header

Mobile / Native App:
  OIDC + PKCE directly with Logto
  token in OS secure storage
  login through Gateway: no (direct to Logto); API calls go through Gateway
  API credential: Bearer header
```

Gateway dual-credential rule:

```text
cookie -> validate session
Authorization header -> validate JWT via cached JWKS
both -> inject the same trusted identity headers
```

Implementation notes:

```text
BFF auth (OIDC client + cookie) is needed only for Web
a mini-program login endpoint does code2session and issues a standard JWT
mobile login does not involve the Gateway
verify Logto WeChat mini-program support; otherwise implement code2session in auth-service
```

Token issuance rule (no ad-hoc tokens):

```text
every token the Gateway accepts must be a standard JWT with a known issuer + published JWKS
mini-program tokens must be issued via Logto/OIDC, or via auth-service as an internal OIDC issuer
the Gateway verifies all clients through the same issuer + JWKS path
do not invent custom per-endpoint tokens the Gateway cannot verify uniformly
```

### Business Services

Language split:

```text
Go platform: gateway, auth-service, audit-service, user-service, file-service, worker
TypeScript domain: node/apps/business-service (NestJS + Drizzle) — primary CRUD under /v1/business/*
Python: heavy AI/data pipelines (batch, training, embeddings at scale)
```

NestJS runs as one HTTP microservice behind the Gateway. It must not parse end-user JWTs,
must not implement a second BFF for the admin SPA, and must not replace Gateway auth.
Next.js Route Handlers are not the primary business backend.

### Web Tier

```text
node/apps/site: Next.js SSR/SEO; Server fetches Gateway /v1; no next-auth; no JWT in browser JS
node/apps/admin: Vite + React SPA; TanStack Query; credentials: include; types from @ting/api-types
Both use the same Web cookie model through Gateway BFF.
Admin static assets may be served without a Node runtime.
```

### Type Sharing

```text
platform-contracts OpenAPI → generate node/packages/api-types (@ting/api-types)
Nest business-service + node/apps/admin + node/apps/site import api-types
Drizzle DB types stay server-side only; map to API DTOs in Nest
proto/buf for common types (identity, error, audit) and V2 internal gRPC
Do not use tRPC or framework-only type tunnels as the cross-client contract.
```

### End-to-End Chain (V1)

```text
Browser → Nginx → Go Gateway (cookie/Bearer → identity headers)
  → /v1/business/* → Nest + Drizzle + outbox
  → /v1/users/*, /v1/files/* → Go services
  → /, SSR paths → Next (internal)
  → /admin/* → Vite static
V1 internal traffic: HTTP. V2 hot paths: gRPC + identity metadata (see identity.proto).
Full route table and phased rollout: docs/ARCHITECTURE.md § End-to-End Request Chain.
```

### Data And Infra

```text
PostgreSQL
Redis
RabbitMQ
S3-compatible storage
```

V1 PostgreSQL layout can be one instance with multiple databases or schemas:

```text
logto_db
app_db
audit_db
```

**Deployment rule:** stateful data stores are not co-deployed with application
services in production. Use managed RDS/Redis/MQ/OSS, or for local dev install
PostgreSQL/Redis/RabbitMQ natively on the host (preferred). Optional Docker infra:
`deploy/docker-compose.infra.yml`. Application services live in
`deploy/docker-compose.yml` and connect via env (`POSTGRES_HOST`, `REDIS_ADDR`, …).

Connection libraries live in `go/pkg/db`, `go/pkg/cache`, `go/pkg/mq`, `go/pkg/storage`.
Services load `.env` via `config.LoadEnvFile()` for local dev; production injects
env at deploy time. Cloud placeholder values (e.g. `rm-xxx`, `*-placeholder`) skip
real connections until replaced.

```text
Local dev:  install PG/Redis locally → cp .env.example .env → make run-gateway / make run-user-service
Optional:   make up-infra (Docker infra) + POSTGRES_HOST=localhost
Production: make up-apps + replace cloud placeholders in .env
```

## Required Service Behavior

Every service must expose:

```text
/healthz
/readyz
/metrics
```

`/healthz` is process liveness only.

`/readyz` checks critical dependencies:

```text
PostgreSQL reachable
Redis reachable if used
RabbitMQ reachable if used
required config loaded
object storage reachable for file-service if required
```

Every service must:

```text
log JSON to stdout
include request_id and trace_id in logs
propagate traceparent
support OpenTelemetry instrumentation
emit CloudEvents-compatible audit events for audit-worthy actions
return unified error responses
follow 12-Factor principles
```

## P0 Decisions To Preserve

Do not defer these design boundaries.

### Multi-Tenant Context

Reserve tenant context from V1:

```text
tenant_id
org_id optional
workspace_id optional
```

Tables that may become tenant-scoped should include `tenant_id`.

### Token Revocation

Access tokens should be short-lived. Refresh tokens are managed by Logto.

For high-risk operations, Gateway or business services may check session state or revocation state.

Revocation/session blocklists should live in Redis, keyed by subject or session id, so Gateway can check them with low latency on sensitive paths.

### Service-to-Service Trust

Identity propagation is not enough; callees must also trust callers.

V1 trust model:

```text
network isolation
shared internal service token or signed internal header
business services reject identity headers from untrusted callers
```

Later trust model:

```text
mutual TLS between services
service identity through mesh or platform layer
```

### Secrets

Rules:

```text
no secrets in source code
no secrets in images
no secrets in logs
production secrets injected at deploy time
rotation must be possible
```

### Backups

V1 requires backups for:

```text
app_db
logto_db
audit_db
object storage
deployment configuration
```

Backups must have restore verification.

### Minimal CI

V1 CI must include:

```text
lint
test
build
container image build
image scan
push image to ACR
version deployment artifacts
```

## Audit Rules

Audit is separate from ordinary logs.

Audit events should use CloudEvents-style envelopes.

Examples of audit event types:

```text
user.login.success
user.login.failed
user.phone.bind
user.role.update
resource.delete
payment.refund
admin.operation
```

V1 path:

```text
service -> outbox table in the same DB transaction as the business write
        -> async dispatch -> Audit Service -> PostgreSQL audit_db
```

Future path:

```text
service -> outbox -> RabbitMQ -> Audit Service -> PostgreSQL audit_db
```

Do not implement audit as best-effort direct HTTP writes on the hot path. Use the Transactional Outbox pattern so a successful business change implies the audit event was recorded.

Audit ownership and delivery (outbox applies only to source 3):

```text
1. identity events (user.login.success/failed, user.logout)
   source: Logto webhook -> auth-service; mini-program login from auth-service code2session
   why not Gateway: mobile login bypasses the Gateway
   delivery: auth-service (its own outbox if it persists) -> Audit Service

2. entry events (api.access.denied, api.rate_limited, api.token.invalid)
   source: Gateway (no business DB write)
   delivery: async, direct or via queue, no outbox

3. domain events (resource.delete, payment.refund, ...)
   source: business services, tied to a business write
   delivery: Transactional Outbox in the same transaction -> dispatch -> Audit Service
```

The CloudEvents `source` field must distinguish idp / gateway / service origins.

## Observability Rules

Use OpenTelemetry as the observability standard.

Use ECS-style fields for structured logs.

Recommended backends:

```text
logs: Loki, Aliyun SLS, or Elastic
metrics: Prometheus
traces: Tempo or Jaeger
dashboards: Grafana
```

All service-to-service calls must propagate:

```text
traceparent
baggage optional
X-Request-Id
X-User-Id
X-Tenant-Id
X-Roles
X-Scopes
X-Auth-Subject
```

## API And Contract Rules

Use `platform-contracts` as the source of truth.

Expected contents:

```text
proto/
buf.yaml
buf.gen.yaml
schemas/logging.schema.json
schemas/audit-event.schema.json
schemas/error-response.schema.json
schemas/identity-context.schema.json
docs/metrics.md
docs/tracing.md
```

Use buf for:

```text
lint
breaking-change checks
multi-language code generation
```

External APIs should be versioned with a `/v1` path prefix. Breaking API and Protobuf contract changes must be caught in CI before deployment.

### Schema Migrations

Database schema changes must use a migration tool such as golang-migrate or Atlas. Migration files are versioned in Git and applied through CI or controlled deployment steps, not by manually changing production databases.

## What AI Agents Should Avoid

Do not:

```text
put business authorization into Gateway
make services parse user JWTs unless explicitly required for an exception
add next-auth, a Nest BFF, or Next Route Handlers that duplicate Gateway OIDC/cookie auth
create language-specific conventions without updating platform-contracts
use tRPC or Eden-only types as the cross-client API contract instead of platform-contracts
write plain-text logs or file-based service logs
mix audit records into ordinary logs only
write audit events directly to Audit Service without outbox when tied to business writes
skip tenant_id on data likely to become tenant-scoped
store secrets in source code, images, or committed env files
add infrastructure like Dapr, Service Mesh, or Backstage in V1 without explicit approval
```

## What AI Agents Should Prefer

Prefer:

```text
small Go services following the service template
node/ pnpm monorepo for all Node/TypeScript
NestJS business-service for domain CRUD (Drizzle, outbox, identity headers)
@ting/api-types generated from platform-contracts OpenAPI
12-Factor app behavior
JSON stdout logs
OpenTelemetry instrumentation
CloudEvents audit emission
Protobuf + buf for cross-language contracts
PostgreSQL for durable data
Redis for cache/ephemeral state
RabbitMQ for async work and retries
S3-compatible storage for files
Docker Compose for V1 deployment
Kubernetes only in later platform stages
```

## Service Template Checklist

When creating a new service, include:

```text
config loading from environment
structured logger
request_id middleware
trace propagation
metrics endpoint
healthz endpoint
readyz endpoint with real dependency checks
unified error responses
identity context extraction
audit event client
database migrations if the service owns data
outbox support for audit-worthy business writes
Dockerfile
minimal CI job
README with service responsibilities
```

## Current Evolution Strategy

V1:

```text
single ECS
Docker Compose (apps separate from data infra)
Nginx
Go Gateway/BFF
Logto
auth-service (IdP integration)
Go platform services (user, file, audit, worker)
node/ pnpm monorepo (apps + packages)
NestJS business-service + Drizzle (node/apps/business-service)
node/apps/site (Next.js) + node/apps/admin (Vite)
node/packages/api-types from platform-contracts OpenAPI
PostgreSQL (managed or local infra compose for dev)
Redis
RabbitMQ
S3/OSS
Audit Service
OpenTelemetry baseline
platform-contracts (OpenAPI + proto)
Go service template + Nest business-service baseline
minimal CI and backups
```

V2:

```text
gRPC + Protobuf for hot internal APIs; external /v1 stays HTTP
buf checks and TypeScript proto generation in CI
full observability pipeline (gRPC interceptors)
async audit through RabbitMQ
Python for heavy AI/data pipelines
Nest AI modules (AI SDK / Mastra) for product AI HTTP APIs
multi-language service templates
```

V3:

```text
Kubernetes / ACK
managed databases and cache
Service Mesh or Dapr if justified
mTLS
progressive delivery
Backstage or internal developer portal
policy-as-code
```

