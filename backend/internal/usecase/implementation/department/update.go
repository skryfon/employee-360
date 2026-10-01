package department

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

type updateDepartmentUseCase struct {
	repo repository.DepartmentRepository
}

// NewUpdateDepartmentUseCase creates a new UpdateDepartmentUseCase.
func NewUpdateDepartmentUseCase(repo repository.DepartmentRepository) deptuc.UpdateDepartmentUseCase {
	return &updateDepartmentUseCase{repo: repo}
}

func (uc *updateDepartmentUseCase) Execute(c context.Context, tenantID, actorID uuid.UUID, input deptuc.UpdateDepartmentInput) (*entity.Department, error) {
	if tenantID == uuid.Nil || actorID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}
	if input.ID == uuid.Nil {
		return nil, domainerrors.ErrDepartmentNotFound
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrDepartmentNameRequired
	}
	if len(name) > 100 {
		return nil, ErrDepartmentNameTooLong
	}

	description := strings.TrimSpace(input.Description)
	if len(description) > 500 {
		return nil, ErrDescriptionTooLong
	}

	existing, err := uc.repo.GetByID(c, tenantID, input.ID)
	if err != nil {
		return nil, err
	}

	// If name has changed, check for duplicate name
	if !strings.EqualFold(existing.Name, name) {
		exists, err := uc.repo.ExistsByName(c, tenantID, name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domainerrors.ErrDepartmentNameTaken
		}
	}

	existing.Name = name
	existing.Description = description
	existing.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(c, tenantID, actorID, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
