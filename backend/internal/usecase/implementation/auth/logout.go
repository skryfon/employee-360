package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	authtypes "github.com/skryfon/employee360/backend/internal/types/auth"
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
func (u *LogoutUseCaseImpl) Execute(ctx context.Context, tenantID, userID uuid.UUID, input authtypes.LogoutRequest) error {
	rawToken := strings.TrimSpace(input.RefreshToken)
	if rawToken == "" {
		return domainerrors.ErrInvalidToken
	}

	claims, err := u.tokenService.ValidateRefreshToken(rawToken)
	if err != nil || claims == nil {
		return domainerrors.ErrInvalidToken
	}

	// The presented refresh token must belong to the authenticated caller.
	if claims.UserID != userID || claims.TenantID != tenantID {
		return domainerrors.ErrInvalidToken
	}

	tokenHash := u.hashService.HashToken(rawToken)

	storedToken, err := u.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || storedToken == nil {
		// Idempotent: token already doesn't exist or is removed
		return nil
	}

	if storedToken.UserID != userID || storedToken.TenantID != tenantID {
		return domainerrors.ErrInvalidToken
	}

	if storedToken.RevokedAt == nil {
		if err := u.refreshTokenRepo.Revoke(ctx, storedToken.ID); err != nil {
			return err
		}
	}

	return nil
}
