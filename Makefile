# PROJECT_NAME - Makefile

# Use bash for shell commands (required for 'set -a' env loading)
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Configuration
COMPOSE_FILE := infra/docker-compose.yml
PROJECT := app
MIGRATIONS_DIR := infra/migrations
POSTGRES_DSN := postgres://platform:platform_dev@localhost:5432/platform?sslmode=disable
TEST_POSTGRES_DSN := postgres://platform:platform_dev@localhost:5433/platform_test?sslmode=disable

# Directories
BIN_DIR := ./bin
LOG_DIR := ./logs
COVERAGE_DIR := artifacts/coverage

ENV_FILE := .env.dev
ENV_TEST := .env.test

# Version injection
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Tools (using go run to avoid global installs)
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0
MIGRATE := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
GOCOVMERGE := go run github.com/wadey/gocovmerge@latest

# Helper: load env file and run command
# Usage: $(call load-env,ENV_FILE,command)
# Also sources .env.local.secrets if it exists
define load-env
set -a; . ./$(1); \
if [ -f .env.local.secrets ]; then . ./.env.local.secrets; fi; \
set +a; $(2)
endef

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Build
# ============================================================================

.PHONY: build build-services build-tools clean
build: build-services build-tools ## Build all binaries

build-services: ## Build service binaries
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o $(BIN_DIR)/backend ./services/backend
	go build -o $(BIN_DIR)/dev ./services/dev

build-tools: ## Build tool binaries
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/seed ./tools/seed

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) $(LOG_DIR) $(COVERAGE_DIR)

# ============================================================================
# Development Infrastructure
# Runs infra in containers, app services locally via hot-reload (air)
# Ports: postgres:5432, minio:9000/9001, elasticmq:9324, mailhog:1025/8025
# ============================================================================

.PHONY: dev-up dev-down dev-reset dev-status dev-logs

dev-up: ## Start dev + test infrastructure (all containerized)
	docker compose -p $(PROJECT) -f $(COMPOSE_FILE) --profile dev --profile test up -d --build
	@sleep 3
	@echo "Dev environment ready at http://localhost:8080"
	@echo "Run 'make dev-logs' to follow backend logs"

dev-down: ## Stop all infrastructure (dev + test)
	docker compose -p $(PROJECT) -f $(COMPOSE_FILE) --profile dev --profile test down

dev-reset: ## Stop all infrastructure and remove volumes (dev + test)
	docker compose -p $(PROJECT) -f $(COMPOSE_FILE) --profile dev --profile test down -v

dev-status: ## Show all infrastructure status (dev + test)
	@docker compose -p $(PROJECT) -f $(COMPOSE_FILE) --profile dev --profile test ps

dev-logs: ## Follow backend logs (hot-reload output)
	docker compose -p $(PROJECT) -f $(COMPOSE_FILE) logs -f backend

.PHONY: seed seed-fresh

seed: build ## Initialize local dev environment (DB, MinIO, seed data)
	@$(call load-env,$(ENV_FILE),$(BIN_DIR)/seed)

seed-fresh: build ## Reset and re-initialize local dev environment (destructive!)
	@echo "WARNING: This will delete ALL data!"
	@$(call load-env,$(ENV_FILE),$(BIN_DIR)/seed --fresh)

# ============================================================================
# Database Migrations
# ============================================================================

.PHONY: migrate-up migrate-down migrate-create migrate-version
migrate-up: ## Run all pending migrations (dev database)
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(POSTGRES_DSN)" up

migrate-down: ## Rollback the last migration (dev database)
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(POSTGRES_DSN)" down 1

migrate-create: ## Create a new migration (usage: make migrate-create name=xxx)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-version: ## Show current migration version (dev database)
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(POSTGRES_DSN)" version

