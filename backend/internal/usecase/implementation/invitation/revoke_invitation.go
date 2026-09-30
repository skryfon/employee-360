package invitation

import (
	"context"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	invusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/invitation"
)

// RevokeInvitationUseCaseImpl implements invusecase.RevokeInvitationUseCase.
type RevokeInvitationUseCaseImpl struct {
	invitationRepo repository.UserInvitationRepository
}

var _ invusecase.RevokeInvitationUseCase = (*RevokeInvitationUseCaseImpl)(nil)

// NewRevokeInvitationUseCase constructs a RevokeInvitationUseCaseImpl.
func NewRevokeInvitationUseCase(invitationRepo repository.UserInvitationRepository) *RevokeInvitationUseCaseImpl {
	return &RevokeInvitationUseCaseImpl{invitationRepo: invitationRepo}
}

// Execute revokes a pending invitation in the caller's tenant.
func (u *RevokeInvitationUseCaseImpl) Execute(c context.Context, id uuid.UUID) error {
	if err := requireAdmin(c); err != nil {
		return err
	}
	tenantID, err := tenantFromContext(c)
	if err != nil {
		return err
	}
	inv, err := u.invitationRepo.GetByID(c, tenantID, id)
	if err != nil || inv == nil {
		return domainerrors.ErrInvitationNotFound
	}
	if !inv.IsPending() {
		return domainerrors.ErrInvitationNotPending
	}
	return u.invitationRepo.MarkRevoked(c, tenantID, inv.ID, time.Now().UTC())
}
