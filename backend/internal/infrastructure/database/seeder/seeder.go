package seeder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/domain/service"
)

// bootstrapLockKey is the Postgres advisory-lock key serializing concurrent
// bootstrap runs, so two simultaneous runs cannot both miss the tenant
// lookup and insert duplicates (tenants has no unique key on name).
const bootstrapLockKey int64 = 0x45333630 // "E360"

// Default names used when the corresponding option is left empty.
const (
	DefaultSystemTenantName = "System"
	superAdminFirstName     = "Platform"
	superAdminLastName      = "Super Admin"
)

// ErrSuperAdminEmailRequired and ErrSuperAdminPasswordRequired are returned
// when the bootstrap credentials are not configured. There are deliberately
// no defaults: credentials must come from configuration/environment.
var (
	ErrSuperAdminEmailRequired    = errors.New("bootstrap: super admin email is not set (BOOTSTRAP_SUPER_ADMIN_EMAIL)")
	ErrSuperAdminPasswordRequired = errors.New("bootstrap: super admin password is not set (BOOTSTRAP_SUPER_ADMIN_PASSWORD)")
	ErrSuperAdminEmailInvalid     = errors.New("bootstrap: super admin email must contain exactly one '@' with a non-empty local part and domain")
)

// Options configures a bootstrap run.
type Options struct {
	SystemTenantName   string
	SuperAdminEmail    string
	SuperAdminPassword string
}

// Result reports what a bootstrap run did, for logging.
type Result struct {
	TenantID       uuid.UUID
	TenantCreated  bool
	RolesCreated   []string
	AdminID        uuid.UUID
	AdminCreated   bool
	AdminRoleAdded bool
	DomainCreated  bool
}

// Validate normalizes the options and checks required credentials.
func (o *Options) Validate() error {
	o.SystemTenantName = strings.TrimSpace(o.SystemTenantName)
	if o.SystemTenantName == "" {
		o.SystemTenantName = DefaultSystemTenantName
	}
	o.SuperAdminEmail = strings.ToLower(strings.TrimSpace(o.SuperAdminEmail))
	if o.SuperAdminEmail == "" {
		return ErrSuperAdminEmailRequired
	}
	if _, err := emailDomain(o.SuperAdminEmail); err != nil {
		return err
	}
	if o.SuperAdminPassword == "" {
		return ErrSuperAdminPasswordRequired
	}
	return nil
}

// emailDomain returns the lowercased domain part of an email that has exactly
// one '@' and non-empty local and domain parts.
func emailDomain(email string) (string, error) {
	if strings.Count(email, "@") != 1 {
		return "", ErrSuperAdminEmailInvalid
	}
	local, domain, _ := strings.Cut(email, "@")
	domain = strings.ToLower(strings.TrimSpace(domain))
	if strings.TrimSpace(local) == "" || domain == "" {
		return "", ErrSuperAdminEmailInvalid
	}
	return domain, nil
}

// Seeder bootstraps the system tenant, default roles and platform Super Admin.
type Seeder struct {
	db     *gorm.DB
	hasher service.HashService
}

// New constructs a Seeder.
func New(db *gorm.DB, hasher service.HashService) *Seeder {
	return &Seeder{db: db, hasher: hasher}
}

// DefaultRoles are the fixed roles seeded for the system tenant.
func DefaultRoles() []string {
	return []string{entity.RoleSuperAdmin, entity.RoleAdmin, entity.RoleEmployee}
}

// Run seeds everything in a single transaction. It is idempotent: existing
// rows are reused and never overwritten (an existing super admin keeps its
// current password), so re-running creates no duplicates.
func (s *Seeder) Run(ctx context.Context, opts Options) (*Result, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Hash before opening the transaction (bcrypt is slow) and before any
	// DB work, so a hashing failure changes nothing.
	hash, err := s.hasher.HashPassword(opts.SuperAdminPassword)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: hash super admin password: %w", err)
	}

	res := &Result{}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", bootstrapLockKey).Error; err != nil {
			return fmt.Errorf("acquire bootstrap lock: %w", err)
		}

		tenant, created, err := s.ensureTenant(tx, opts.SystemTenantName)
		if err != nil {
			return err
		}
		res.TenantID, res.TenantCreated = tenant.ID, created

		// Login resolves the tenant from the email domain, so the super
		// admin's domain must map to the system tenant.
		domain, _ := emailDomain(opts.SuperAdminEmail) // validated above
		domainCreated, err := s.ensureTenantDomain(tx, tenant.ID, domain)
		if err != nil {
			return err
		}
		res.DomainCreated = domainCreated

		roleIDs := make(map[string]uuid.UUID, 3)
		for _, name := range DefaultRoles() {
			role, created, err := s.ensureRole(tx, tenant.ID, name)
			if err != nil {
				return err
			}
			roleIDs[name] = role.ID
			if created {
				res.RolesCreated = append(res.RolesCreated, name)
			}
		}

		user, created, err := s.ensureSuperAdmin(tx, tenant.ID, opts.SuperAdminEmail, hash)
		if err != nil {
			return err
		}
		res.AdminID, res.AdminCreated = user.ID, created

		added, err := s.ensureUserRole(tx, tenant.ID, user.ID, roleIDs[entity.RoleSuperAdmin])
		if err != nil {
			return err
		}
		res.AdminRoleAdded = added
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Seeder) ensureTenant(tx *gorm.DB, name string) (*entity.Tenant, bool, error) {
	var t entity.Tenant
	err := tx.Where("name = ?", name).Order("created_at ASC").First(&t).Error
	if err == nil {
		return &t, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("lookup system tenant: %w", err)
	}
	now := time.Now().UTC()
	t = entity.Tenant{ID: uuid.New(), Name: name, IsActive: true, CreatedAt: now, UpdatedAt: now}
	if err := tx.Create(&t).Error; err != nil {
		return nil, false, fmt.Errorf("create system tenant: %w", err)
	}
	return &t, true, nil
}

