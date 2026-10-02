package department

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// CreateDepartmentInput holds validated data required to create a department.
type CreateDepartmentInput struct {
	Name        string
	Description string
	IsActive    *bool // nil defaults to true
}

// CreateDepartmentUseCase defines the contract for creating a department within the caller's tenant.
type CreateDepartmentUseCase interface {
	Execute(ctx context.Context, tenantID, actorID uuid.UUID, input CreateDepartmentInput) (*entity.Department, error)
}
