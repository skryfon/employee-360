package department

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// GetDepartmentInput specifies the department to look up.
type GetDepartmentInput struct {
	ID uuid.UUID
}

// GetDepartmentUseCase defines the contract for retrieving a department by ID within the caller's tenant.
type GetDepartmentUseCase interface {
	Execute(ctx context.Context, tenantID uuid.UUID, input GetDepartmentInput) (*entity.Department, error)
}
