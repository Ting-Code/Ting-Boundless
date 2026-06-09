# E2E: Gateway Cookie Login → Admin → business-service

Local dev path without Logto: **Gateway dev BFF** (`/sign-in/dev`) + Redis sessions +
HS256 dev JWT (`GATEWAY_DEV_JWT_SECRET`).

## Prerequisites

| Component | Command / check |
|-----------|-----------------|
| PostgreSQL `app_db` | `scripts/setup-local.ps1` |
| Redis | `localhost:6379` |
| `.env` | `cp .env.example .env` — ensure `GATEWAY_BFF_DEV_LOGIN=true`, `GATEWAY_DEV_JWT_SECRET`, `BUSINESS_SERVICE_URL=http://127.0.0.1:3005` |
| Go Gateway | `make run-gateway` (`:8080`) |
| Nest business | `make run-business` (`:3005`) |

Logto is **not** required when `GATEWAY_BFF_DEV_LOGIN=true`.

## Automated smoke (no browser)

```powershell
# Windows
powershell -ExecutionPolicy Bypass -File scripts/e2e-admin-gateway.ps1
```

Flow: `GET /sign-in/dev` → cookie → `GET /v1/business/me` → `POST/GET /v1/business/items`.

## Browser: Admin SPA

```bash
make run-gateway
make run-business
make run-admin
```

1. Open `http://localhost:5173/admin/items`
2. Click **开发环境登录** (or visit `http://localhost:5173/sign-in/dev?return_to=/admin/items`)
3. Header shows `e2e-user` / `dev-user` after redirect
4. Create a row in the table — requests go `Admin → Vite proxy → Gateway → Nest`

`node/apps/admin/.env.development` sets `VITE_DEV_LOGIN=true` and `VITE_SIGN_IN_PATH=/sign-in/dev`.

## Request path

```text
Browser (localhost:5173)
  → Vite proxy /v1, /sign-in/*
  → Go Gateway :8080
      cookie → Redis session → dev JWT verify
      inject X-User-Id, X-Tenant-Id, ...
  → Nest business-service :3005
      /v1/business/items
```

## Production / Logto

Set `VITE_DEV_LOGIN=false`, `VITE_SIGN_IN_PATH=/sign-in`, configure `OIDC_CLIENT_*`,
and use Gateway `/sign-in` → Logto → `/callback` (same cookie model).

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `session_unavailable` | Start Redis |
| `dev_auth_unconfigured` | Set `GATEWAY_DEV_JWT_SECRET` in `.env` |
| 401 on `/v1/business/*` | Visit `/sign-in/dev` first; check cookie `tb_session` |
| Empty item list after create | Tenant mismatch — dev login uses `tenant_id=dev-tenant` by default |
| Gateway → business 502 | `BUSINESS_SERVICE_URL` must be `:3005`, not `:3001` (Logto port) |
