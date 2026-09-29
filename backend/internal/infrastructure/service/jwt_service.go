package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
	defaultIssuer    = "employee360"
)

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type accessPayload struct {
	UserID    string   `json:"user_id"`
	TenantID  string   `json:"tenant_id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"token_type"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
	Issuer    string   `json:"iss,omitempty"`
}

type refreshPayload struct {
	UserID    string   `json:"user_id"`
	TenantID  string   `json:"tenant_id"`
	TokenID   string   `json:"token_id"`
	Family    string   `json:"family"`
	Roles     []string `json:"roles,omitempty"`
	TokenType string   `json:"token_type"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
	Issuer    string   `json:"iss,omitempty"`
}

type jwtService struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

// NewJWTService creates a new JWT-based TokenService implementation.
func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) domainservice.TokenService {
	return &jwtService{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        defaultIssuer,
	}
}

// GenerateAccessToken generates a signed access JWT for a user.
func (s *jwtService) GenerateAccessToken(claims domainservice.AccessTokenClaims) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessExpiry)

	payload := accessPayload{
		UserID:    claims.UserID.String(),
		TenantID:  claims.TenantID.String(),
		Email:     claims.Email,
		Roles:     claims.Roles,
		TokenType: tokenTypeAccess,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
		Issuer:    s.issuer,
	}

	tokenStr, err := s.signToken(payload)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

// GenerateRefreshToken generates a signed refresh JWT for a session.
func (s *jwtService) GenerateRefreshToken(claims domainservice.RefreshTokenClaims) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.refreshExpiry)

	payload := refreshPayload{
		UserID:    claims.UserID.String(),
		TenantID:  claims.TenantID.String(),
		TokenID:   claims.TokenID.String(),
		Family:    claims.Family.String(),
		Roles:     claims.Roles,
		TokenType: tokenTypeRefresh,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
		Issuer:    s.issuer,
	}

	tokenStr, err := s.signToken(payload)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

// GenerateTokenPair generates both an access and refresh token.
func (s *jwtService) GenerateTokenPair(accessClaims domainservice.AccessTokenClaims, refreshClaims domainservice.RefreshTokenClaims) (*domainservice.TokenPair, error) {
	accessToken, accessExpiresAt, err := s.GenerateAccessToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := s.GenerateRefreshToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domainservice.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken parses and validates a signed access JWT string.
func (s *jwtService) ValidateAccessToken(tokenString string) (*domainservice.AccessTokenClaims, error) {
	var payload accessPayload
	if err := s.verifyToken(tokenString, &payload); err != nil {
		return nil, err
	}

	if payload.TokenType != tokenTypeAccess {
		return nil, domainerrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return nil, domainerrors.ErrInvalidToken
	}

	tenantID, err := uuid.Parse(payload.TenantID)
	if err != nil {
		return nil, domainerrors.ErrInvalidToken
	}

	return &domainservice.AccessTokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    payload.Email,
		Roles:    payload.Roles,
	}, nil
}

// ValidateRefreshToken parses and validates a signed refresh JWT string.
func (s *jwtService) ValidateRefreshToken(tokenString string) (*domainservice.RefreshTokenClaims, error) {
	var payload refreshPayload
	if err := s.verifyToken(tokenString, &payload); err != nil {
		return nil, err
	}

	if payload.TokenType != tokenTypeRefresh {
		return nil, domainerrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return nil, domainerrors.ErrInvalidToken
	}

	tenantID, err := uuid.Parse(payload.TenantID)
	if err != nil {
		return nil, domainerrors.ErrInvalidToken
	}

	tokenID, err := uuid.Parse(payload.TokenID)
	if err != nil {
		return nil, domainerrors.ErrInvalidToken
	}

	family, err := uuid.Parse(payload.Family)
	if err != nil {
		return nil, domainerrors.ErrInvalidToken
	}

	return &domainservice.RefreshTokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		TokenID:  tokenID,
		Family:   family,
		Roles:    payload.Roles,
	}, nil
}

func (s *jwtService) signToken(payload interface{}) (string, error) {
	header := jwtHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := headerB64 + "." + payloadB64

	sig := s.computeHMAC(signingInput)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return signingInput + "." + sigB64, nil
}

func (s *jwtService) verifyToken(tokenString string, targetPayload interface{}) error {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return domainerrors.ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := s.computeHMAC(signingInput)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return domainerrors.ErrInvalidToken
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return domainerrors.ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return domainerrors.ErrInvalidToken
	}

	if err := json.Unmarshal(payloadBytes, targetPayload); err != nil {
		return domainerrors.ErrInvalidToken
	}

	var expCheck struct {
		ExpiresAt int64 `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &expCheck); err == nil {
		if expCheck.ExpiresAt > 0 && time.Now().Unix() > expCheck.ExpiresAt {
			return domainerrors.ErrTokenExpired
		}
	}

	return nil
}

func (s *jwtService) computeHMAC(message string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(message))
	return mac.Sum(nil)
}
