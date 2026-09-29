package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	authusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/auth"
)

func TestTokenRefreshUseCase_Success(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	family := uuid.New()
	tokenID := uuid.New()
	rawRefreshToken := "valid-refresh-jwt-token"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{
		validateRefreshClaims: &domainservice.RefreshTokenClaims{
			UserID:   userID,
			TenantID: tenantID,
			TokenID:  tokenID,
			Family:   family,
			Roles:    []string{"admin"},
		},
	}

	user := &entity.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "alice@example.com",
		IsActive: true,
		Roles: []entity.Role{
			{ID: uuid.New(), TenantID: tenantID, Name: "admin"},
		},
	}
	_ = userRepo.Create(context.Background(), user)

	hashedToken := hashSvc.HashToken(rawRefreshToken)
	storedRT := &entity.RefreshToken{
		ID:        tokenID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: hashedToken,
		Family:    family,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = refreshTokenRepo.Create(context.Background(), storedRT)

	uc := NewTokenRefreshUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	output, err := uc.Execute(context.Background(), authusecase.TokenRefreshInput{
		RefreshToken: rawRefreshToken,
		IPAddress:    "127.0.0.1",
		UserAgent:    "TestBrowser/1.0",
	})

	if err != nil {
		t.Fatalf("expected successful token refresh, got: %v", err)
	}

	if output.AccessToken == "" || output.RefreshToken == "" {
		t.Fatalf("expected non-empty access and refresh tokens")
	}

	// Verify old token was revoked
	if storedRT.RevokedAt == nil {
		t.Errorf("expected old token to be marked as revoked")
	}

	// Verify new token was created with same family
	newHashedToken := hashSvc.HashToken(output.RefreshToken)
	newStoredRT, ok := refreshTokenRepo.tokensByHash[newHashedToken]
	if !ok {
		t.Fatalf("expected new refresh token to be persisted")
	}
	if newStoredRT.Family != family {
		t.Errorf("expected family %s to be preserved, got %s", family, newStoredRT.Family)
	}
	if newStoredRT.ID == tokenID {
		t.Errorf("expected new token to have a different ID than old token")
	}
}

func TestTokenRefreshUseCase_TokenReuseDetected(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	family := uuid.New()
	tokenID := uuid.New()
	rawRefreshToken := "reused-refresh-token"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{
		validateRefreshClaims: &domainservice.RefreshTokenClaims{
			UserID:   userID,
			TenantID: tenantID,
			TokenID:  tokenID,
			Family:   family,
		},
	}

	hashedToken := hashSvc.HashToken(rawRefreshToken)
	revokedAt := time.Now().Add(-10 * time.Minute)
	storedRT := &entity.RefreshToken{
		ID:        tokenID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: hashedToken,
		Family:    family,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		RevokedAt: &revokedAt,
	}
	_ = refreshTokenRepo.Create(context.Background(), storedRT)

	uc := NewTokenRefreshUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	_, err := uc.Execute(context.Background(), authusecase.TokenRefreshInput{
		RefreshToken: rawRefreshToken,
	})

	if err != domainerrors.ErrTokenRevoked {
		t.Errorf("expected ErrTokenRevoked on reuse, got: %v", err)
	}

	// Verify entire family was revoked
	if len(refreshTokenRepo.revokedFamilies) == 0 || refreshTokenRepo.revokedFamilies[0] != family {
		t.Errorf("expected entire family %s to be revoked on token reuse", family)
	}
}

func TestTokenRefreshUseCase_ExpiredToken(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	family := uuid.New()
	tokenID := uuid.New()
	rawRefreshToken := "expired-refresh-token"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{
		validateRefreshClaims: &domainservice.RefreshTokenClaims{
			UserID:   userID,
			TenantID: tenantID,
			TokenID:  tokenID,
			Family:   family,
		},
	}

	hashedToken := hashSvc.HashToken(rawRefreshToken)
	storedRT := &entity.RefreshToken{
		ID:        tokenID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: hashedToken,
		Family:    family,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}
	_ = refreshTokenRepo.Create(context.Background(), storedRT)

	uc := NewTokenRefreshUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	_, err := uc.Execute(context.Background(), authusecase.TokenRefreshInput{
		RefreshToken: rawRefreshToken,
	})

	if err != domainerrors.ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got: %v", err)
	}
}

func TestTokenRefreshUseCase_InvalidToken(t *testing.T) {
	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{
		validateErr: domainerrors.ErrInvalidToken,
	}

	uc := NewTokenRefreshUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	_, err := uc.Execute(context.Background(), authusecase.TokenRefreshInput{
		RefreshToken: "invalid-token",
	})

	if err != domainerrors.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
