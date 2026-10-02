package department

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// UpdateDepartmentInput specifies the department to modify and its new attributes.
type UpdateDepartmentInput struct {
	ID          uuid.UUID
	Name        string
	Description string
}

// UpdateDepartmentUseCase defines the contract for updating a department within the caller's tenant.
type UpdateDepartmentUseCase interface {
	Execute(ctx context.Context, tenantID, actorID uuid.UUID, input UpdateDepartmentInput) (*entity.Department, error)
}
