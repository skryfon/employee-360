package invitation

import (
	"context"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	invusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/invitation"
)

// ListInvitationsUseCaseImpl implements invusecase.ListInvitationsUseCase.
type ListInvitationsUseCaseImpl struct {
	invitationRepo repository.UserInvitationRepository
}

var _ invusecase.ListInvitationsUseCase = (*ListInvitationsUseCaseImpl)(nil)

// NewListInvitationsUseCase constructs a ListInvitationsUseCaseImpl.
func NewListInvitationsUseCase(invitationRepo repository.UserInvitationRepository) *ListInvitationsUseCaseImpl {
	return &ListInvitationsUseCaseImpl{invitationRepo: invitationRepo}
}

// Execute lists invitations for the caller's tenant only.
func (u *ListInvitationsUseCaseImpl) Execute(c context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserInvitation, int64, error) {
	if tenantID == uuid.Nil {
		return nil, 0, domainerrors.ErrUnauthorized
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.invitationRepo.List(c, tenantID, limit, offset)
}
