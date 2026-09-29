package repository

import (
	"context"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// TenantDomainRepository defines the data access methods for tenant domains.
type TenantDomainRepository interface {
	// FindTenantByDomain finds the tenant associated with the given domain name (used during login to resolve tenant_id).
	FindTenantByDomain(ctx context.Context, domain string) (*entity.Tenant, error)
}
