package department

import (
	"context"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

type deleteDepartmentUseCase struct {
	repo repository.DepartmentRepository
}

// NewDeleteDepartmentUseCase creates a new DeleteDepartmentUseCase.
func NewDeleteDepartmentUseCase(repo repository.DepartmentRepository) deptuc.DeleteDepartmentUseCase {
	return &deleteDepartmentUseCase{repo: repo}
}

func (uc *deleteDepartmentUseCase) Execute(c context.Context, tenantID uuid.UUID, input deptuc.DeleteDepartmentInput) error {
	if tenantID == uuid.Nil {
		return domainerrors.ErrUnauthorized
	}
	if input.ID == uuid.Nil {
		return domainerrors.ErrDepartmentNotFound
	}

	// Verify department exists in the caller's tenant
	_, err := uc.repo.GetByID(c, tenantID, input.ID)
	if err != nil {
		return err
	}

	// Verify department is not referenced by active users or invitations
	isReferenced, err := uc.repo.IsReferenced(c, tenantID, input.ID)
	if err != nil {
		return err
	}
	if isReferenced {
		return domainerrors.ErrDepartmentInUse
	}

	return uc.repo.Delete(c, tenantID, input.ID)
}
