---
name: new-go-service
description: Scaffold a new Go service in the Ting Boundless monorepo following the platform baseline (health, ECS logging, identity, audit, unified errors, Dockerfile, compose). Use when adding a new backend service, microservice, or "services/<name>" in this repo.
---

# New Go Service

Scaffold a service under `services/<name>/` that conforms to the architecture
baseline in `docs/AI_CONTEXT.md` and `.cursor/rules/`.

## Workflow

```
- [ ] 1. Pick a kebab-case name (e.g. order-service)
- [ ] 2. Create services/<name>/main.go from the template below
- [ ] 3. Put service-private code in services/<name>/internal/
- [ ] 4. Add services/<name>/README.md (responsibilities, endpoints, deps)
- [ ] 5. Register the service in `deploy/docker-compose.yml` (SERVICE build arg)
- [ ] 6. If it owns data: add a database + golang-migrate migrations + outbox
- [ ] 7. If the Gateway should route to it: add a prefix in services/gateway/main.go
- [ ] 8. Run: make build && make vet
```

## Rules to honor

- Business services NEVER parse end-user JWTs. Read identity from context via
  `identity.FromContext`; wrap routes with `identity.Middleware`.
- Reuse `pkg/`: `logger`, `httpx`, `identity`, `errs`, `audit`. Don't reinvent.
- Expose `/healthz` (liveness) and `/readyz` (real dependency probes).
- Log JSON to stdout via `pkg/logger`. Return errors with `errs.Write`.
- Config from env via `httpx.Env`. No secrets in code.
- Domain audit events use the outbox pattern (see `.cursor/rules/audit-events.mdc`).

## main.go template

```go
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "ORDER-service" // replace

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
	}

	health := httpx.NewHealth()
	db.RegisterHealth(health, "postgres", pg.Probe)

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("GET /v1/orders/", identity.Middleware(http.HandlerFunc(handleList)))

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8080")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleList(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.JSON(w, http.StatusUnauthorized, map[string]string{"code": "unauthenticated"})
		return
	}
	// Enforce tenant isolation: scope all queries by id.TenantID.
	httpx.JSON(w, http.StatusOK, map[string]any{"tenant_id": id.TenantID, "items": []any{}})
}
```

## docker-compose entry

Add to `deploy/docker-compose.yml` (application services only — not `docker-compose.infra.yml`):

```yaml
  order-service:
    <<: *service-build
    build:
      context: ..
      dockerfile: deploy/Dockerfile
      args: { SERVICE: order-service }
```

Services connect to PostgreSQL/Redis via env (`POSTGRES_HOST`, `REDIS_ADDR`), not via
`depends_on` on data stores. Readiness probes (`/readyz`) verify connectivity.

## Verify

```bash
make build   # go build ./...
make vet
```
