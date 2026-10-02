package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/internal/infrastructure/database"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
)

// gormTenantDomainRepository is a GORM-backed adapter implementing
// repository.TenantDomainRepository.
//
// tenant_domains is the tenant-resolution table itself: FindTenantByDomain is
// the (only) pre-authentication lookup that derives a tenant from server-side
// data (the email's domain), so it is intentionally not scoped by a context
// tenant. No other methods exist: unscoped mutations/reads are deliberately
// not exposed.
type gormTenantDomainRepository struct {
	db *gorm.DB
}

// NewGormTenantDomainRepository constructs a GORM-backed TenantDomainRepository.
func NewGormTenantDomainRepository(db *gorm.DB) repository.TenantDomainRepository {
	return &gormTenantDomainRepository{db: db}
}

var _ repository.TenantDomainRepository = (*gormTenantDomainRepository)(nil)

// FindTenantByDomain returns the active tenant that owns the given email domain.
func (r *gormTenantDomainRepository) FindTenantByDomain(c context.Context, domain string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	if err := database.DBFromContext(c, r.db).
		Table("tenants").
		Joins("JOIN tenant_domains ON tenant_domains.tenant_id = tenants.id").
		Where("tenant_domains.domain = ? AND tenants.is_active = TRUE AND tenants.deleted_at IS NULL AND tenant_domains.deleted_at IS NULL", strings.ToLower(domain)).
		First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// DomainBelongsToTenant reports whether domain is a live (not soft-deleted)
// domain of the given active, not soft-deleted tenant. tenant_domains has no
// is_active/verified column.
func (r *gormTenantDomainRepository) DomainBelongsToTenant(c context.Context, tenantID uuid.UUID, domain string) (bool, error) {
	var n int64
	if err := database.DBFromContext(c, r.db).
		Table("tenant_domains").
		Joins("JOIN tenants ON tenants.id = tenant_domains.tenant_id").
		Where("tenant_domains.tenant_id = ? AND tenant_domains.domain = ? AND tenant_domains.deleted_at IS NULL AND tenants.is_active = TRUE AND tenants.deleted_at IS NULL",
			tenantID, strings.ToLower(strings.TrimSpace(domain))).
		Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
