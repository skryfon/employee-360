.PHONY: dev migrate migrate-down bootstrap test build lint tidy

# Run the API server.
dev:
	go run ./cmd/api

# Apply all pending migrations.
migrate:
	go run ./cmd/migrate up

# Revert all applied migrations.
migrate-down:
	go run ./cmd/migrate down

# Seed system tenant, core roles, and platform Super Admin.
# No-op beyond a connectivity check until Cycle 2+ adds the entities to seed.
bootstrap:
	go run ./cmd/bootstrap

# Run tests. No tests exist yet (Cycle 1 is scaffolding only); this
# target still runs cleanly against the whole module.
test:
	go test ./...

# Build all binaries into ./bin.
build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/bootstrap ./cmd/bootstrap

# Run golangci-lint if it's installed; otherwise fall back to go vet.
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, falling back to go vet"; \
		go vet ./...; \
	fi

tidy:
	go mod tidy
