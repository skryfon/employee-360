SHELL       := /bin/bash
BACKEND_DIR := backend
SCOPE       ?= employee-360

.PHONY: help bootstrap dev dev-api dev-admin dev-employee \
        migrate migrate-down migrate-status migrate-version migrate-reset \
        seed bootstrap-admin \
        test test-all test-backend test-backend-cover cover-func cover-html test-clients \
        check lint lint-backend lint-clients typecheck build clean tidy

help: ## Show this help menu
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## --- Bootstrap ---

bootstrap: ## Install frontend and backend dependencies
	@if [ -f package.json ]; then pnpm install; fi
	cd $(BACKEND_DIR) && go mod download

## --- Dev ---

dev: dev-api ## Run backend API (or full stack in future cycles)

dev-api: ## Run the Go backend API
	cd $(BACKEND_DIR) && go run ./cmd/api

dev-admin: ## Run admin client (clients/admin)
	pnpm --filter admin dev

dev-employee: ## Run employee client (clients/employee)
	pnpm --filter employee dev

## --- Migrations ---

migrate: ## Apply all pending database migrations
	cd $(BACKEND_DIR) && go run ./cmd/migrate up

migrate-down: ## Roll back the most recent database migration
	cd $(BACKEND_DIR) && go run ./cmd/migrate down

migrate-status: ## Show migration status
	cd $(BACKEND_DIR) && go run ./cmd/migrate status

migrate-version: ## Show current database migration version
	cd $(BACKEND_DIR) && go run ./cmd/migrate version

migrate-reset: ## Reset database migrations (revert all and reapply)
	cd $(BACKEND_DIR) && go run ./cmd/migrate reset

## --- Seeding & Bootstrap ---

bootstrap-admin: ## Seed core roles, tenant, and Super Admin
	cd $(BACKEND_DIR) && go run ./cmd/bootstrap

seed: bootstrap-admin ## Run all dev seeding

## --- Tests ---

test-all: check ## Run lint, typecheck, and all test suites

test: test-backend test-clients ## Run backend and frontend tests

test-backend: ## Run Go unit tests
	cd $(BACKEND_DIR) && go test ./...

test-backend-cover: ## Run Go tests with coverage output
	cd $(BACKEND_DIR) && go test -coverprofile=coverage.out ./...

cover-func: ## Show Go test coverage breakdown by function
	cd $(BACKEND_DIR) && go tool cover -func=coverage.out

cover-html: ## Open Go test coverage report in browser
	cd $(BACKEND_DIR) && go tool cover -html=coverage.out

test-clients: ## Run frontend tests
	@if [ -f package.json ]; then pnpm -r test; else echo "Frontend clients not initialized yet (Cycle 1 scope)"; fi

## --- Quality & Linting ---

check: ## Run full verification suite (backend + frontend)
	cd $(BACKEND_DIR) && $(MAKE) check
	@if [ -f package.json ]; then pnpm -r lint && pnpm -r typecheck; fi

lint: lint-backend lint-clients ## Run all linters

lint-backend: ## Run golangci-lint on backend
	cd $(BACKEND_DIR) && $(MAKE) lint

lint-clients: ## Run ESLint on frontend clients
	@if [ -f package.json ]; then pnpm -r lint; else echo "Frontend clients not initialized yet (Cycle 1 scope)"; fi

typecheck: ## Run TypeScript type checking
	@if [ -f package.json ]; then pnpm -r typecheck; else echo "Frontend clients not initialized yet (Cycle 1 scope)"; fi

tidy: ## Tidy Go module dependencies
	cd $(BACKEND_DIR) && go mod tidy

## --- Build & Clean ---

build: ## Build backend binaries and client bundles
	cd $(BACKEND_DIR) && $(MAKE) build-bin
	@if [ -f package.json ]; then pnpm -r build; fi

clean: ## Remove build artifacts and temporary files
	rm -rf $(BACKEND_DIR)/bin $(BACKEND_DIR)/tmp $(BACKEND_DIR)/coverage.out
	@if [ -f package.json ]; then pnpm -r exec rm -rf dist .turbo; fi
