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
| `deploy/` | docker-compose, nginx, OpenTelemetry collector |
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

## Quick start (dev)

```bash
cp .env.example .env      # fill in secrets; never commit .env
make build                # compile everything
make run-gateway          # run a single service
make up                   # start the full stack (requires docker compose)
```

## Conventions

Every service: `/healthz`, `/readyz`, `/metrics`, JSON stdout logs (ECS-style),
`traceparent` propagation, unified error responses, 12-Factor config via env.
