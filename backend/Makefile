.PHONY: dev migrate migrate-down bootstrap test build build-bin vet fmt-check check-structure check lint tidy

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

# Compile every package in the module — matches the "go build ./..."
# acceptance criterion exactly.
build:
	go build ./...

# Build runnable binaries into ./bin. Not required by any acceptance
# criterion; a convenience for local use/deployment.
build-bin:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/bootstrap ./cmd/bootstrap

vet:
	go vet ./...

# Fail if any file needs gofmt formatting.
fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needs to be run on:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Verify required scaffolding paths exist. "go build"/"go vet"/"go test"
# only validate whatever currently exists on disk — they wouldn't fail
# if e.g. usecase/implementation/ucshared were deleted — so this checks
# presence directly (per the ticket's usecase/interface,
# usecase/implementation, ucshared acceptance criterion).
check-structure:
	@missing=0; \
	for d in internal/usecase/interface internal/usecase/implementation internal/usecase/implementation/ucshared; do \
		if [ ! -d "$$d" ]; then \
			echo "missing required directory: $$d"; \
			missing=1; \
		fi; \
	done; \
	exit $$missing

# Full verification suite — the single command CI runs.
check: build vet fmt-check check-structure test

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
