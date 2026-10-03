package position

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	posuc "github.com/skryfon/employee360/backend/internal/usecase/interface/position"
)

type updatePositionUseCase struct {
	repo       repository.PositionRepository
	auditRepo  repository.AuditRepository
	transactor ucshared.Transactor
}

// NewUpdatePositionUseCase creates a new UpdatePositionUseCase.
func NewUpdatePositionUseCase(
	repo repository.PositionRepository,
	auditRepo repository.AuditRepository,
	transactor ucshared.Transactor,
) posuc.UpdatePositionUseCase {
	if transactor == nil {
		transactor = ucshared.NewNopTransactor()
	}
	return &updatePositionUseCase{
		repo:       repo,
		auditRepo:  auditRepo,
		transactor: transactor,
	}
}

func (uc *updatePositionUseCase) Execute(c context.Context, tenantID, actorID uuid.UUID, input posuc.UpdatePositionInput) (*entity.Position, error) {
	if tenantID == uuid.Nil || actorID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}
	if input.ID == uuid.Nil {
		return nil, domainerrors.ErrPositionNotFound
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrPositionNameRequired
	}
	if utf8.RuneCountInString(name) > 100 {
		return nil, ErrPositionNameTooLong
	}

	description := strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(description) > 500 {
		return nil, ErrDescriptionTooLong
	}

	var existing *entity.Position
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
				return domainerrors.ErrPositionNameTaken
			}
		}

		wasActive := existing.IsActive
		existing.Name = name
		existing.Description = description
		if input.IsActive != nil {
			existing.IsActive = *input.IsActive
		}
		existing.UpdatedAt = time.Now().UTC()

		if err := uc.repo.Update(txCtx, tenantID, actorID, existing); err != nil {
			return err
		}

		if err := writeAudit(txCtx, uc.auditRepo, tenantID, actorID, existing.ID, auditActionUpdate, map[string]any{
			"name":      existing.Name,
			"is_active": existing.IsActive,
		}); err != nil {
			return err
		}
		if wasActive == existing.IsActive {
			return nil
		}
		action := auditActionDeactivate
		if existing.IsActive {
			action = auditActionActivate
		}
		return writeAudit(txCtx, uc.auditRepo, tenantID, actorID, existing.ID, action, map[string]any{
			"name": existing.Name,
		})
	})
	if err != nil {
		return nil, err
	}

	return existing, nil
}
