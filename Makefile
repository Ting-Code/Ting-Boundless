.DEFAULT_GOAL := help
GO_DIR := go
SERVICES := gateway auth-service user-service file-service audit-service worker

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## go mod tidy
	cd $(GO_DIR) && go mod tidy

.PHONY: build
build: ## Build all Go packages
	cd $(GO_DIR) && go build ./...

.PHONY: vet
vet: ## go vet
	cd $(GO_DIR) && go vet ./...

.PHONY: test
test: ## Run Go tests
	cd $(GO_DIR) && go test ./...

.PHONY: lint
lint: ## Run golangci-lint (install: https://golangci-lint.run/welcome/install/)
	cd $(GO_DIR) && golangci-lint run

.PHONY: run-%
run-%: ## Run a single Go service, e.g. make run-gateway
	cd $(GO_DIR) && go run ./services/$*

.PHONY: proto
proto: ## Lint + generate code from platform-contracts (requires buf)
	cd platform-contracts && buf lint && buf generate
	@test -f go/gen/go/ting/common/v1/identity.pb.go || (echo "proto generate failed: missing go/gen/go stubs" && exit 1)

.PHONY: proto-breaking
proto-breaking: ## Check breaking proto changes against main (requires buf)
	cd platform-contracts && buf breaking --against '.git#branch=main'

.PHONY: lint-openapi
lint-openapi: ## Lint OpenAPI specs with Redocly (requires Node/npx)
	cd platform-contracts && npx --yes @redocly/cli@1.34.3 lint openapi/*.yaml

.PHONY: openapi-breaking
openapi-breaking: ## Check breaking OpenAPI changes vs main (requires oasdiff)
	bash scripts/oasdiff-breaking.sh

COMPOSE_INFRA := docker compose -f deploy/docker-compose.infra.yml
COMPOSE_APPS  := docker compose -f deploy/docker-compose.yml
COMPOSE_LOCAL := docker compose -f deploy/docker-compose.infra.yml -f deploy/docker-compose.yml -f deploy/docker-compose.local.yml
COMPOSE_ALL   := $(COMPOSE_LOCAL)
COMPOSE_OBS   := docker compose -f deploy/docker-compose.yml

.PHONY: node-install
node-install: ## pnpm install in node/ monorepo
	cd node && pnpm install

.PHONY: generate-api
generate-api: ## Generate @ting/api types from platform-contracts OpenAPI
	cd node && pnpm generate:api

.PHONY: up-infra
up-infra: ## Start data infra only (Postgres, Redis, RabbitMQ, MinIO)
	$(COMPOSE_INFRA) up -d

.PHONY: up-apps
up-apps: ## Start application services only (use managed RDS/Redis in .env for prod)
	$(COMPOSE_APPS) up -d --build

.PHONY: up-obs
up-obs: ## Start observability stack only (Tempo, Loki, OTel Collector, Grafana)
	$(COMPOSE_OBS) up -d tempo loki otel-collector grafana

.PHONY: up
up: ## Start infra + apps with healthy startup ordering (local dev)
	$(COMPOSE_ALL) up -d --build

.PHONY: down-apps
down-apps: ## Stop application services
	$(COMPOSE_APPS) down

.PHONY: down-infra
down-infra: ## Stop data infra (keeps volumes unless -v)
	$(COMPOSE_INFRA) down

.PHONY: down
down: ## Stop infra + apps
	$(COMPOSE_ALL) down

.PHONY: down-obs
down-obs: ## Stop observability stack
	$(COMPOSE_OBS) stop tempo loki otel-collector grafana

.PHONY: node-build
node-build: generate-api ## Build logger + @ting/api + Nest business-service
	cd node && pnpm --filter @ting/logger build && pnpm --filter @ting/api build && pnpm --filter @ting/business-service build

.PHONY: node-typecheck
node-typecheck: generate-api ## Typecheck admin + business-service (requires pnpm install)
	cd node && pnpm typecheck

.PHONY: node-test
node-test: ## Run @ting/api unit tests
	cd node && pnpm test

.PHONY: run-business
run-business: ## Run Nest business-service (:3005)
	cd node && pnpm --filter @ting/logger build && pnpm --filter @ting/api build && pnpm dev:business

.PHONY: run-admin
run-admin: ## Run Vite admin dev server (:5173, proxies /v1 → Gateway)
	cd node && pnpm --filter @ting/api build && pnpm dev:admin

.PHONY: run-site
run-site: ## Run Next.js site dev server (:3006; browse via Gateway :8080/)
	cd node && pnpm --filter @ting/api build && pnpm dev:site

.PHONY: e2e-admin
e2e-admin: ## Smoke: Gateway dev cookie -> business items + audit (needs Redis, gateway, business, audit)
	powershell -ExecutionPolicy Bypass -File scripts/e2e-admin-gateway.ps1

.PHONY: e2e-miniprogram
e2e-miniprogram: ## Smoke: mock mini-program login -> Bearer JWT -> /v1/users/me
	powershell -ExecutionPolicy Bypass -File scripts/e2e-miniprogram-gateway.ps1

.PHONY: e2e-mobile
e2e-mobile: ## Smoke: dev Bearer JWT -> /v1/users/me + /v1/business/me (see docs/MOBILE_AUTH.md)
	powershell -ExecutionPolicy Bypass -File scripts/e2e-mobile-bearer.ps1

.PHONY: migrate
migrate: ## Run SQL migrations (Go services; requires Postgres)
	cd $(GO_DIR) && go run ./cmd/migrate

.PHONY: dev-jwt
dev-jwt: ## Print a dev HS256 JWT (optional USER_ID arg)
	cd $(GO_DIR) && go run ./cmd/dev-jwt $(filter-out $@,$(MAKECMDGOALS))

.PHONY: enqueue-ping
enqueue-ping: ## Publish a ping job to RabbitMQ (needs RABBITMQ_URL + running worker)
	cd $(GO_DIR) && go run ./cmd/enqueue-job -type ping

.PHONY: up-logto
up-logto: ## Start Logto only (Docker; see docs/LOGTO_SETUP.md for native)
	docker compose -f deploy/docker-compose.logto.yml up -d

.PHONY: ci
ci: tidy vet build test node-test node-typecheck ## Local CI bundle (go + node subset; full CI includes buf + Trivy)
