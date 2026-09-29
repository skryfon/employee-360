package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
)

func TestJWTService_AccessTokenFlow(t *testing.T) {
	jwtSvc := NewJWTService("test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	tenantID := uuid.New()
	claims := domainservice.AccessTokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}

	tokenStr, expiresAt, err := jwtSvc.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("unexpected error generating access token: %v", err)
	}

	if tokenStr == "" {
		t.Fatalf("generated token string should not be empty")
	}

	if expiresAt.Before(time.Now()) {
		t.Fatalf("expiration should be in the future")
	}

	parsedClaims, err := jwtSvc.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if parsedClaims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, parsedClaims.UserID)
	}
	if parsedClaims.TenantID != tenantID {
		t.Errorf("expected TenantID %s, got %s", tenantID, parsedClaims.TenantID)
	}
	if parsedClaims.Email != "test@example.com" {
		t.Errorf("expected Email test@example.com, got %s", parsedClaims.Email)
	}
	if len(parsedClaims.Roles) != 1 || parsedClaims.Roles[0] != "admin" {
		t.Errorf("expected roles [admin], got %v", parsedClaims.Roles)
	}
}

func TestJWTService_RefreshTokenFlow(t *testing.T) {
	jwtSvc := NewJWTService("test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	tenantID := uuid.New()
	tokenID := uuid.New()
	family := uuid.New()

	claims := domainservice.RefreshTokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		TokenID:  tokenID,
		Family:   family,
		Roles:    []string{"employee"},
	}

	tokenStr, expiresAt, err := jwtSvc.GenerateRefreshToken(claims)
	if err != nil {
		t.Fatalf("unexpected error generating refresh token: %v", err)
	}

	if tokenStr == "" {
		t.Fatalf("generated token string should not be empty")
	}

	if expiresAt.Before(time.Now()) {
		t.Fatalf("expiration should be in the future")
	}

	parsedClaims, err := jwtSvc.ValidateRefreshToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if parsedClaims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, parsedClaims.UserID)
	}
	if parsedClaims.TenantID != tenantID {
		t.Errorf("expected TenantID %s, got %s", tenantID, parsedClaims.TenantID)
	}
	if parsedClaims.TokenID != tokenID {
		t.Errorf("expected TokenID %s, got %s", tokenID, parsedClaims.TokenID)
	}
	if parsedClaims.Family != family {
		t.Errorf("expected Family %s, got %s", family, parsedClaims.Family)
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	jwtSvc := NewJWTService("test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	tenantID := uuid.New()

	accessClaims := domainservice.AccessTokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    "pair@example.com",
		Roles:    []string{"super_admin"},
	}

	refreshClaims := domainservice.RefreshTokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		TokenID:  uuid.New(),
		Family:   uuid.New(),
		Roles:    []string{"super_admin"},
	}

	pair, err := jwtSvc.GenerateTokenPair(accessClaims, refreshClaims)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("tokens in pair should not be empty")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token type Bearer, got %s", pair.TokenType)
	}

	validatedAccess, err := jwtSvc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate access token from pair: %v", err)
	}
	if validatedAccess.Email != "pair@example.com" {
		t.Errorf("expected email pair@example.com, got %s", validatedAccess.Email)
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	jwtSvc := NewJWTService("test-secret-key-32-bytes-long!", -1*time.Minute, -1*time.Minute)

	tokenStr, _, err := jwtSvc.GenerateAccessToken(domainservice.AccessTokenClaims{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		Email:    "expired@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = jwtSvc.ValidateAccessToken(tokenStr)
	if err != domainerrors.ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestJWTService_InvalidSignature(t *testing.T) {
	jwtSvc1 := NewJWTService("secret-1-32-bytes-long-padding!", 15*time.Minute, 7*24*time.Hour)
	jwtSvc2 := NewJWTService("secret-2-32-bytes-long-padding!", 15*time.Minute, 7*24*time.Hour)

	tokenStr, _, err := jwtSvc1.GenerateAccessToken(domainservice.AccessTokenClaims{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		Email:    "test@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = jwtSvc2.ValidateAccessToken(tokenStr)
	if err != domainerrors.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken on signature mismatch, got %v", err)
	}
}
