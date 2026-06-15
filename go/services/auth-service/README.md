# auth-service

IdP integration layer (platform service, no domain logic).

- Receives + verifies Logto webhooks; converts identity events → audit events.
- WeChat mini-program login: `code2session` → issues a **standard RS256 JWT**.
- Internal OIDC issuer: `AUTH_OIDC_ISSUER` + `GET /v1/auth/jwks` (Gateway trusts via `AUTH_JWKS_URL`).

## Endpoints

| Method | Path | Auth |
|--------|------|------|
| POST | `/v1/auth/miniprogram/login` | Gateway anon whitelist |
| GET | `/v1/auth/jwks` | Gateway anon whitelist |
| POST | `/internal/webhooks/logto` | `X-Internal-Token` + Logto HMAC signature |
| POST | `/internal/identity/resolve` | `X-Internal-Token` |

### Logto webhook

Logto Console → **Webhooks** → endpoint (via Gateway/nginx internal route):

`POST https://<api>/internal/webhooks/logto` with header `X-Internal-Token`.

Subscribe to at least: `PostSignIn`, `PostRegister`, `Identifier.Lockout`.

Set `LOGTO_WEBHOOK_SIGNING_KEY` from the webhook detail page. Local dev without Logto:
`LOGTO_WEBHOOK_SKIP_VERIFY=true` (never in production).

Mapped audit types: `user.login.success`, `user.register.success`, `user.login.failed`, etc.
Delivered to `AUDIT_SERVICE_URL` (`POST /internal/audit/events`).

### Mini-program login

```http
POST /v1/auth/miniprogram/login
Content-Type: application/json

{"code": "<wx.login() code>"}
```

Response:

```json
{
  "access_token": "<JWT>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user_id": "<platform user id>"
}
```

Use `Authorization: Bearer <access_token>` on subsequent API calls through the Gateway.

## Local dev (no WeChat app)

Set `WECHAT_MOCK_MODE=true`. Login codes map to mock identities:

| `code` | openid | unionid |
|--------|--------|---------|
| `abc` | `mock_abc` | — |
| `app_a\|shared` | `mock_app_a` | `union_shared` |

Use `openid|union` to test **unionid binding** across mini-programs (same union → same `user_id`).

Gateway must include:

```
AUTH_OIDC_ISSUER=http://127.0.0.1:8084/oidc
AUTH_JWKS_URL=http://127.0.0.1:8084/v1/auth/jwks
```

Without `AUTH_JWT_PRIVATE_KEY_PEM`, auth-service generates an ephemeral RSA key (tokens invalid after restart).

## Env

`HTTP_ADDR`, `OIDC_AUDIENCE`, `AUTH_OIDC_ISSUER`, `AUTH_JWKS_URL`, `AUTH_JWT_PRIVATE_KEY_PEM`,
`AUTH_JWT_ACCESS_TTL_SECONDS`, `WECHAT_APP_ID`, `WECHAT_APP_SECRET`, `WECHAT_MOCK_MODE`,
`LOGTO_WEBHOOK_SIGNING_KEY`, `LOGTO_WEBHOOK_SKIP_VERIFY`, `AUDIT_SERVICE_URL`,
`POSTGRES_*`, `REDIS_ADDR`, `INTERNAL_API_TOKEN`.

## Migrations

`user_identities`, `webhook_deliveries` in `app_db` (`go/migrations/auth-service/`). Applied on startup and via `make migrate`.
