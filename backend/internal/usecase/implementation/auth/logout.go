package auth

import (
	"context"
	"strings"

	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
)

// LogoutUseCaseImpl implements authusecase.LogoutUseCase.
type LogoutUseCaseImpl struct {
	tokenService     service.TokenService
	hashService      service.HashService
	refreshTokenRepo repository.RefreshTokenRepository
}

var _ authusecase.LogoutUseCase = (*LogoutUseCaseImpl)(nil)

// NewLogoutUseCase constructs a new LogoutUseCaseImpl.
func NewLogoutUseCase(
	tokenService service.TokenService,
	hashService service.HashService,
	refreshTokenRepo repository.RefreshTokenRepository,
) *LogoutUseCaseImpl {
	return &LogoutUseCaseImpl{
		tokenService:     tokenService,
		hashService:      hashService,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Execute revokes the provided refresh token session.
func (u *LogoutUseCaseImpl) Execute(ctx context.Context, input authusecase.LogoutInput) error {
	rawToken := strings.TrimSpace(input.RefreshToken)
	if rawToken == "" {
		return domainerrors.ErrInvalidToken
	}

	if _, err := u.tokenService.ValidateRefreshToken(rawToken); err != nil {
		return domainerrors.ErrInvalidToken
	}

	tokenHash := u.hashService.HashToken(rawToken)

	storedToken, err := u.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || storedToken == nil {
		// Idempotent: token already doesn't exist or is removed
		return nil
	}

	if storedToken.RevokedAt == nil {
		if err := u.refreshTokenRepo.Revoke(ctx, storedToken.ID); err != nil {
			return err
		}
	}

	return nil
}
