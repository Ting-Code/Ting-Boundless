# @ting/api

Single package for **Gateway `/v1` contract** in TypeScript:

| Module | Purpose |
|--------|---------|
| `paths` | Web 后台 URL 常量（business / users / files / audit） |
| `types/*` | OpenAPI-generated named exports per domain |
| `request` | `apiFetch`, `apiUpload`, `ApiError`, `isApiError` (browser cookie or SSR `baseUrl`) |

Consumed by `@ting/business-service` (types only), `@ting/admin`, and `@ting/site`.

## Regenerate types

```bash
make generate-api
# or
cd node && pnpm generate:api
```

Requires Node ≥20. Edit specs under `platform-contracts/openapi/` first.

## Tests

```bash
cd node && pnpm --filter @ting/api test
```

Vitest covers path helpers (`eventsQuery`, `listQuery`) and `isApiError` / `resolveApiUrl`.

## OpenAPI specs

| File | Domain |
|------|--------|
| `common.v1.yaml` | Shared `ErrorEnvelope` |
| `business.v1.yaml` | Nest `/v1/business/*` |
| `users.v1.yaml` | Go user-service `/v1/users/*` |
| `files.v1.yaml` | Go file-service `/v1/files/*` |
| `audit.v1.yaml` | Go audit-service `/v1/audit/events` |

## Layout

```text
src/
  generated/*.v1.ts          openapi-typescript output (gitignored)
  types/*.ts                 friendly re-exports from generated schemas
  paths.ts                   route constants
  request.ts                 fetch wrappers
```
