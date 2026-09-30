//go:build integration

package eventing_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
	"github.com/skryfon/employee360/backend/internal/infrastructure/eventing"
)

func migrationsDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller for migrationsDir")
	}
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "migrations")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected migrations directory at %s: %v", dir, err)
	}
	return dir
}

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

func setup(t *testing.T) (*gorm.DB, *eventing.RiverPublisher, *database.GormTransactor) {
	t.Helper()
	cfg, err := config.Load("../../../..")
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
	db, err := database.Connect(cfg.Database)
	if err != nil {
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatalf("PostgreSQL must be reachable in CI: %v", err)
		}
		t.Skipf("skipping: PostgreSQL unreachable: %v", err)
	}
	ensureMigrated(t, cfg.Database.URL())
	require.True(t, db.Migrator().HasTable("river_job"), "river_job missing: run `make migrate`")
	sqlDB, err := db.DB()
	require.NoError(t, err)
	client, err := eventing.NewInsertOnlyClient(sqlDB)
	require.NoError(t, err)
	return db, eventing.NewRiverPublisher(client, eventing.NewDispatcher()), database.NewGormTransactor(db)
}

func jobCount(t *testing.T, db *gorm.DB, to string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Raw("SELECT count(*) FROM river_job WHERE kind = 'send_email' AND args->>'to' = ?", to).Scan(&n).Error)
	return n
}

func tenantCount(t *testing.T, db *gorm.DB, id uuid.UUID) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&entity.Tenant{}).Where("id = ?", id).Count(&n).Error)
	return n
}

func newEvent(tenantID uuid.UUID, to string) event.Event {
	return event.Event{
		ID: uuid.New(), TenantID: tenantID, EventType: event.EventTypePasswordResetRequested,
		Payload: event.PasswordResetRequestedPayload{
			UserID: uuid.New(), TenantID: tenantID, Email: to, ResetURL: "http://x", ExpiresAt: time.Now().Add(time.Hour),
		},
		OccurredAt: time.Now(),
	}
}

func work(ctx context.Context, db *gorm.DB, pub *eventing.RiverPublisher, tenant *entity.Tenant, to string) error {
	if err := database.DBFromContext(ctx, db).Create(tenant).Error; err != nil {
		return err
	}
	return pub.Publish(ctx, newEvent(tenant.ID, to))
}

func TestOutbox_RollbackLeavesNeitherRowNorJob(t *testing.T) {
	db, pub, tx := setup(t)
	to := fmt.Sprintf("rollback-%s@example.test", uuid.NewString())
	tenant := &entity.Tenant{ID: uuid.New(), Name: "outbox-rollback-" + uuid.NewString(), IsActive: true}
	sentinel := errors.New("force rollback")

	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		require.NoError(t, work(ctx, db, pub, tenant, to))
		// Both writes are visible inside the tx...
		var inTx int64
		require.NoError(t, database.DBFromContext(ctx, db).Model(&entity.Tenant{}).Where("id = ?", tenant.ID).Count(&inTx).Error)
		require.EqualValues(t, 1, inTx)
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	require.EqualValues(t, 0, tenantCount(t, db, tenant.ID), "business row must be rolled back")
	require.EqualValues(t, 0, jobCount(t, db, to), "river job must be rolled back")
}

func TestOutbox_CommitPersistsBoth(t *testing.T) {
	db, pub, tx := setup(t)
	to := fmt.Sprintf("commit-%s@example.test", uuid.NewString())
	tenant := &entity.Tenant{ID: uuid.New(), Name: "outbox-commit-" + uuid.NewString(), IsActive: true}
	t.Cleanup(func() {
		db.Exec("DELETE FROM river_job WHERE args->>'to' = ?", to)
		db.Unscoped().Where("id = ?", tenant.ID).Delete(&entity.Tenant{})
	})

	require.NoError(t, tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		return work(ctx, db, pub, tenant, to)
	}))
	require.EqualValues(t, 1, tenantCount(t, db, tenant.ID))
	require.EqualValues(t, 1, jobCount(t, db, to))

	var tenantInArgs string
	require.NoError(t, db.Raw("SELECT args->>'tenant_id' FROM river_job WHERE args->>'to' = ?", to).Scan(&tenantInArgs).Error)
	require.Equal(t, tenant.ID.String(), tenantInArgs)
}

func TestOutbox_PublishOutsideTransactionFails(t *testing.T) {
	_, pub, _ := setup(t)
	err := pub.Publish(context.Background(), newEvent(uuid.New(), "x@example.test"))
	require.ErrorIs(t, err, database.ErrNoSQLTx)
}
