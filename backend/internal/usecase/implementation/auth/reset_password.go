package auth

import (
	"context"
	"strings"
	"time"

	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
)

// ResetPasswordUseCaseImpl implements authusecase.ResetPasswordUseCase.
type ResetPasswordUseCaseImpl struct {
	userRepo          repository.UserRepository
	passwordResetRepo repository.PasswordResetRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	hashService       service.HashService
	transactor        ucshared.Transactor
}

var _ authusecase.ResetPasswordUseCase = (*ResetPasswordUseCaseImpl)(nil)

// NewResetPasswordUseCase constructs a new ResetPasswordUseCaseImpl.
func NewResetPasswordUseCase(
	userRepo repository.UserRepository,
	passwordResetRepo repository.PasswordResetRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	hashService service.HashService,
	transactor ucshared.Transactor,
) *ResetPasswordUseCaseImpl {
	return &ResetPasswordUseCaseImpl{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		refreshTokenRepo:  refreshTokenRepo,
		hashService:       hashService,
		transactor:        transactor,
	}
}

// Execute consumes a valid reset token, updates the password, and revokes all active sessions.
func (u *ResetPasswordUseCaseImpl) Execute(ctx context.Context, input authusecase.ResetPasswordInput) error {
	plainToken := strings.TrimSpace(input.Token)
	if plainToken == "" {
		return domainerrors.ErrInvalidToken
	}

	if len(input.NewPassword) < 8 {
		return domainerrors.ErrInvalidPassword
	}

	tokenHash := u.hashService.HashToken(plainToken)

	resetToken, err := u.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || resetToken == nil {
		return domainerrors.ErrInvalidToken
	}

	if !resetToken.IsValid() {
		return domainerrors.ErrInvalidToken
	}

	user, err := u.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil || user == nil {
		return domainerrors.ErrUserNotFound
	}

	if !user.IsActive {
		return domainerrors.ErrUserInactive
	}

	hashedPassword, err := u.hashService.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.passwordResetRepo.MarkAsUsed(txCtx, resetToken.ID); err != nil {
			return err
		}

		if err := u.passwordResetRepo.InvalidateAllForUser(txCtx, user.ID); err != nil {
			return err
		}

		now := time.Now().UTC()
		user.PasswordHash = &hashedPassword
		user.UpdatedAt = now

		if err := u.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Invalidate all existing refresh tokens for the user across all sessions
		if err := u.refreshTokenRepo.RevokeAllForUser(txCtx, user.ID); err != nil {
			return err
		}

		return nil
	})
}
