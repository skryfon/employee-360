MAKEFLAGS += --no-print-directory

.PHONY: dev migrate migrate-down migrate-status migrate-version migrate-reset bootstrap test check build

dev:
	cd backend && go run ./cmd/api

migrate:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down

migrate-status:
	cd backend && go run ./cmd/migrate status

migrate-version:
	cd backend && go run ./cmd/migrate version

migrate-reset:
	cd backend && go run ./cmd/migrate reset

bootstrap:
	cd backend && go run ./cmd/bootstrap

test:
	cd backend && go test ./...

check:
	cd backend && $(MAKE) check

build:
	cd backend && $(MAKE) build
