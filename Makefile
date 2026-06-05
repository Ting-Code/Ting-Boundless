.DEFAULT_GOAL := help
SERVICES := gateway auth-service user-service business-service file-service audit-service worker

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: build
build: ## Build all packages
	go build ./...

.PHONY: vet
vet: ## go vet
	go vet ./...

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: lint
lint: ## Run golangci-lint if installed
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

.PHONY: run-%
run-%: ## Run a single service, e.g. make run-gateway
	go run ./services/$*

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

.PHONY: ci
ci: tidy vet build test ## Local CI bundle
