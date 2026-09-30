package persistence

import (
	"context"
	"errors"
	"strings"

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
		Where("tenant_domains.domain = ? AND tenants.is_active = TRUE", strings.ToLower(domain)).
		First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}
