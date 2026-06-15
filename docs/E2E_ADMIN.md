# E2E: Gateway dev cookie login → business + audit (Web admin path)

Local dev without Logto: **Gateway dev BFF** (`/sign-in/dev`) + Redis sessions +
HS256 dev JWT (`GATEWAY_DEV_JWT_SECRET`).

## Prerequisites

| Component | Command / check |
|-----------|-----------------|
| PostgreSQL `app_db` + `audit_db` | `make migrate` |
| Redis | `localhost:6379` |
| `.env` | `GATEWAY_BFF_DEV_LOGIN=true`, `GATEWAY_DEV_JWT_SECRET`, `BUSINESS_SERVICE_URL=http://127.0.0.1:3005` |
| Gateway | `make run-gateway` (`:8080`) |
| Nest business | `make run-business` (`:3005`) |
| audit-service | `make run-audit` (`:8085`) — for `/admin/audit` smoke |

Logto is **not** required when `GATEWAY_BFF_DEV_LOGIN=true`.

Dev sign-in defaults to roles **`user,admin`** so the audit API (`GET /v1/audit/events`) works.
Override: `/sign-in/dev?roles=user` (no audit) or `roles=admin` only.

## Automated smoke (no browser)

```powershell
powershell -ExecutionPolicy Bypass -File scripts/e2e-admin-gateway.ps1
```

Or: `make e2e-admin`

Flow:

1. `GET /v1/business/ping`
2. `GET /sign-in/dev` → session cookie
3. `GET /v1/business/me`
4. `POST` + `GET /v1/business/items`
5. `GET /v1/audit/events` (requires audit-service + `admin` role)

## Browser: Admin SPA

```bash
make run-gateway
make run-business
make run-audit      # optional but needed for 审计 page
make run-admin
```

1. Open `http://localhost:5173/admin/items`
2. **开发环境登录** → header shows user / tenant
3. Pages: **业务条目** · **文件** · **账户** · **审计** (`/admin/audit`)

`node/apps/admin/.env.development`: `VITE_DEV_LOGIN=true`, `VITE_SIGN_IN_PATH=/sign-in/dev`.

## Request path

```text
Browser (localhost:5173)
  → Vite proxy /v1, /sign-in/*
  → Go Gateway :8080
      cookie → Redis session → dev JWT (roles: user, admin)
      inject X-User-Id, X-Tenant-Id, X-User-Roles, ...
  → Nest :3005  /v1/business/*
  → Go   :8085  /v1/audit/*   (admin role)
  → Go   :8081  /v1/users/*
  → Go   :8083  /v1/files/*
```

## Production / Logto

**完整文档：** [BFF_LOGTO.md](./BFF_LOGTO.md)（Gateway 已实现 `/sign-in` `/callback` `/sign-out`；Admin 只需跳转 + `credentials: 'include'`）。

快速切换：

| 场景 | Admin env | Gateway `.env` |
|------|-----------|----------------|
| 本地 dev cookie | `VITE_DEV_LOGIN=true` | `GATEWAY_BFF_DEV_LOGIN=true` |
| 本地 / 生产 Logto | `VITE_DEV_LOGIN=false`, `VITE_SIGN_IN_PATH=/sign-in` | `OIDC_CLIENT_*`, `GATEWAY_BFF_DEV_LOGIN=false` |

Logto 运维账号需在 access token 中带 `roles` 含 `admin` 才能访问 `/admin/audit`（见 BFF_LOGTO § 角色）。

本地 Logto 安装：[LOGTO_SETUP.md](./LOGTO_SETUP.md)。

## OpenAPI

Web admin types: `make generate-api` after editing `platform-contracts/openapi/*.yaml`.
Lint: `make lint-openapi`.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `session_unavailable` | Start Redis |
| `dev_auth_unconfigured` | Set `GATEWAY_DEV_JWT_SECRET` in `.env` |
| 401 on `/v1/business/*` | Visit `/sign-in/dev` first |
| 403 on `/v1/audit/events` | Dev login needs `admin` role (default since V1 web focus); or `?roles=user,admin` |
| Audit page `database_unavailable` | `make run-audit` + `audit_db` migrated |
| Gateway → business 502 | `BUSINESS_SERVICE_URL=:3005` not `:3001` |
