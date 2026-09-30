package invitation

import (
	"context"

	"github.com/skryfon/employee360/backend/internal/domain/entity"
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
func (u *ListInvitationsUseCaseImpl) Execute(c context.Context, limit, offset int) ([]*entity.UserInvitation, int64, error) {
	if err := requireAdmin(c); err != nil {
		return nil, 0, err
	}
	tenantID, err := tenantFromContext(c)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.invitationRepo.List(c, tenantID, limit, offset)
}
