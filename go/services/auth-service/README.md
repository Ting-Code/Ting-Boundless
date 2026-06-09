# auth-service

IdP integration layer (platform service, no domain logic).

- Receives + verifies Logto webhooks; converts identity events → audit events.
- WeChat mini-program login: `code2session` → issues a **standard JWT**.
- Optionally an internal OIDC issuer (own issuer + JWKS the Gateway trusts).

Endpoints: `POST /internal/webhooks/logto`, `POST /v1/auth/miniprogram/login`.

Env: `HTTP_ADDR`, `OIDC_ISSUER`, `OIDC_JWKS_URL`, `WECHAT_APP_ID`,
`WECHAT_APP_SECRET`, `POSTGRES_*`, `REDIS_ADDR`.

TODO: Logto signature verification, code2session, JWT issuance/JWKS.
