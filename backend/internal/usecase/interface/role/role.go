// Package role declares the role usecase ports.
package role

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// ListAssignableRolesUseCase lists the caller-tenant roles an admin may assign
// (e.g. when inviting a user). super_admin is never included.
type ListAssignableRolesUseCase interface {
	Execute(ctx context.Context, tenantID uuid.UUID) ([]*entity.Role, error)
}
