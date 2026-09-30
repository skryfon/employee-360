//go:build integration

package persistence

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

// migrationsDir resolves the real backend/migrations directory relative to
// this source file's location (not the test binary's working directory), so
// it works regardless of how `go test` is invoked.
func migrationsDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller for migrationsDir")
	}
	// this file: backend/internal/infrastructure/persistence/testutil_test.go
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "migrations")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected migrations directory at %s: %v", dir, err)
	}
	return dir
}

// setupTestDB connects to a live PostgreSQL instance (same config resolution
// as backend/integration/migrate_test.go: config.Load
// defaults + DATABASE_* env vars) and ensures the full schema is applied.
// It skips the test (fails in CI) if PostgreSQL is unreachable, matching the
// existing convention in this repo for DB-touching tests.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// This test file lives 4 directories below the repo root, where .env
	// actually lives (persistence -> infrastructure -> internal -> backend
	// -> repo root), so it needs an explicit search path -- config.Load's
	// own built-in candidates only look up to 2 levels up (see
	// config.loadDotEnv), matching backend/integration/'s shallower package
	// depth, not this one's.
	cfg, err := config.Load("../../../..")
	if err != nil {
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatalf("failed to load config in CI: %v", err)
		}
		t.Skipf("skipping live database test: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatalf("PostgreSQL must be reachable in CI: %v", err)
		}
		t.Skipf("skipping live database test (PostgreSQL unreachable: %v)", err)
	}

	ensureMigrated(t, cfg.Database.URL())

	return db
}

// ensureMigrated applies every migration in backend/migrations idempotently.
// golang-migrate takes a Postgres advisory lock for the duration of Up(), so
// this is safe to call concurrently with other test packages/processes doing
// the same thing against a shared database.
func ensureMigrated(t *testing.T, dbURL string) {
	t.Helper()

	dir := migrationsDir(t)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("failed to resolve absolute migrations dir: %v", err)
	}

	m, err := migrate.New("file://"+absDir, dbURL)
	if err != nil {
		t.Fatalf("failed to initialize migrator: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("failed to apply migrations: %v", err)
	}
}

// createTestTenant inserts a tenant row directly (tenants carries no
// tenant_id itself) and registers cleanup to remove it (cascades to any
// users/tokens created under it, per the migrations' ON DELETE CASCADE FKs).
func createTestTenant(t *testing.T, db *gorm.DB, name string) *entity.Tenant {
	t.Helper()

	tenant := &entity.Tenant{
		ID:        uuid.New(),
		Name:      name,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("failed to create test tenant: %v", err)
	}
	t.Cleanup(func() {
		db.Unscoped().Where("id = ?", tenant.ID).Delete(&entity.Tenant{})
	})
	return tenant
}
