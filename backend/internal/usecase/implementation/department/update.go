package department

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
)

type updateDepartmentUseCase struct {
	repo       repository.DepartmentRepository
	auditRepo  repository.AuditRepository
	transactor ucshared.Transactor
}

// NewUpdateDepartmentUseCase creates a new UpdateDepartmentUseCase.
func NewUpdateDepartmentUseCase(
	repo repository.DepartmentRepository,
	auditRepo repository.AuditRepository,
	transactor ucshared.Transactor,
) deptuc.UpdateDepartmentUseCase {
	if transactor == nil {
		transactor = ucshared.NewNopTransactor()
	}
	return &updateDepartmentUseCase{
		repo:       repo,
		auditRepo:  auditRepo,
		transactor: transactor,
	}
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

	var existing *entity.Department
	err := uc.transactor.WithinTransaction(c, func(txCtx context.Context) error {
		var err error
		existing, err = uc.repo.GetByID(txCtx, tenantID, input.ID)
		if err != nil {
			return err
		}

		// If name has changed, check for duplicate name
		if !strings.EqualFold(existing.Name, name) {
			exists, err := uc.repo.ExistsByName(txCtx, tenantID, name)
			if err != nil {
				return err
			}
			if exists {
				return domainerrors.ErrDepartmentNameTaken
			}
		}

		existing.Name = name
		existing.Description = description
		existing.UpdatedAt = time.Now().UTC()

		if err := uc.repo.Update(txCtx, tenantID, actorID, existing); err != nil {
			return err
		}

		return writeAudit(txCtx, uc.auditRepo, tenantID, actorID, existing.ID, auditActionUpdate, map[string]any{
			"name": existing.Name,
		})
	})
	if err != nil {
		return nil, err
	}

	return existing, nil
}
