package position

import (
	"context"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	posuc "github.com/skryfon/employee360/backend/internal/usecase/interface/position"
)

type deletePositionUseCase struct {
	repo       repository.PositionRepository
	auditRepo  repository.AuditRepository
	transactor ucshared.Transactor
}

// NewDeletePositionUseCase creates a new DeletePositionUseCase.
func NewDeletePositionUseCase(
	repo repository.PositionRepository,
	auditRepo repository.AuditRepository,
	transactor ucshared.Transactor,
) posuc.DeletePositionUseCase {
	if transactor == nil {
		transactor = ucshared.NewNopTransactor()
	}
	return &deletePositionUseCase{
		repo:       repo,
		auditRepo:  auditRepo,
		transactor: transactor,
	}
}

func (uc *deletePositionUseCase) Execute(c context.Context, tenantID, actorID uuid.UUID, input posuc.DeletePositionInput) error {
	if tenantID == uuid.Nil || actorID == uuid.Nil {
		return domainerrors.ErrUnauthorized
	}
	if input.ID == uuid.Nil {
		return domainerrors.ErrPositionNotFound
	}

	return uc.transactor.WithinTransaction(c, func(txCtx context.Context) error {
		// Lock the position row so a concurrent invite cannot attach it
		// between the reference check and the soft delete.
		pos, err := uc.repo.GetByIDForUpdate(txCtx, tenantID, input.ID)
		if err != nil {
			return err
		}

		// Verify position is not referenced by active users or invitations
		isReferenced, err := uc.repo.IsReferenced(txCtx, tenantID, input.ID)
		if err != nil {
			return err
		}
		if isReferenced {
			return domainerrors.ErrPositionInUse
		}

		if err := uc.repo.Delete(txCtx, tenantID, input.ID, actorID); err != nil {
			return err
		}

		return writeAudit(txCtx, uc.auditRepo, tenantID, actorID, pos.ID, auditActionDelete, map[string]any{
			"name": pos.Name,
		})
	})
}
