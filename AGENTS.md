# AGENTS.md

Entry point for AI coding agents working in this repository.

## Read first

1. `docs/AI_CONTEXT.md` — compact, durable architecture rules. Read before proposing
   plans, creating services, touching auth, observability, or cross-language APIs.
2. `docs/ARCHITECTURE.md` — full V1 architecture and rationale.
4. `docs/E2E_ADMIN.md` — Gateway cookie → Admin → business-service local flow.
3. `.cursor/rules/` — enforced conventions (auto-applied).

## Non-negotiables (summary; AI_CONTEXT.md is authoritative)

- Gateway decides **who** the caller is; business services decide **what** they may do.
- Business services do not parse end-user JWTs; they trust identity headers injected by the Gateway.
- The Gateway strips client-supplied identity headers before injecting trusted ones.
- Cross-language behavior is defined in `platform-contracts/` first, language SDKs second.
- Every service exposes `/healthz`, `/readyz`, `/metrics`; logs JSON to stdout (ECS-style); propagates `traceparent`.
- Audit events are separate from logs; domain audit uses the Transactional Outbox pattern.
- Secrets never enter source, images, or logs.

## Layout

```
platform-contracts/  cross-language source of truth (OpenAPI, proto + buf, JSON schemas)
go/                  Go monorepo (module root: go/go.mod)
  pkg/               shared Go libraries (thin SDK over the contracts)
  services/          Go platform deployables (gateway, auth, user, file, audit, worker)
  cmd/               dev-jwt, migrate
  migrations/        SQL migrations per Go service
node/                pnpm monorepo — all Node/TypeScript (apps + packages)
  apps/              business-service (Nest), admin (Vite), site (Next.js)
  packages/          api-types, api-client (from OpenAPI)
deploy/              docker-compose (apps + infra split), nginx, otel collector
docs/                architecture docs (full chain in ARCHITECTURE.md § End-to-End Request Chain)
.cursor/             rules + skills
```

## Language split

- **Go** (`go/services/`): gateway, auth-service, audit-service, user-service, file-service, worker — platform logic only once.
- **TypeScript** (`node/`): Nest `@ting/business-service` for domain CRUD under `/v1/business/*`; `@ting/admin` and `@ting/site` for frontends; share `@ting/api-types`; never duplicate OIDC in Nest or Next.

Domain CRUD belongs in `node/apps/business-service`, not new Go services, unless explicitly approved.

## Create a new Go service

Use the `new-go-service` skill (`.cursor/skills/new-go-service/`). It encodes the
required structure and baseline (health, logging, identity, audit, Dockerfile).

## Build / verify

```
make build        # cd go && go build ./...
make vet
make test
make run-gateway   # cd go && go run ./services/gateway
make run-business  # Nest business-service :3005
make run-admin     # Vite admin :5173
make node-install  # pnpm install in node/
make node-build    # build api-types + business-service
```
