package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// TenantDomainRepository defines the data access methods for tenant domains.
type TenantDomainRepository interface {
	// FindTenantByDomain finds the tenant associated with the given domain name (used during login to resolve tenant_id).
	FindTenantByDomain(ctx context.Context, domain string) (*entity.Tenant, error)

	Create(ctx context.Context, tenantDomain *entity.TenantDomain) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantDomain, error)
	GetByDomain(ctx context.Context, domain string) (*entity.TenantDomain, error)
	ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantDomain, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
