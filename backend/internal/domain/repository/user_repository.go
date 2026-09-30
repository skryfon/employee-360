package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// UserRepository defines the data access methods for users.
//
// Every method that needs tenant scoping takes tenantID as an explicit
// parameter rather than resolving it from ctx itself -- the repository layer
// has zero knowledge of how tenant_id was resolved (handler middleware, a
// token row, a domain lookup, ...). Callers (usecases) are responsible for
// obtaining a trustworthy tenantID and passing it in; Create is the one
// exception, since it already receives a full *entity.User with TenantID set
// by the caller.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.User, error)
	GetByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error)
	GetByIDWithRoles(ctx context.Context, tenantID, id uuid.UUID) (*entity.User, error)
	GetByTenantAndEmailWithRoles(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error)
	Update(ctx context.Context, tenantID uuid.UUID, user *entity.User) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.User, int64, error)
}
