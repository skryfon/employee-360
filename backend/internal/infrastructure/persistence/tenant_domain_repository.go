package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
)

// gormTenantDomainRepository is a GORM-backed adapter implementing
// repository.TenantDomainRepository.
//
// tenant_domains is the tenant-resolution table itself: FindTenantByDomain and
// GetByDomain are the (only) pre-authentication lookups that derive a tenant
// from server-side data (the email's domain), so they are intentionally not
// scoped by a context tenant.
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
	if err := r.db.WithContext(c).
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

func (r *gormTenantDomainRepository) Create(c context.Context, td *entity.TenantDomain) error {
	td.Domain = strings.ToLower(td.Domain)
	return r.db.WithContext(c).Table("tenant_domains").Create(td).Error
}

func (r *gormTenantDomainRepository) GetByID(c context.Context, id uuid.UUID) (*entity.TenantDomain, error) {
	var td entity.TenantDomain
	if err := r.db.WithContext(c).Table("tenant_domains").Where("id = ?", id).First(&td).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &td, nil
}

func (r *gormTenantDomainRepository) GetByDomain(c context.Context, domain string) (*entity.TenantDomain, error) {
	var td entity.TenantDomain
	if err := r.db.WithContext(c).Table("tenant_domains").Where("domain = ?", strings.ToLower(domain)).First(&td).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &td, nil
}

func (r *gormTenantDomainRepository) ListByTenantID(c context.Context, tenantID uuid.UUID) ([]*entity.TenantDomain, error) {
	var list []*entity.TenantDomain
	if err := r.db.WithContext(c).Table("tenant_domains").Where("tenant_id = ?", tenantID).Order("domain").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormTenantDomainRepository) Delete(c context.Context, id uuid.UUID) error {
	res := r.db.WithContext(c).Table("tenant_domains").Where("id = ?", id).Delete(&entity.TenantDomain{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainerrors.ErrNotFound
	}
	return nil
}
