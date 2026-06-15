# Mobile / Native App Authentication (OIDC + PKCE)

Native iOS and Android apps authenticate **directly with Logto** using standard OIDC
Authorization Code + **PKCE**. Login does **not** pass through the Gateway BFF cookie flow.
API calls use `Authorization: Bearer <access_token>` through the Gateway, which verifies
the JWT via cached JWKS and injects trusted identity headers downstream.

This matches `docs/AI_CONTEXT.md` (S-04) and `docs/ARCHITECTURE.md` § Client Auth Models.

## Credential model (vs Web / Mini Program)

| Client | Login path | API credential |
|--------|------------|----------------|
| Web / Admin SPA | Gateway BFF (`/sign-in` → Logto) | HttpOnly session cookie |
| WeChat Mini Program | Gateway → `POST /v1/auth/miniprogram/login` | Bearer (auth-service JWT) |
| **Mobile / Native** | **App → Logto (PKCE)** | **Bearer (Logto access token)** |

Business services never parse end-user JWTs; they only trust Gateway-injected headers.

## Architecture

```text
┌─────────────┐     OIDC + PKCE      ┌──────────┐
│  Native App │ ───────────────────► │  Logto   │
└─────────────┘   (no Gateway)       └──────────┘
       │
       │  Authorization: Bearer <access_token>
       ▼
┌─────────────┐   identity headers   ┌──────────────────┐
│   Gateway   │ ───────────────────► │ user / business  │
└─────────────┘                      └──────────────────┘
```

## Logto application setup

Prerequisites: Logto running locally or in cloud — see [LOGTO_SETUP.md](./LOGTO_SETUP.md).

### 1. API resource (shared with Web)

Ensure the API resource exists (same as Gateway BFF):

- **Identifier:** `https://api.ting-boundless.local` (must match `OIDC_AUDIENCE` in `.env`)

### 2. Native application

In Logto Admin Console → **Applications → Create → Native App**:

| Field | Example |
|-------|---------|
| Name | `Ting Boundless Mobile` |
| Redirect URI | `com.ting.boundless://callback` (custom URL scheme) |
| Post sign-out redirect URI | `com.ting.boundless://signout` (optional) |

Copy **App ID** into your mobile app config (`OIDC_CLIENT_ID`). Native apps are
**public clients** — no client secret; PKCE is mandatory.

### 3. Gateway environment (unchanged)

The Gateway verifies Logto-issued access tokens using the same JWKS as Web:

```env
OIDC_ISSUER=http://127.0.0.1:3001/oidc
OIDC_JWKS_URL=http://127.0.0.1:3001/oidc/jwks
OIDC_AUDIENCE=https://api.ting-boundless.local
```

Restart Gateway after Logto is configured. Mobile apps do **not** need `OIDC_CLIENT_SECRET`.

## OIDC + PKCE flow (app-side)

Use a standards-compliant library:

| Platform | Library |
|----------|---------|
| iOS / macOS | [AppAuth-iOS](https://github.com/openid/AppAuth-iOS) |
| Android | [AppAuth-Android](https://github.com/openid/AppAuth-Android) |
| React Native | `react-native-app-auth` (wraps AppAuth) |
| Flutter | `flutter_appauth` |

### Discovery

Fetch OpenID Provider Metadata from:

```text
{OIDC_ISSUER}/.well-known/openid-configuration
```

Local dev: `http://127.0.0.1:3001/oidc/.well-known/openid-configuration`

### Authorization request

```text
response_type=code
client_id=<native_app_id>
redirect_uri=com.ting.boundless://callback
scope=openid profile email offline_access
code_challenge=<S256 challenge>
code_challenge_method=S256
resource=https://api.ting-boundless.local
```

`resource` (RFC 8707) must be the **absolute URI** of the API resource in Logto so the
access token carries the correct `aud` for Gateway verification.

### Token exchange

Exchange `code` + `code_verifier` at Logto's token endpoint. Store tokens in OS secure
storage only:

- iOS: Keychain (`kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly` or stricter)
- Android: EncryptedSharedPreferences / Keystore-backed storage

Never log tokens or persist them in plain UserDefaults / SharedPreferences.

### Refresh

Use the refresh token (when `offline_access` is granted) before access token expiry.
On refresh failure (revoked session), restart the authorization flow.

### Sign-out

1. Delete tokens from secure storage.
2. Optionally open Logto `end_session_endpoint` with `id_token_hint` for server-side logout.
3. Sensitive API paths may also check Redis revocation (`GATEWAY_SENSITIVE_PREFIXES`); sign-out
   from Web BFF revokes the **session** only — mobile apps manage their own token lifecycle.

## API calls

All REST calls go through the Gateway (or nginx → Gateway in production):

```http
GET /v1/users/me HTTP/1.1
Host: api.example.com
Authorization: Bearer <logto_access_token>
Accept: application/json
```

Same OpenAPI contract as Web and mini-program (`platform-contracts/openapi/`).
Use `@ting/api` if you add a TypeScript/React Native shared package later.

### Base URLs

| Environment | API base |
|-------------|----------|
| Local (host-native) | `http://127.0.0.1:8080` |
| Local (nginx) | `http://127.0.0.1` |
| Production | `https://api.<your-domain>` |

Do not call `user-service:8081` or `business-service:3005` directly from the app.

## Local dev without a native app

### Option A — Logto token (full path)

1. Complete [LOGTO_SETUP.md](./LOGTO_SETUP.md).
2. Use Logto's **Try it** in the Native app settings, or a desktop OIDC client with the same
   redirect URI registered for debugging (`http://127.0.0.1:3006/callback` as an extra redirect
   URI in dev only).
3. Call Gateway with the access token.

### Option B — Dev HS256 JWT (Bearer path smoke only)

When `GATEWAY_DEV_JWT_SECRET` is set (local dev **only**):

```bash
TOKEN=$(cd go && go run ./cmd/dev-jwt)
curl -s http://127.0.0.1:8080/v1/users/me -H "Authorization: Bearer $TOKEN"
```

Automated smoke:

```bash
make e2e-mobile
```

This validates the Gateway **Bearer → identity headers** path; it does not exercise Logto PKCE.
Use Option A before shipping a real mobile client.

## Security checklist

- [ ] PKCE on every authorization request (S256)
- [ ] Tokens only in OS secure storage
- [ ] API base URL is HTTPS in production
- [ ] Certificate pinning considered for high-threat deployments (optional V1)
- [ ] No client secret in the mobile binary
- [ ] `GATEWAY_DEV_JWT_SECRET` disabled in production (`GATEWAY_BFF_DEV_LOGIN=false`)

## Related docs

- [LOGTO_SETUP.md](./LOGTO_SETUP.md) — Logto install + API resource
- [E2E_MINIPROGRAM.md](./E2E_MINIPROGRAM.md) — Bearer flow via auth-service (mini program)
- [E2E_ADMIN.md](./E2E_ADMIN.md) — Web cookie BFF flow
- [AI_CONTEXT.md](./AI_CONTEXT.md) — architecture rules (Gateway knows who)
