# Taskboard - Makefile

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Directories
BIN_DIR := ./bin
COVERAGE_DIR := artifacts/coverage

# Version injection
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Tools
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Build
# ============================================================================

.PHONY: build clean

build: ## Build the taskboard binary
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o $(BIN_DIR)/taskboard ./cmd/taskboard

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) $(COVERAGE_DIR) artifacts/

# ============================================================================
# Run
# ============================================================================

.PHONY: run dev seed

run: build ## Build and run taskboard
	$(BIN_DIR)/taskboard

dev: ## Initialize sample data and start dev server with hot reload (air)
	@if [ ! -f .taskboard/config.yaml ]; then go run ./cmd/taskboard --init; fi
	@if ! ls .taskboard/*.yaml 2>/dev/null | grep -qv config.yaml; then $(MAKE) seed; fi
	go run github.com/air-verse/air@v1.61.7

seed: ## Create sample task files in .taskboard/
	@if [ ! -f .taskboard/config.yaml ]; then go run ./cmd/taskboard --init; fi
	@go run ./cmd/seed

# ============================================================================
# Testing
# ============================================================================

.PHONY: test test-all

test: ## Run unit tests
	@mkdir -p $(COVERAGE_DIR) artifacts
	go test ./... -coverprofile=$(COVERAGE_DIR)/unit.out -covermode=atomic 2>&1 | tee artifacts/test.log
	@echo "Coverage: $$(go tool cover -func=$(COVERAGE_DIR)/unit.out | tail -1 | awk '{print $$3}')" | tee -a artifacts/test.log

test-all: test ## Run all tests (alias for test — no integration tests)

# ============================================================================
# Linting
# ============================================================================

.PHONY: lint lint-check

lint: ## Run linter with auto-fix
	@mkdir -p artifacts
	$(GOLANGCI_LINT) run --fix ./... 2>&1 | tee artifacts/lint.log

lint-check: ## Run linter without auto-fix (for CI)
	$(GOLANGCI_LINT) run ./...

# ============================================================================
# Utilities
# ============================================================================

.PHONY: cloc
cloc: ## Count lines of code in the repository
	@cloc . --vcs=git --fullpath --not-match-d='internal/web/static'
