package department

import (
	"context"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

type listDepartmentsUseCase struct {
	repo repository.DepartmentRepository
}

// NewListDepartmentsUseCase creates a new ListDepartmentsUseCase.
func NewListDepartmentsUseCase(repo repository.DepartmentRepository) deptuc.ListDepartmentsUseCase {
	return &listDepartmentsUseCase{repo: repo}
}

func (uc *listDepartmentsUseCase) Execute(c context.Context, tenantID uuid.UUID, input deptuc.ListDepartmentsInput) (*deptuc.ListDepartmentsOutput, error) {
	if tenantID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}

	page := input.Page
	if page < 1 {
		page = 1
	}

	pageSize := input.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	items, total, err := uc.repo.List(c, tenantID, input.IsActive, limit, offset)
	if err != nil {
		return nil, err
	}

	return &deptuc.ListDepartmentsOutput{
		Departments: items,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}
