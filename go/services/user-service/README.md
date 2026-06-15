# user-service

User domain. Trusts identity injected by the Gateway; never parses JWTs.

| Method | Path | Notes |
|--------|------|-------|
| GET | `/v1/users/me` | Get or create profile for current user |
| PATCH | `/v1/users/me` | Update `display_name` (`{"display_name":"..."}`) |
| GET | `/v1/users/` | List users in current tenant (`admin` role, `?limit=50`) |

Env: `HTTP_ADDR`, `POSTGRES_*`, `INTERNAL_API_TOKEN`.

```bash
make run-user-service
curl -H "Authorization: Bearer $(make dev-jwt)" http://127.0.0.1:8080/v1/users/me
curl -X PATCH -H "Authorization: Bearer $(make dev-jwt)" -H "Content-Type: application/json" \
  -d '{"display_name":"Ting"}' http://127.0.0.1:8080/v1/users/me
curl -H "Authorization: Bearer $(make dev-jwt)" 'http://127.0.0.1:8080/v1/users/?limit=20'
```
