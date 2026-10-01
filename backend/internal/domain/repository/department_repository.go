package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// DepartmentRepository defines data access methods for tenant-scoped departments.
// Every method takes tenantID explicitly (resolved by the handler from the auth
// context); implementations must not read the tenant from context.Context.
// Create and Update take actorID to record created_by and updated_by audit columns.
type DepartmentRepository interface {
	Create(ctx context.Context, tenantID, actorID uuid.UUID, department *entity.Department) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.Department, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Department, int64, error)
	Update(ctx context.Context, tenantID, actorID uuid.UUID, department *entity.Department) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	ExistsByName(ctx context.Context, tenantID uuid.UUID, name string) (bool, error)
	IsReferenced(ctx context.Context, tenantID, id uuid.UUID) (bool, error)
}
