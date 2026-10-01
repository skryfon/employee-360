package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// TenantReader is the read-only tenant lookup port used to verify that a
// tenant referenced by a token is real and usable.
type TenantReader interface {
	// GetByID returns the tenant with the given id, excluding soft-deleted
	// rows (deleted_at IS NULL). It returns domainerrors.ErrTenantNotFound
	// when no such live tenant exists. The is_active flag is NOT filtered
	// here; callers decide how to treat inactive tenants.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
}

// TenantRepository defines the data access methods for tenants.
type TenantRepository interface {
	TenantReader
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByName(ctx context.Context, name string) (*entity.Tenant, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Tenant, int64, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}
