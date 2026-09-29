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

func TestLogoutUseCase_Success(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	tokenID := uuid.New()
	rawRefreshToken := "active-refresh-token"

	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{
		validateRefreshClaims: &domainservice.RefreshTokenClaims{
			UserID:   userID,
			TenantID: tenantID,
			TokenID:  tokenID,
		},
	}

	hashedToken := hashSvc.HashToken(rawRefreshToken)
	storedRT := &entity.RefreshToken{
		ID:        tokenID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: hashedToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = refreshTokenRepo.Create(context.Background(), storedRT)

	uc := NewLogoutUseCase(tokenSvc, hashSvc, refreshTokenRepo)

	err := uc.Execute(context.Background(), authusecase.LogoutInput{
		RefreshToken: rawRefreshToken,
	})

	if err != nil {
		t.Fatalf("expected successful logout, got: %v", err)
	}

	if len(refreshTokenRepo.revokedTokenIDs) == 0 || refreshTokenRepo.revokedTokenIDs[0] != tokenID {
		t.Errorf("expected token %s to be revoked", tokenID)
	}
	if storedRT.RevokedAt == nil {
		t.Errorf("expected stored token RevokedAt to be set")
	}
}

func TestLogoutUseCase_InvalidOrMissingToken(t *testing.T) {
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{
		validateErr: domainerrors.ErrInvalidToken,
	}

	uc := NewLogoutUseCase(tokenSvc, hashSvc, refreshTokenRepo)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid signature", "corrupted.jwt.token"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := uc.Execute(context.Background(), authusecase.LogoutInput{
				RefreshToken: tc.token,
			})
			if err != domainerrors.ErrInvalidToken {
				t.Errorf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

func TestLogoutUseCase_IdempotentOnNonexistentToken(t *testing.T) {
	tenantID := uuid.New()
	tokenSvc := &mockTokenService{
		validateRefreshClaims: &domainservice.RefreshTokenClaims{
			UserID:   uuid.New(),
			TenantID: tenantID,
			TokenID:  uuid.New(),
		},
	}
	hashSvc := &mockHashService{}
	refreshTokenRepo := newMockRefreshTokenRepository()

	uc := NewLogoutUseCase(tokenSvc, hashSvc, refreshTokenRepo)

	err := uc.Execute(context.Background(), authusecase.LogoutInput{
		RefreshToken: "valid-jwt-but-not-in-db",
	})

	if err != nil {
		t.Fatalf("expected nil error on nonexistent token (idempotent logout), got %v", err)
	}
}
