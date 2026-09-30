package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// UserRoleRepository defines the data access methods for user-role associations.
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userRole *entity.UserRole) error
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetRolesByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error)
	GetUserRolesByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.UserRole, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
