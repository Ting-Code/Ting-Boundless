# go/ — Go monorepo

All Go code lives under this directory. The module root is `go/go.mod`
(`github.com/ting-boundless/boundless`).

## Layout

```text
go/
  pkg/          shared libraries (identity, httpx, db, …)
  services/     platform deployables (gateway, auth, user, file, audit, worker)
  cmd/          migrate, dev-jwt
  migrations/   SQL migrations per service
```

Domain CRUD is **not** here — see `node/apps/business-service` (NestJS).

## Commands

From repo root (preferred):

```bash
make build
make run-gateway
make run-user-service
make migrate
make dev-jwt
```

From this directory:

```bash
go build ./...
go run ./services/gateway
go run ./cmd/migrate
```

`.env` is loaded from `go/.env` or repo root `../.env` via `pkg/config.LoadEnvFile()`.
