# gateway

Edge API entry (Go Gateway/BFF). Decides **who** the caller is; never decides
domain permissions.

- Verifies caller credential (Bearer JWT via cached JWKS; web BFF cookie).
- Strips client-supplied identity headers, injects trusted ones (`pkg/identity`).
- Routes by path prefix to business services (`internal/proxy`).
- Rate limiting, unified errors, entry-level audit events.

## Environment

| Variable | Purpose |
|----------|---------|
| `HTTP_ADDR` | Listen address (default `:8080`) |
| `USER_SERVICE_URL`, `BUSINESS_SERVICE_URL`, `FILE_SERVICE_URL`, `AUTH_SERVICE_URL`, `AUDIT_SERVICE_URL` | Upstream bases |
| `SITE_SERVICE_URL` | Next.js SSR site (default `http://127.0.0.1:3006`; placeholder disables) |
| `OIDC_*`, `GATEWAY_PUBLIC_URL` | Logto BFF (`/sign-in`, `/callback`) — see `docs/BFF_LOGTO.md` |
| `GATEWAY_BFF_DEV_LOGIN` | When `true`, adds `/sign-in/dev` to anonymous paths |
| `/sign-in/dev` query | `user_id`, `tenant_id`, `return_to`, `roles` (comma-separated; default `user,admin`) |
| `GATEWAY_ANON_PREFIXES` | Override anonymous path rules (see below) |
| `INTERNAL_API_TOKEN` | Injected on upstream proxy; services verify `X-Internal-Token` |
| `AUDIT_SERVICE_URL` | Async entry audit (`api.access.denied`, `api.token.invalid`) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector (optional; disables tracing when unset) |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` (default :4317) or `http/protobuf` (:4318) |
| `GATEWAY_RATE_LIMIT_AUTH_RPS` | Stricter per-IP limit for `/v1/auth/`, login paths (default 5/s) |
| `GATEWAY_RATE_LIMIT_GENERAL_RPS` | General API per-IP limit (default 50/s) |
| `GATEWAY_RATE_LIMIT_REDIS` | `true` = shared Redis counters across Gateway replicas |
| `GATEWAY_SENSITIVE_PREFIXES` | Paths requiring revocation check (default `/v1/files/`; `-` disables) |
| `AUTH_REVOCATION_TTL` | Redis TTL for revoked subjects/sessions (default 30d) |
| `REDIS_ADDR` | BFF sessions, distributed rate limits, revocation blocklist |
| `SESSION_COOKIE_SECURE` | `Set-Cookie Secure` (default `true` when `APP_ENV=production`) |

## Anonymous path rules (`GATEWAY_ANON_PREFIXES`)

Paths in this list may reach the Gateway **without** a Bearer JWT or BFF session cookie.
Business services still enforce their own rules; most `/v1/*` APIs require identity
headers injected only after successful Gateway authentication.

**Syntax** (comma-separated):

| Form | Meaning | Example |
|------|---------|---------|
| `=/path` | Exact path only | `=/sign-in` allows `/sign-in` but **not** `/sign-in/dev` |
| `/path` | Prefix match | `/admin` allows `/admin/items` |

**Default** (when env unset):

```text
=/healthz,=/readyz,=/metrics
=/sign-in,=/callback,=/sign-out
=/internal/webhooks/logto
=/                  (exact — public SSR home via Next)
=/v1/business/ping
/admin          (prefix — SPA static)
/v1/auth/       (prefix — login integration; nginx rate-limited)
```

When `GATEWAY_BFF_DEV_LOGIN=true`, `/sign-in/dev` is appended automatically.

**Production:** set `GATEWAY_ANON_PREFIXES` explicitly; keep `GATEWAY_BFF_DEV_LOGIN=false`.

Implementation: `internal/auth/anon.go`, `ResolveAnonPrefixes()`.

## Local dev

```bash
make run-gateway
# OIDC: docs/LOGTO_SETUP.md
# Admin SPA: cd node && pnpm --filter @ting/admin build  (served at /admin when dist exists)
# Public site: make run-site  (Next :3006; Gateway proxies / → SITE_SERVICE_URL)
```

## Tests

```bash
cd go && go test ./services/gateway/... -run Integration
```

Integration tests cover Bearer JWT proxying, edge 401, rate limiting on `/v1/auth/`,
revocation on sensitive `/v1/files/*`, and S-08 request_id regeneration.
