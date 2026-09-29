package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/domain/service"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
)

const (
	dummyBcryptHash      = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	defaultRefreshExpiry = 7 * 24 * time.Hour
)

// LoginUseCaseImpl implements authusecase.LoginUseCase.
type LoginUseCaseImpl struct {
	userRepo         repository.UserRepository
	tokenService     service.TokenService
	hashService      service.HashService
	refreshTokenRepo repository.RefreshTokenRepository
	refreshExpiry    time.Duration
}

var _ authusecase.LoginUseCase = (*LoginUseCaseImpl)(nil)

// NewLoginUseCase constructs a new LoginUseCaseImpl.
func NewLoginUseCase(
	userRepo repository.UserRepository,
	tokenService service.TokenService,
	hashService service.HashService,
	refreshTokenRepo repository.RefreshTokenRepository,
) *LoginUseCaseImpl {
	return &LoginUseCaseImpl{
		userRepo:         userRepo,
		tokenService:     tokenService,
		hashService:      hashService,
		refreshTokenRepo: refreshTokenRepo,
		refreshExpiry:    defaultRefreshExpiry,
	}
}

// Execute authenticates a user by email and password and returns a token pair.
func (u *LoginUseCaseImpl) Execute(ctx context.Context, input authusecase.LoginInput) (*authusecase.LoginOutput, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := input.Password

	if email == "" || password == "" {
		_ = u.hashService.ComparePassword(dummyBcryptHash, "dummy")
		return nil, domainerrors.ErrInvalidCredentials
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		_ = u.hashService.ComparePassword(dummyBcryptHash, "dummy")
		return nil, domainerrors.ErrInvalidCredentials
	}

	user, err := u.userRepo.GetByEmailWithRoles(ctx, email)
	if err != nil || user == nil {
		_ = u.hashService.ComparePassword(dummyBcryptHash, password)
		return nil, domainerrors.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domainerrors.ErrUserInactive
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		_ = u.hashService.ComparePassword(dummyBcryptHash, password)
		return nil, domainerrors.ErrInvalidCredentials
	}

	if err := u.hashService.ComparePassword(*user.PasswordHash, password); err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	roleNames := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
	}

	refreshTokenID := uuid.New()
	family := uuid.New()

	accessClaims := service.AccessTokenClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Roles:    roleNames,
		Email:    user.Email,
	}

	refreshClaims := service.RefreshTokenClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		TokenID:  refreshTokenID,
		Family:   family,
		Roles:    roleNames,
	}

	tokenPair, err := u.tokenService.GenerateTokenPair(accessClaims, refreshClaims)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiry := now.Add(u.refreshExpiry)
	tokenHash := u.hashService.HashToken(tokenPair.RefreshToken)

	rt := &entity.RefreshToken{
		ID:        refreshTokenID,
		TenantID:  user.TenantID,
		UserID:    user.ID,
		TokenHash: tokenHash,
		Family:    family,
		ExpiresAt: expiry,
		IPAddress: input.IPAddress,
		UserAgent: input.UserAgent,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	user.LastLoginAt = &now
	_ = u.userRepo.Update(ctx, user)

	return &authusecase.LoginOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User:         user,
	}, nil
}
