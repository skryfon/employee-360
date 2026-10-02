package department

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// ListDepartmentsInput holds pagination parameters for listing departments.
type ListDepartmentsInput struct {
	Page     int
	PageSize int
	IsActive *bool // nil returns all departments
}

// ListDepartmentsOutput contains paginated departments and pagination metadata.
type ListDepartmentsOutput struct {
	Departments []*entity.Department
	Total       int64
	Page        int
	PageSize    int
}

// ListDepartmentsUseCase defines the contract for listing departments within the caller's tenant.
type ListDepartmentsUseCase interface {
	Execute(ctx context.Context, tenantID uuid.UUID, input ListDepartmentsInput) (*ListDepartmentsOutput, error)
}
