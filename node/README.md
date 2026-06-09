# node/ — TypeScript monorepo (pnpm)

All Node.js and TypeScript code lives under this directory: Nest domain API,
frontends, and shared packages. Go platform services live in `go/services/`.

## Layout

```text
node/
  apps/
    business-service/   NestJS + Drizzle — /v1/business/*
    admin/              Vite + React SPA — /admin
    site/               Next.js SSR public site
  packages/
    logger/             ECS JSON logger (@ting/logger); Nest LoggingInterceptor
    api-types/          Generated types from platform-contracts OpenAPI
    api-client/         Optional orval + TanStack Query hooks
```

## Prerequisites

- **Node.js 20+** (repo ships `node/.nvmrc`; current shell may need `nvm use` or install from [nodejs.org](https://nodejs.org/))
- [pnpm](https://pnpm.io/installation) 9+ (`corepack enable && corepack prepare pnpm@9.15.0 --activate`)

## Commands (from this directory)

```bash
pnpm install
pnpm dev:business    # Nest, default :3005
pnpm dev:admin       # Vite admin
pnpm dev:site        # Next.js
pnpm generate:api-types
pnpm build           # all workspaces that define build
```

From repo root: `make node-install` or `cd node && pnpm install`.

## Conventions

- Workspace packages use the `@ting/*` scope.
- Apps call Gateway `/v1` only; no duplicate OIDC in Nest or Next.
- Types come from `@ting/api-types` (generated from `platform-contracts/`).
- Drizzle schema types stay inside `business-service`; map to API DTOs at the boundary.

Gateway proxies `/v1/business/*` to Nest (`BUSINESS_SERVICE_URL`, default `http://127.0.0.1:3005`).
The Go `business-service` stub has been removed; domain CRUD is implemented only under
`node/apps/business-service`.

See `docs/ARCHITECTURE.md` § End-to-End Request Chain.