.PHONY: db-reset-test seed-test seed-test-fresh
db-reset-test: ## Reset test database (create if needed, drop schema, re-migrate)
	@echo "Resetting test database..."
	docker compose -p $(PROJECT) -f $(COMPOSE_FILE) exec -T test-postgres \
		psql -U platform -d postgres -c "SELECT 1 FROM pg_database WHERE datname = 'platform_test'" | grep -q 1 || \
		docker compose -p $(PROJECT) -f $(COMPOSE_FILE) exec -T test-postgres \
			psql -U platform -d postgres -c "CREATE DATABASE platform_test OWNER platform;"
	docker compose -p $(PROJECT) -f $(COMPOSE_FILE) exec -T test-postgres \
		psql -U platform -d platform_test -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(TEST_POSTGRES_DSN)" up
	@echo "Test database ready."

seed-test: build ## Initialize test environment (DB, MinIO, seed data)
	@$(call load-env,$(ENV_TEST),$(BIN_DIR)/seed)

seed-test-fresh: build ## Reset and re-initialize test environment (destructive!)
	@$(call load-env,$(ENV_TEST),$(BIN_DIR)/seed --fresh)

# ============================================================================
# Testing
# ============================================================================

.PHONY: test test-all
test: ## Run unit tests
	@mkdir -p $(COVERAGE_DIR)
	go test ./... -coverprofile=$(COVERAGE_DIR)/unit.out -covermode=atomic 2>&1 | tee artifacts/test.log
	@echo "Coverage: $$(go tool cover -func=$(COVERAGE_DIR)/unit.out | tail -1 | awk '{print $$3}')" | tee -a artifacts/test.log

test-all: ## Run unit and integration tests (requires dev-up first)
	@mkdir -p $(COVERAGE_DIR)
	@$(call load-env,$(ENV_TEST),go test -tags=integration -count=1 -timeout 120s \
		-coverprofile=$(COVERAGE_DIR)/integration.out -covermode=atomic \
		-coverpkg=./... ./...) 2>&1 | tee artifacts/test-all.log
	@echo "Coverage: $$(go tool cover -func=$(COVERAGE_DIR)/integration.out | tail -1 | awk '{print $$3}')" | tee -a artifacts/test-all.log

# ============================================================================
# Code Coverage
# ============================================================================

.PHONY: coverage coverage-html
coverage: test test-all ## Generate combined coverage report
	@mkdir -p $(COVERAGE_DIR)
	$(GOCOVMERGE) $(COVERAGE_DIR)/unit.out $(COVERAGE_DIR)/integration.out > $(COVERAGE_DIR)/all.out
	@echo "Combined coverage: $$(go tool cover -func=$(COVERAGE_DIR)/all.out | tail -1 | awk '{print $$3}')"

coverage-html: coverage ## Generate HTML coverage reports
	go tool cover -html=$(COVERAGE_DIR)/unit.out -o $(COVERAGE_DIR)/unit.html
	go tool cover -html=$(COVERAGE_DIR)/integration.out -o $(COVERAGE_DIR)/integration.html
	go tool cover -html=$(COVERAGE_DIR)/all.out -o $(COVERAGE_DIR)/all.html
	@echo "Reports: $(COVERAGE_DIR)/*.html"

coverage-summary: ## Show uncovered functions in internal packages
	@go tool cover -func=$(COVERAGE_DIR)/unit.out 2>/dev/null | \
		grep -E "^github.com/OWNER/PROJECT_NAME/internal" | \
		grep -E "[^0-9]0\.0%$$" | \
		sed 's|github.com/OWNER/PROJECT_NAME/||' | \
		head -50 || echo "Run 'make test' first to generate coverage"

.PHONY: coverage-file coverage-func coverage-low coverage-check

coverage-file: coverage ## Per-file coverage report (sorted ascending)
	@go tool cover -func=$(COVERAGE_DIR)/all.out | \
		grep -E "^github.com/OWNER/PROJECT_NAME/internal" | \
		grep -v "total:" | \
		sed 's|github.com/OWNER/PROJECT_NAME/||' | \
		awk '{split($$1,a,":"); file=a[1]; pct=$$NF; gsub(/%/,"",pct); funcs[file]++; total[file]+=pct} END {for(f in funcs){avg=total[f]/funcs[f]; printf "%5.1f%% %s\n",avg,f}}' | \
		sort -n

