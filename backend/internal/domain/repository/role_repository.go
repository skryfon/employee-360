package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// RoleRepository defines the data access methods for roles.
type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	List(ctx context.Context) ([]*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
}
