package invitation

import (
	"context"
	"errors"
	"strings"
	"time"

	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
	invusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/invitation"
)

// ValidateInvitationUseCaseImpl implements invusecase.ValidateInvitationUseCase.
type ValidateInvitationUseCaseImpl struct {
	userRepo       repository.UserRepository
	invitationRepo repository.UserInvitationRepository
	hashService    service.HashService
}

var _ invusecase.ValidateInvitationUseCase = (*ValidateInvitationUseCaseImpl)(nil)

// NewValidateInvitationUseCase constructs a ValidateInvitationUseCaseImpl.
func NewValidateInvitationUseCase(
	userRepo repository.UserRepository,
	invitationRepo repository.UserInvitationRepository,
	hashService service.HashService,
) *ValidateInvitationUseCaseImpl {
	return &ValidateInvitationUseCaseImpl{userRepo: userRepo, invitationRepo: invitationRepo, hashService: hashService}
}

// Execute checks if the invitation token is valid, pending, and unexpired.
func (u *ValidateInvitationUseCaseImpl) Execute(ctx context.Context, token string) (*invtypes.ValidateInvitationResponse, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil, domainerrors.ErrInvalidToken
	}
	tokenHash := u.hashService.HashToken(trimmed)
	now := time.Now().UTC()

	inv, err := u.invitationRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || inv == nil || !inv.IsUsable(now) {
		return nil, domainerrors.ErrInvalidToken
	}

	user, err := u.userRepo.GetByTenantAndEmail(ctx, inv.TenantID, inv.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) || errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, domainerrors.ErrInvalidToken
		}
		return nil, err
	}
	if user == nil || user.IsActive {
		return nil, domainerrors.ErrInvalidToken
	}

	return &invtypes.ValidateInvitationResponse{
		Valid: true,
		Email: inv.Email,
	}, nil
}
