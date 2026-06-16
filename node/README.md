# node/ — TypeScript monorepo (pnpm)

All Node.js and TypeScript code lives under this directory: Nest domain API,
frontends, and shared packages. Go platform services live in `go/services/`.

## V1 scope (web admin first)

**In focus** — cookie BFF via Gateway, no end-user JWT in browser JS:

| Piece | Role |
|-------|------|
| `@ting/admin` | Vite + React 后台 SPA (`/admin`) |
| `@ting/business-service` | Nest 域 API `/v1/business/*` |
| `@ting/api` | OpenAPI 类型 + 路径 + `apiFetch`（business / users / files / audit） |
| `@ting/logger` | Nest 日志与 OTLP |

**Deferred** (platform Go 能力可保留，Node 侧暂不投入):

- 小程序 / 原生 App 的 TS 客户端、`auth.v1` OpenAPI
- `@ting/site` 公开站打磨（脚手架在，非当前主线）
- 移动端 OIDC 集成文档仅作平台参考（`docs/MOBILE_AUTH.md`）

## Layout

```text
node/
  apps/
    business-service/   NestJS + Drizzle — /v1/business/*
    admin/              Vite + React SPA — /admin  ← V1 主线
    site/               Next.js SSR（延后）
  packages/
    logger/             ECS JSON logger (@ting/logger)
    api/                @ting/api — paths, types, request helpers
```

## Prerequisites

- **Node.js 20+** (repo ships `node/.nvmrc`)
- [pnpm](https://pnpm.io/installation) 9+

## Commands (from this directory)

```bash
pnpm install
pnpm dev:business    # Nest, default :3005
pnpm dev:admin       # Vite admin
pnpm generate:api
pnpm build           # logger + api + business-service + admin
pnpm typecheck       # tsc --noEmit for admin + business-service
pnpm test            # @ting/api vitest
```

From repo root: `make run-admin`, `make run-business`, `make generate-api`.

`pnpm dev:site` / `make run-site` 存在，但不在 V1 后台主线内。

## Conventions

- Workspace packages use the `@ting/*` scope.
- Apps call Gateway `/v1` only; no duplicate OIDC in Nest or admin.
- Paths, types, and `apiFetch` come from `@ting/api` (`platform-contracts/openapi/`).
- Drizzle schema types stay inside `business-service`; map to API DTOs at the boundary.

Gateway proxies `/v1/business/*` → Nest (`BUSINESS_SERVICE_URL`, default `http://127.0.0.1:3005`).

See `docs/ARCHITECTURE.md` § End-to-End Request Chain.
