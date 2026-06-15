# E2E: WeChat Mini-Program Login (Mock) → Bearer JWT → user-service

Local dev uses `WECHAT_MOCK_MODE=true` — no real WeChat app credentials required.

## Prerequisites

| Component | Command / check |
|-----------|-----------------|
| PostgreSQL `app_db` | `scripts/setup-local.ps1` or `make migrate` |
| `.env` | `WECHAT_MOCK_MODE=true`, `AUTH_OIDC_ISSUER`, `AUTH_JWKS_URL` (see `.env.example`) |
| auth-service | `make run-auth` (`:8084`) |
| user-service | `make run-user` (`:8081`) |
| Gateway | `make run-gateway` (`:8080`) — must load `AUTH_JWKS_URL` |

Logto and Redis are **not** required for this path.

## Mock code format

When `WECHAT_MOCK_MODE=true`, the login `code` field supports:

| Code | Mock openid | Mock unionid |
|------|-------------|--------------|
| `abc123` | `mock_abc123` | (empty) |
| `app_a\|shared` | `mock_app_a` | `union_shared` |

Use `openid|union` to test cross–mini-program binding: two different openids with the same union suffix resolve to one `user_id`.

## Automated smoke

```powershell
powershell -ExecutionPolicy Bypass -File scripts/e2e-miniprogram-gateway.ps1
```

Or:

```bash
make e2e-miniprogram
```

Flow:

1. `GET /v1/auth/jwks` via Gateway (anon whitelist)
2. `POST /v1/auth/miniprogram/login` with `e2e_mp_a|e2e_union_1`
3. Same union, different openid: `e2e_mp_b|e2e_union_1` → same `user_id`
4. `GET /v1/users/me` with `Authorization: Bearer <access_token>`

## Manual curl

```bash
curl -s -X POST http://127.0.0.1:8080/v1/auth/miniprogram/login \
  -H "Content-Type: application/json" \
  -d '{"code":"my_openid|my_union"}'

curl -s http://127.0.0.1:8080/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

## Production

- Set real `WECHAT_APP_ID` / `WECHAT_APP_SECRET`, disable `WECHAT_MOCK_MODE`
- Configure fixed `AUTH_JWT_PRIVATE_KEY_PEM`
- WeChat returns `unionid` when the mini-program is bound to the same WeChat Open Platform account as your App / other mini-programs
