# user-service

User domain. Trusts identity injected by the Gateway; never parses JWTs.

Endpoints: `GET /v1/users/me`.

Env: `HTTP_ADDR`, `POSTGRES_*`, `REDIS_ADDR`.

TODO: persistence (app_db) + migrations; profile/membership model.
