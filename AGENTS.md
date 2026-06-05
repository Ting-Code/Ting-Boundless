# AGENTS.md

Entry point for AI coding agents working in this repository.

## Read first

1. `docs/AI_CONTEXT.md` — compact, durable architecture rules. Read before proposing
   plans, creating services, touching auth, observability, or cross-language APIs.
2. `docs/ARCHITECTURE.md` — full V1 architecture and rationale.
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
platform-contracts/  cross-language source of truth (proto + buf, JSON schemas)
pkg/                 shared Go libraries (thin SDK over the contracts)
services/            deployable services (one folder each)
deploy/              docker-compose (apps + infra split), nginx, otel collector
docs/                architecture docs
.cursor/             rules + skills
```

## Create a new Go service

Use the `new-go-service` skill (`.cursor/skills/new-go-service/`). It encodes the
required structure and baseline (health, logging, identity, audit, Dockerfile).

## Build / verify

```
make build   # go build ./...
make vet
make test
```
