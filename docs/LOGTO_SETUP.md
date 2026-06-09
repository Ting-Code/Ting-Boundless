# Logto setup (local dev)

Gateway BFF routes (`/sign-in`, `/callback`, `/sign-out`) use Logto as the OIDC IdP.
Business services never talk to Logto directly; they trust identity headers from the Gateway.

## Prerequisites

- PostgreSQL with `logto_db` (see `deploy/postgres/setup-local.sql`)
- Redis (BFF session store)
- Node **22.x** via nvm (`nvm install 22 && nvm use 22`)
- pnpm 10 (`npm install -g pnpm@10 --force` if corepack fails on Windows)
- Ports **3001** (OIDC) and **3002** (Admin Console) free

## 1. Start Logto

**Windows (recommended — prebuilt release via CLI):**

```powershell
nvm install 22.14.0
nvm use 22.14.0
powershell -ExecutionPolicy Bypass -File scripts/start-logto.ps1 -SeedOnly
powershell -ExecutionPolicy Bypass -File scripts/start-logto.ps1
```

First run downloads ~257MB from GitHub; allow several minutes on slow networks.

**Manual (Git Bash):**

```bash
export PATH="$APPDATA/nvm/v22.14.0:$PATH"
npx @logto/cli init -p deploy/logto \
  --db-url "postgresql://ting:change-me@127.0.0.1:5432/logto_db" --dapc
cd deploy/logto && npm start
```

Do **not** build from `deploy/logto-src` on Windows (Vite/Sass path issues). Use the CLI prebuilt install or Docker.

- OIDC issuer: `http://127.0.0.1:3001/oidc`
- Admin Console: `http://127.0.0.1:3002`

On first visit to Admin Console, create the admin account.

## 2. Configure Logto application

**Automated (local dev):**

```powershell
powershell -ExecutionPolicy Bypass -File scripts/configure-logto-local.ps1 -UpdateEnv
```

Creates API resource `ting-boundless`, Traditional Web App `Gateway BFF`, and writes
`OIDC_CLIENT_ID` / `OIDC_CLIENT_SECRET` to `.env`.

**Manual** — In Admin Console (`http://127.0.0.1:3002`):

### API resource

1. **Applications → API resources → Create**
2. Name: `Ting Boundless API`
3. API identifier: `https://api.ting-boundless.local` (absolute URI; must match `OIDC_AUDIENCE` / `OIDC_RESOURCE` in `.env`)

### Traditional Web App (Gateway BFF)

1. **Applications → Create application → Traditional Web App**
2. Name: `Gateway BFF`
3. Redirect URIs: `http://127.0.0.1:8080/callback`
4. Copy **App ID** → `OIDC_CLIENT_ID`
5. Copy **App secret** → `OIDC_CLIENT_SECRET`

## 3. Update `.env` and restart Gateway

```env
OIDC_ISSUER=http://127.0.0.1:3001/oidc
OIDC_JWKS_URL=http://127.0.0.1:3001/oidc/jwks
OIDC_AUDIENCE=https://api.ting-boundless.local
OIDC_RESOURCE=https://api.ting-boundless.local
OIDC_CLIENT_ID=<from Logto console>
OIDC_CLIENT_SECRET=<from Logto console>
GATEWAY_PUBLIC_URL=http://127.0.0.1:8080
GATEWAY_BFF_DEV_LOGIN=false
```

```bash
make run-gateway
```

## 4. Test real `/sign-in`

1. Build admin SPA once: `cd node && pnpm --filter @ting/admin build` (Gateway serves it at `/admin`)
2. Open `http://127.0.0.1:8080/sign-in?return_to=/admin/items`
3. Log in at Logto → redirect to `/callback` → session cookie set
4. Verify:

```bash
curl -b cookies.txt -c cookies.txt -L "http://127.0.0.1:8080/sign-in?return_to=/"
curl -b cookies.txt http://127.0.0.1:8080/v1/users/me
```

## Docker (optional)

If Docker Desktop is installed:

```bash
docker compose -f deploy/docker-compose.logto.yml up -d
```

Uses host Postgres via `host.docker.internal`.

## Logto Cloud (no self-host)

Point `OIDC_ISSUER`, `OIDC_JWKS_URL`, and client credentials at your cloud tenant.
Same BFF redirect URI and API resource identifier rules apply.

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `oidc_not_configured` on `/sign-in` | Set `OIDC_CLIENT_ID` + `OIDC_CLIENT_SECRET` |
| `token_exchange_failed` | Check redirect URI matches Logto app; ensure API resource exists |
| `invalid_target: resource indicator must be an absolute URI` | Logto requires RFC 8707 absolute URI (e.g. `https://api.ting-boundless.local`), not a bare name |
| `audience mismatch` | API identifier in Logto must equal `OIDC_AUDIENCE` |
| Node v16 errors | Use nvm 22: `export PATH="$APPDATA/nvm/v22.13.0:$PATH"` |
| corepack signature error | `npm install -g pnpm@10 --force` with Node 22 |
| `permission denied to create role` on seed | Logto seed creates DB roles; grant `CREATEROLE` to your app user: `ALTER ROLE ting CREATEROLE;` (as `postgres` superuser). `setup-local.sql` includes this for new installs. |
| Windows prebuilt tarball symlink errors | Use `deploy/logto-src` instead: `pnpm build` in `packages/core`, then `pnpm run cli db seed` and `pnpm start` from repo root `deploy/logto-src/`. |
