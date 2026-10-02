package department

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

type getDepartmentUseCase struct {
	repo repository.DepartmentRepository
}

// NewGetDepartmentUseCase creates a new GetDepartmentUseCase.
func NewGetDepartmentUseCase(repo repository.DepartmentRepository) deptuc.GetDepartmentUseCase {
	return &getDepartmentUseCase{repo: repo}
}

func (uc *getDepartmentUseCase) Execute(c context.Context, tenantID uuid.UUID, input deptuc.GetDepartmentInput) (*entity.Department, error) {
	if tenantID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}
	if input.ID == uuid.Nil {
		return nil, domainerrors.ErrDepartmentNotFound
	}
	return uc.repo.GetByID(c, tenantID, input.ID)
}
