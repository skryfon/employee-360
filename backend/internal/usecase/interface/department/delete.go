package department

import (
	"context"

	"github.com/google/uuid"
)

// DeleteDepartmentInput specifies the department to remove.
type DeleteDepartmentInput struct {
	ID uuid.UUID
}

// DeleteDepartmentUseCase defines the contract for deleting a department within the caller's tenant.
type DeleteDepartmentUseCase interface {
	Execute(ctx context.Context, tenantID, actorID uuid.UUID, input DeleteDepartmentInput) error
}
