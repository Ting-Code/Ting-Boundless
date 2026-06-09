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
lint: ## Run golangci-lint if installed
	@command -v golangci-lint >/dev/null 2>&1 && cd $(GO_DIR) && golangci-lint run || echo "golangci-lint not installed; skipping"

.PHONY: run-%
run-%: ## Run a single Go service, e.g. make run-gateway
	cd $(GO_DIR) && go run ./services/$*

.PHONY: proto
proto: ## Lint + generate code from platform-contracts (requires buf)
	cd platform-contracts && buf lint && buf generate

.PHONY: proto-breaking
proto-breaking: ## Check breaking proto changes against main (requires buf)
	cd platform-contracts && buf breaking --against '.git#branch=main'

COMPOSE_INFRA := docker compose -f deploy/docker-compose.infra.yml
COMPOSE_APPS  := docker compose -f deploy/docker-compose.yml
COMPOSE_ALL   := docker compose -f deploy/docker-compose.infra.yml -f deploy/docker-compose.yml

.PHONY: up-infra
up-infra: ## Start data infra only (Postgres, Redis, RabbitMQ, MinIO)
	$(COMPOSE_INFRA) up -d

.PHONY: up-apps
up-apps: ## Start application services only (use managed RDS/Redis in .env for prod)
	$(COMPOSE_APPS) up -d --build

.PHONY: up
up: ## Start infra + apps (local dev)
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

.PHONY: node-install
node-install: ## pnpm install in node/ monorepo
	cd node && pnpm install

.PHONY: node-build
node-build: ## Build logger + api-types + Nest business-service
	cd node && pnpm --filter @ting/logger build && pnpm --filter @ting/api-types build && pnpm --filter @ting/business-service build

.PHONY: run-business
run-business: ## Run Nest business-service (:3005)
	cd node && pnpm --filter @ting/logger build && pnpm --filter @ting/api-types build && pnpm dev:business

.PHONY: run-admin
run-admin: ## Run Vite admin dev server (:5173, proxies /v1 → Gateway)
	cd node && pnpm --filter @ting/api-types build && pnpm dev:admin

.PHONY: e2e-admin
e2e-admin: ## Smoke: Gateway dev cookie -> /v1/business/items (needs Redis + gateway + business)
	powershell -ExecutionPolicy Bypass -File scripts/e2e-admin-gateway.ps1

.PHONY: migrate
migrate: ## Run SQL migrations (user-service + audit-service; requires Postgres)
	cd $(GO_DIR) && go run ./cmd/migrate

.PHONY: dev-jwt
dev-jwt: ## Print a dev HS256 JWT (optional USER_ID arg)
	cd $(GO_DIR) && go run ./cmd/dev-jwt $(filter-out $@,$(MAKECMDGOALS))

.PHONY: up-logto
up-logto: ## Start Logto only (Docker; see docs/LOGTO_SETUP.md for native)
	docker compose -f deploy/docker-compose.logto.yml up -d

.PHONY: ci
ci: tidy vet build test ## Local CI bundle
