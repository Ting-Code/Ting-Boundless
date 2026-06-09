# gateway

Edge API entry (Go Gateway/BFF). Decides **who** the caller is; never decides
domain permissions.

- Verifies caller credential (Bearer JWT via cached JWKS; web BFF cookie).
- Strips client-supplied identity headers, injects trusted ones (`pkg/identity`).
- Routes by path prefix to business services (`internal/proxy`).
- Rate limiting, unified errors, entry-level audit events.

Env: `HTTP_ADDR`, `USER_SERVICE_URL`, `BUSINESS_SERVICE_URL`, `FILE_SERVICE_URL`,
`OIDC_ISSUER`, `OIDC_JWKS_URL`, `OIDC_AUDIENCE`, `REDIS_ADDR`.

TODO: real JWT/cookie verification, revocation lookup, rate limiter.
