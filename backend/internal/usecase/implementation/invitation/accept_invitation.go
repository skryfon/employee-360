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
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	invusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/invitation"
)

// AcceptInvitationUseCaseImpl implements invusecase.AcceptInvitationUseCase.
type AcceptInvitationUseCaseImpl struct {
	userRepo       repository.UserRepository
	invitationRepo repository.UserInvitationRepository
	hashService    service.HashService
	transactor     ucshared.Transactor
}

var _ invusecase.AcceptInvitationUseCase = (*AcceptInvitationUseCaseImpl)(nil)

// NewAcceptInvitationUseCase constructs an AcceptInvitationUseCaseImpl.
func NewAcceptInvitationUseCase(
	userRepo repository.UserRepository,
	invitationRepo repository.UserInvitationRepository,
	hashService service.HashService,
	transactor ucshared.Transactor,
) *AcceptInvitationUseCaseImpl {
	return &AcceptInvitationUseCaseImpl{userRepo: userRepo, invitationRepo: invitationRepo, hashService: hashService, transactor: transactor}
}

// Execute consumes the token, sets the password (mandatory for every role) and activates the user.
// The tenant is derived from the invitation row found by token hash, never from input.
func (u *AcceptInvitationUseCaseImpl) Execute(c context.Context, req invtypes.AcceptInvitationRequest) error {
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return domainerrors.ErrInvalidToken
	}
	if len(req.Password) < minPasswordLength {
		return domainerrors.ErrInvalidPassword
	}

	now := time.Now().UTC()
	inv, err := u.invitationRepo.GetByTokenHash(c, u.hashService.HashToken(token))
	if err != nil || inv == nil || !inv.IsUsable(now) {
		return domainerrors.ErrInvalidToken
	}

	user, err := u.userRepo.GetByTenantAndEmail(c, inv.TenantID, inv.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) || errors.Is(err, domainerrors.ErrUserNotFound) {
			return domainerrors.ErrInvalidToken
		}
		return err
	}
	if user == nil || user.IsActive {
		return domainerrors.ErrInvalidToken
	}

	hashed, err := u.hashService.HashPassword(req.Password)
	if err != nil {
		return err
	}
	user.PasswordHash = &hashed
	user.IsActive = true
	user.EmailVerifiedAt = &now
	user.UpdatedAt = now

	return u.transactor.WithinTransaction(c, func(txCtx context.Context) error {
		if err := u.invitationRepo.MarkAccepted(txCtx, inv.TenantID, inv.ID, now); err != nil {
			return err
		}
		return u.userRepo.Update(txCtx, inv.TenantID, user)
	})
}
