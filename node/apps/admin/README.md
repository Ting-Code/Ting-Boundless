# @ting/admin

Vite + React admin SPA (`basename` `/admin`). TanStack Query → Gateway `/v1` with
`credentials: include`. Types from `@ting/api-types`.

## Run (dev)

Gateway, Redis, and business-service must be running. Vite proxies `/v1` and `/sign-in` to Gateway.

```bash
# repo root
make run-gateway
make run-business
make run-admin
# http://localhost:5173/admin/items
```

**Dev login (no Logto):** click「开发环境登录」or open `/sign-in/dev?return_to=/admin/items`.
Requires `GATEWAY_BFF_DEV_LOGIN=true` and Redis in `.env`.

See [`docs/E2E_ADMIN.md`](../../docs/E2E_ADMIN.md) and `make e2e-admin` for scripted smoke.

## Pages (V1)

| Route | Description |
|-------|-------------|
| `/admin/items` | List + create business items (`/v1/business/items`) |

Unauthenticated requests redirect to Gateway `/sign-in`.