coverage-func: coverage ## Per-function coverage report (top 50, sorted ascending)
	@go tool cover -func=$(COVERAGE_DIR)/all.out | \
		grep -E "^github.com/OWNER/PROJECT_NAME/internal" | \
		grep -v "total:" | \
		sed 's|github.com/OWNER/PROJECT_NAME/||' | \
		awk '{pct=$$NF; gsub(/%/,"",pct); print pct" "$$0}' | \
		sort -n | cut -d' ' -f2- | head -50

coverage-low: coverage ## Show files with coverage < 60%
	@go tool cover -func=$(COVERAGE_DIR)/all.out | \
		grep -E "^github.com/OWNER/PROJECT_NAME/internal" | \
		grep -v "total:" | \
		sed 's|github.com/OWNER/PROJECT_NAME/||' | \
		awk '{split($$1,a,":"); file=a[1]; pct=$$NF; gsub(/%/,"",pct); funcs[file]++; total[file]+=pct} END {for(f in funcs){avg=total[f]/funcs[f]; if(avg<60) printf "%5.1f%% %s\n",avg,f}}' | \
		sort -n

coverage-check: coverage ## Fail if any file < 60% coverage
	@go tool cover -func=$(COVERAGE_DIR)/all.out | \
		grep -E "^github.com/OWNER/PROJECT_NAME/internal" | \
		grep -v "total:" | \
		awk '{split($$1,a,":"); file=a[1]; pct=$$NF; gsub(/%/,"",pct); funcs[file]++; total[file]+=pct} END {fail=0; for(f in funcs){avg=total[f]/funcs[f]; if(avg<60){printf "FAIL: %5.1f%% %s\n",avg,f; fail=1}} exit fail}'

# ============================================================================
# Linting
# ============================================================================

.PHONY: lint lint-check
lint: ## Run linter
	@mkdir -p artifacts
	$(GOLANGCI_LINT) run --fix ./... 2>&1 | tee artifacts/lint.log

lint-check: ## Run linter without auto-fix (for CI)
	$(GOLANGCI_LINT) run ./...

# ============================================================================
# AWS CDK Infrastructure
# ============================================================================

CDK_ENV ?= dev

.PHONY: cdk-deps cdk-synth cdk-diff cdk-deploy-network cdk-deploy-stateful cdk-deploy-compute cdk-deploy-all cdk-destroy cdk-bootstrap

cdk-deps: ## Install CDK CLI globally
	npm install -g aws-cdk

cdk-synth: ## Synthesize CloudFormation templates
	cd cdk && cdk synth

cdk-diff: ## Show pending infrastructure changes
	cd cdk && CDK_ENV=$(CDK_ENV) cdk diff

cdk-deploy-network: ## Deploy network stack (VPC, security groups)
	cd cdk && CDK_ENV=$(CDK_ENV) cdk deploy App-$(CDK_ENV)-Network --require-approval never

cdk-deploy-stateful: ## Deploy stateful stack (RDS, S3, SQS) - careful!
	cd cdk && CDK_ENV=$(CDK_ENV) cdk deploy App-$(CDK_ENV)-Stateful --require-approval broadening

cdk-deploy-compute: ## Deploy compute stack (ECS)
	cd cdk && CDK_ENV=$(CDK_ENV) cdk deploy App-$(CDK_ENV)-Compute --require-approval never

cdk-deploy-all: ## Deploy all stacks in order
	cd cdk && CDK_ENV=$(CDK_ENV) cdk deploy --all --require-approval broadening

cdk-destroy: ## Destroy all stacks (DANGEROUS - requires confirmation)
	@echo "WARNING: This will destroy ALL infrastructure in $(CDK_ENV)!"
	@read -p "Type 'destroy' to confirm: " confirm && [ "$$confirm" = "destroy" ]
	cd cdk && CDK_ENV=$(CDK_ENV) cdk destroy --all

cdk-bootstrap: ## Bootstrap CDK in AWS account (first-time setup)
	cd cdk && cdk bootstrap

# ============================================================================
# Utilities
# ============================================================================

.PHONY: cloc
cloc: ## Count lines of code in the repository
	@cloc --vcs=git --fullpath --not-match-d='internal/web/static' .
