//go:build integration

package seeder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
	"github.com/skryfon/employee360/backend/internal/infrastructure/persistence"
	"github.com/skryfon/employee360/backend/internal/infrastructure/service"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	inCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"

	// seeder -> database -> infrastructure -> internal -> backend -> repo root
	cfg, err := config.Load("../../../../..")
	if err != nil {
		if inCI {
			t.Fatalf("failed to load config in CI: %v", err)
		}
		t.Skipf("skipping live database test: %v", err)
	}
	db, err := database.Connect(cfg.Database)
	if err != nil {
		if inCI {
			t.Fatalf("PostgreSQL must be reachable in CI: %v", err)
		}
		t.Skipf("skipping live database test (PostgreSQL unreachable: %v)", err)
	}

	_, thisFile, _, _ := runtime.Caller(0)
	dir, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "migrations"))
	require.NoError(t, err)
	m, err := migrate.New("file://"+dir, cfg.Database.URL())
	require.NoError(t, err)
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	return db
}

func count(t *testing.T, db *gorm.DB, model any, where string, args ...any) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(model).Where(where, args...).Count(&n).Error)
	return n
}

func TestRun_IsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	name := "seeder-test-" + uuid.NewString()
	domain := "seeder-" + uuid.NewString() + ".example.com"
	email := "root@" + domain
	t.Cleanup(func() { db.Where("name = ?", name).Delete(&entity.Tenant{}) }) // cascades

	s := New(db, service.NewHashService(bcrypt.MinCost))
	opts := Options{SystemTenantName: name, SuperAdminEmail: email, SuperAdminPassword: "s3cret-pass"}

	first, err := s.Run(context.Background(), opts)
	require.NoError(t, err)
	assert.True(t, first.TenantCreated)
	assert.True(t, first.AdminCreated)
	assert.True(t, first.DomainCreated)
	assert.ElementsMatch(t, DefaultRoles(), first.RolesCreated)

	second, err := s.Run(context.Background(), opts)
	require.NoError(t, err)
	assert.Equal(t, first.TenantID, second.TenantID)
	assert.False(t, second.TenantCreated)
	assert.False(t, second.AdminCreated)
	assert.False(t, second.AdminRoleAdded)
	assert.False(t, second.DomainCreated)
	assert.EqualValues(t, 1, count(t, db, &entity.TenantDomain{}, "domain = ?", domain))
	assert.Empty(t, second.RolesCreated)

	assert.EqualValues(t, 1, count(t, db, &entity.Tenant{}, "name = ?", name))
	assert.EqualValues(t, 3, count(t, db, &entity.Role{}, "tenant_id = ?", first.TenantID))
	assert.EqualValues(t, 1, count(t, db, &entity.User{}, "tenant_id = ? AND email = ?", first.TenantID, email))
	assert.EqualValues(t, 1, count(t, db, &entity.UserRole{}, "tenant_id = ? AND user_id = ?", first.TenantID, first.AdminID))

	// Password is stored hashed, never plaintext.
	var u entity.User
	require.NoError(t, db.First(&u, "id = ?", first.AdminID).Error)
	require.NotNil(t, u.PasswordHash)
	assert.NotEqual(t, "s3cret-pass", *u.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte("s3cret-pass")))
}

func TestRun_MissingPasswordLeavesDBUntouched(t *testing.T) {
	db := setupTestDB(t)
	name := "seeder-test-" + uuid.NewString()

	_, err := New(db, service.NewHashService(bcrypt.MinCost)).Run(context.Background(),
		Options{SystemTenantName: name, SuperAdminEmail: "x@example.com"})
	require.ErrorIs(t, err, ErrSuperAdminPasswordRequired)
	assert.EqualValues(t, 0, count(t, db, &entity.Tenant{}, "name = ?", name))
}

func TestRun_DomainOwnedByOtherTenantErrors(t *testing.T) {
	db := setupTestDB(t)
	domain := "seeder-" + uuid.NewString() + ".example.com"
	sysName := "seeder-test-" + uuid.NewString()
	other := entity.Tenant{ID: uuid.New(), Name: "seeder-other-" + uuid.NewString(), IsActive: true}
	require.NoError(t, db.Create(&other).Error)
	t.Cleanup(func() {
		db.Where("id = ?", other.ID).Delete(&entity.Tenant{})
		db.Where("name = ?", sysName).Delete(&entity.Tenant{})
	})
	require.NoError(t, db.Exec("INSERT INTO tenant_domains (id, tenant_id, domain) VALUES (?, ?, ?)",
		uuid.New(), other.ID, domain).Error)

	_, err := New(db, service.NewHashService(bcrypt.MinCost)).Run(context.Background(),
		Options{SystemTenantName: sysName, SuperAdminEmail: "root@" + domain, SuperAdminPassword: "pw"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain "+domain+" already belongs to another tenant")
	// Transaction rolled back: nothing was created for the system tenant.
	assert.EqualValues(t, 0, count(t, db, &entity.Tenant{}, "name = ?", sysName))
}

func TestRun_ExistingBootstrapWithoutDomainGetsBackfilled(t *testing.T) {
	db := setupTestDB(t)
	name := "seeder-test-" + uuid.NewString()
	domain := "seeder-" + uuid.NewString() + ".example.com"
	email := "Root@" + domain // mixed case: domain must be lowercased
	t.Cleanup(func() { db.Where("name = ?", name).Delete(&entity.Tenant{}) })

	s := New(db, service.NewHashService(bcrypt.MinCost))
	opts := Options{SystemTenantName: name, SuperAdminEmail: email, SuperAdminPassword: "s3cret-pass"}
	first, err := s.Run(context.Background(), opts)
	require.NoError(t, err)

	// Simulate a database bootstrapped before domain registration existed.
	require.NoError(t, db.Exec("DELETE FROM tenant_domains WHERE tenant_id = ?", first.TenantID).Error)
	var before entity.User
	require.NoError(t, db.First(&before, "id = ?", first.AdminID).Error)

	// Re-run with a different password: domain is added, password untouched.
	opts.SuperAdminPassword = "different-pass"
	second, err := s.Run(context.Background(), opts)
	require.NoError(t, err)
	assert.True(t, second.DomainCreated)
	assert.False(t, second.AdminCreated)
	var after entity.User
	require.NoError(t, db.First(&after, "id = ?", first.AdminID).Error)
	assert.Equal(t, before.PasswordHash, after.PasswordHash)

	// The bootstrapped domain resolves to the system tenant for login.
	got, err := persistence.NewGormTenantDomainRepository(db).FindTenantByDomain(context.Background(), domain)
	require.NoError(t, err)
	assert.Equal(t, first.TenantID, got.ID)
}