// ensureTenantDomain registers domain on the tenant. The domain is unique
// across all tenants, so if another tenant already owns it, login would not
// resolve to this tenant and we fail loudly instead of silently succeeding.
func (s *Seeder) ensureTenantDomain(tx *gorm.DB, tenantID uuid.UUID, domain string) (bool, error) {
	now := time.Now().UTC()
	td := entity.TenantDomain{ID: uuid.New(), TenantID: tenantID, Domain: domain, CreatedAt: now, UpdatedAt: now}
	res := tx.Table("tenant_domains").Clauses(clause.OnConflict{DoNothing: true}).Create(&td)
	if res.Error != nil {
		return false, fmt.Errorf("create tenant domain %q: %w", domain, res.Error)
	}
	if res.RowsAffected == 1 {
		return true, nil
	}
	var owner entity.TenantDomain
	if err := tx.Table("tenant_domains").Where("domain = ?", domain).First(&owner).Error; err != nil {
		return false, fmt.Errorf("lookup tenant domain %q: %w", domain, err)
	}
	if owner.TenantID != tenantID {
		return false, fmt.Errorf("bootstrap: domain %s already belongs to another tenant", domain)
	}
	return false, nil
}

func (s *Seeder) ensureRole(tx *gorm.DB, tenantID uuid.UUID, name string) (*entity.Role, bool, error) {
	now := time.Now().UTC()
	role := entity.Role{ID: uuid.New(), TenantID: tenantID, Name: name, CreatedAt: now, UpdatedAt: now}
	// Unique key: (tenant_id, name).
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&role)
	if res.Error != nil {
		return nil, false, fmt.Errorf("create role %q: %w", name, res.Error)
	}
	if res.RowsAffected == 1 {
		return &role, true, nil
	}
	var existing entity.Role
	if err := tx.Where("tenant_id = ? AND name = ?", tenantID, name).First(&existing).Error; err != nil {
		return nil, false, fmt.Errorf("lookup role %q: %w", name, err)
	}
	return &existing, false, nil
}

func (s *Seeder) ensureSuperAdmin(tx *gorm.DB, tenantID uuid.UUID, email, hash string) (*entity.User, bool, error) {
	now := time.Now().UTC()
	user := entity.User{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Email:           email,
		FirstName:       superAdminFirstName,
		LastName:        superAdminLastName,
		PasswordHash:    &hash,
		EmailVerifiedAt: &now,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	// Unique key: (tenant_id, email).
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&user)
	if res.Error != nil {
		return nil, false, fmt.Errorf("create super admin: %w", res.Error)
	}
	if res.RowsAffected == 1 {
		return &user, true, nil
	}
	var existing entity.User
	if err := tx.Where("tenant_id = ? AND email = ?", tenantID, email).First(&existing).Error; err != nil {
		return nil, false, fmt.Errorf("lookup super admin: %w", err)
	}
	return &existing, false, nil
}

func (s *Seeder) ensureUserRole(tx *gorm.DB, tenantID, userID, roleID uuid.UUID) (bool, error) {
	now := time.Now().UTC()
	ur := entity.UserRole{ID: uuid.New(), TenantID: tenantID, UserID: userID, RoleID: roleID, CreatedAt: now, UpdatedAt: now}
	// Unique key: (user_id, role_id).
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&ur)
	if res.Error != nil {
		return false, fmt.Errorf("assign super_admin role: %w", res.Error)
	}
	return res.RowsAffected == 1, nil
}
