package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	authtypes "github.com/skryfon/employee360/backend/internal/types/auth"
)

// tenantEmailKey composite-keys a user by (tenantID, email) so the mock can
// actually distinguish users that share an email across different tenants.
type tenantEmailKey struct {
	tenantID uuid.UUID
	email    string
}

// MockUserRepository implements repository.UserRepository for testing.
type mockUserRepository struct {
	usersByID          map[uuid.UUID]*entity.User
	usersByTenantEmail map[tenantEmailKey]*entity.User
	updatedUsers       []*entity.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByID:          make(map[uuid.UUID]*entity.User),
		usersByTenantEmail: make(map[tenantEmailKey]*entity.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.usersByID[user.ID] = user
	m.usersByTenantEmail[tenantEmailKey{tenantID: user.TenantID, email: user.Email}] = user
	return nil
}

// GetByID enforces the tenantID param against the stored user's own tenant --
// a mismatch is treated exactly like "not found", mirroring the real
// GORM adapter's "id = ? AND tenant_id = ?" scoping, so a test that passes
// the wrong tenant here is actually caught rather than silently ignored.
func (m *mockUserRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.User, error) {
	u, ok := m.usersByID[id]
	if !ok || u.TenantID != tenantID {
		return nil, domainerrors.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
	u, ok := m.usersByTenantEmail[tenantEmailKey{tenantID: tenantID, email: email}]
	if !ok {
		return nil, domainerrors.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByIDWithRoles(ctx context.Context, tenantID, id uuid.UUID) (*entity.User, error) {
	return m.GetByID(ctx, tenantID, id)
}

func (m *mockUserRepository) GetByTenantAndEmailWithRoles(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
	return m.GetByTenantAndEmail(ctx, tenantID, email)
}

// Update enforces the tenantID param against the user's own TenantID -- a
// mismatch is rejected (not found) rather than silently applied, mirroring
// the real GORM adapter's "id = ? AND tenant_id = ?" scoping.
func (m *mockUserRepository) Update(ctx context.Context, tenantID uuid.UUID, user *entity.User) error {
	existing, ok := m.usersByID[user.ID]
	if !ok || existing.TenantID != tenantID || user.TenantID != tenantID {
		return domainerrors.ErrUserNotFound
	}
	m.updatedUsers = append(m.updatedUsers, user)
	m.usersByID[user.ID] = user
	m.usersByTenantEmail[tenantEmailKey{tenantID: user.TenantID, email: user.Email}] = user
	return nil
}

// Delete enforces the tenantID param against the stored user's own tenant --
// a mismatch leaves the row untouched and returns not found, mirroring the
// real GORM adapter's "id = ? AND tenant_id = ?" scoping.
func (m *mockUserRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	u, ok := m.usersByID[id]
	if !ok || u.TenantID != tenantID {
		return domainerrors.ErrUserNotFound
	}
	delete(m.usersByID, id)
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.User, int64, error) {
	var list []*entity.User
	for _, u := range m.usersByID {
		if u.TenantID == tenantID {
			list = append(list, u)
		}
	}
	return list, int64(len(list)), nil
}

// MockRefreshTokenRepository implements repository.RefreshTokenRepository.
type mockRefreshTokenRepository struct {
	tokensByHash    map[string]*entity.RefreshToken
	tokensByID      map[uuid.UUID]*entity.RefreshToken
	revokedTokenIDs []uuid.UUID
	revokedFamilies []uuid.UUID
	revokedUserIDs  []uuid.UUID
}

func newMockRefreshTokenRepository() *mockRefreshTokenRepository {
	return &mockRefreshTokenRepository{
		tokensByHash: make(map[string]*entity.RefreshToken),
		tokensByID:   make(map[uuid.UUID]*entity.RefreshToken),
	}
}

func (m *mockRefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	m.tokensByHash[token.TokenHash] = token
	m.tokensByID[token.ID] = token
	return nil
}

func (m *mockRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	t, ok := m.tokensByHash[tokenHash]
	if !ok {
		return nil, domainerrors.ErrNotFound
	}
	return t, nil
}

func (m *mockRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	m.revokedTokenIDs = append(m.revokedTokenIDs, id)
	if t, ok := m.tokensByID[id]; ok {
		now := time.Now().UTC()
		t.RevokedAt = &now
	}
	return nil
}

func (m *mockRefreshTokenRepository) RevokeFamily(ctx context.Context, family uuid.UUID) error {
	m.revokedFamilies = append(m.revokedFamilies, family)
	now := time.Now().UTC()
	for _, t := range m.tokensByID {
		if t.Family == family {
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	m.revokedUserIDs = append(m.revokedUserIDs, userID)
	now := time.Now().UTC()
	for _, t := range m.tokensByID {
		if t.UserID == userID {
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockRefreshTokenRepository) DeleteExpiredTokens(ctx context.Context, before time.Time) error {
	return nil
}

// MockHashService implements service.HashService.
type mockHashService struct {
	compareErr error
}

func (m *mockHashService) HashPassword(password string) (string, error) {
	return "hashed_" + password, nil
}

func (m *mockHashService) ComparePassword(hashedPassword, password string) error {
	if m.compareErr != nil {
		return m.compareErr
	}
	if hashedPassword != "hashed_"+password {
		return domainerrors.ErrInvalidCredentials
	}
	return nil
}

func (m *mockHashService) HashToken(plainToken string) string {
	return "sha256_" + plainToken
}

// MockTokenService implements service.TokenService.
type mockTokenService struct {
	validateAccessClaims  *domainservice.AccessTokenClaims
	validateRefreshClaims *domainservice.RefreshTokenClaims
	validateErr           error
}

func (m *mockTokenService) GenerateAccessToken(claims domainservice.AccessTokenClaims) (string, time.Time, error) {
	return "access_jwt_" + claims.Email, time.Now().Add(15 * time.Minute), nil
}

func (m *mockTokenService) GenerateRefreshToken(claims domainservice.RefreshTokenClaims) (string, time.Time, error) {
	return "refresh_jwt_" + claims.TokenID.String(), time.Now().Add(7 * 24 * time.Hour), nil
}

func (m *mockTokenService) GenerateTokenPair(accessClaims domainservice.AccessTokenClaims, refreshClaims domainservice.RefreshTokenClaims) (*domainservice.TokenPair, error) {
	acc, exp, _ := m.GenerateAccessToken(accessClaims)
	ref, _, _ := m.GenerateRefreshToken(refreshClaims)
	return &domainservice.TokenPair{
		AccessToken:  acc,
		RefreshToken: ref,
		ExpiresAt:    exp,
		TokenType:    "Bearer",
	}, nil
}

func (m *mockTokenService) ValidateAccessToken(tokenString string) (*domainservice.AccessTokenClaims, error) {
	if m.validateErr != nil {
		return nil, m.validateErr
	}
	return m.validateAccessClaims, nil
}

func (m *mockTokenService) ValidateRefreshToken(tokenString string) (*domainservice.RefreshTokenClaims, error) {
	if m.validateErr != nil {
		return nil, m.validateErr
	}
	return m.validateRefreshClaims, nil
}

func TestLoginUseCase_Success(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	pwHash := "hashed_Secret123!"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{}

	user := &entity.User{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "alice@example.com",
		PasswordHash: &pwHash,
		IsActive:     true,
		Roles: []entity.Role{
			{ID: uuid.New(), TenantID: tenantID, Name: entity.RoleAdmin},
		},
	}
	_ = userRepo.Create(context.Background(), user)

	uc := NewLoginUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	output, err := uc.Execute(context.Background(), tenantID, authtypes.LoginRequest{
		Email:    "alice@example.com",
		Password: "Secret123!",
	}, "127.0.0.1", "TestBrowser/1.0")

	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}

	if output.AccessToken == "" || output.RefreshToken == "" {
		t.Fatalf("expected non-empty access and refresh tokens")
	}

	if output.User.ID != userID {
		t.Errorf("expected user ID %s, got %s", userID, output.User.ID)
	}

	// Verify refresh token was hashed and stored
	hashedRefToken := hashSvc.HashToken(output.RefreshToken)
	storedRT, ok := refreshTokenRepo.tokensByHash[hashedRefToken]
	if !ok {
		t.Fatalf("expected refresh token to be hashed and persisted in repo")
	}
	if storedRT.UserID != userID {
		t.Errorf("expected stored token user ID %s, got %s", userID, storedRT.UserID)
	}
	if storedRT.TenantID != tenantID {
		t.Errorf("expected stored token tenant ID %s, got %s", tenantID, storedRT.TenantID)
	}

	// Verify last login was updated
	if len(userRepo.updatedUsers) == 0 || userRepo.updatedUsers[0].LastLoginAt == nil {
		t.Errorf("expected user LastLoginAt to be updated")
	}
}

func TestLoginUseCase_InvalidCredentials(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	pwHash := "hashed_Secret123!"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{}

	user := &entity.User{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "alice@example.com",
		PasswordHash: &pwHash,
		IsActive:     true,
	}
	_ = userRepo.Create(context.Background(), user)

	uc := NewLoginUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"wrong password", "alice@example.com", "WrongPassword!"},
		{"unknown user", "bob@example.com", "Secret123!"},
		{"empty email", "", "Secret123!"},
		{"empty password", "alice@example.com", ""},
		{"invalid email format", "not-an-email", "Secret123!"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), tenantID, authtypes.LoginRequest{
				Email:    tc.email,
				Password: tc.password,
			}, "", "")
			if err != domainerrors.ErrInvalidCredentials {
				t.Errorf("expected ErrInvalidCredentials, got %v", err)
			}
		})
	}
}

func TestLoginUseCase_InactiveUser(t *testing.T) {
	tenantID := uuid.New()
	pwHash := "hashed_Secret123!"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{}

	user := &entity.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        "inactive@example.com",
		PasswordHash: &pwHash,
		IsActive:     false,
	}
	_ = userRepo.Create(context.Background(), user)

	uc := NewLoginUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	_, err := uc.Execute(context.Background(), tenantID, authtypes.LoginRequest{
		Email:    "inactive@example.com",
		Password: "Secret123!",
	}, "", "")

	if err != domainerrors.ErrUserInactive {
		t.Errorf("expected ErrUserInactive, got %v", err)
	}
}

// TestLoginUseCase_InactiveUser_WrongPassword proves that an incorrect
// password for an inactive user's email still yields ErrInvalidCredentials
// (not ErrUserInactive) — the account's active/inactive state must never be
// distinguishable from "unknown user" or "wrong password" before the
// password has actually been verified, to avoid a user-enumeration leak.
func TestLoginUseCase_InactiveUser_WrongPassword(t *testing.T) {
	tenantID := uuid.New()
	pwHash := "hashed_Secret123!"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{}

	user := &entity.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        "inactive@example.com",
		PasswordHash: &pwHash,
		IsActive:     false,
	}
	_ = userRepo.Create(context.Background(), user)

	uc := NewLoginUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	_, err := uc.Execute(context.Background(), tenantID, authtypes.LoginRequest{
		Email:    "inactive@example.com",
		Password: "WrongPassword!",
	}, "", "")

	if err != domainerrors.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for a wrong password against an inactive account, got %v", err)
	}
}

// TestLoginUseCase_TenantIsolation proves that two tenants with a user
// sharing the same email address are never conflated: logging in with
// tenantA's ID and tenantA's password must succeed against tenantA's user
// only, and tenantA's correct password must not authenticate tenantB's row
// (and vice versa).
func TestLoginUseCase_TenantIsolation(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	userA := uuid.New()
	userB := uuid.New()
	const sharedEmail = "shared@example.com"
	pwHashA := "hashed_PasswordForA!"
	pwHashB := "hashed_PasswordForB!"

	userRepo := newMockUserRepository()
	refreshTokenRepo := newMockRefreshTokenRepository()
	hashSvc := &mockHashService{}
	tokenSvc := &mockTokenService{}

	_ = userRepo.Create(context.Background(), &entity.User{
		ID:           userA,
		TenantID:     tenantA,
		Email:        sharedEmail,
		PasswordHash: &pwHashA,
		IsActive:     true,
	})
	_ = userRepo.Create(context.Background(), &entity.User{
		ID:           userB,
		TenantID:     tenantB,
		Email:        sharedEmail,
		PasswordHash: &pwHashB,
		IsActive:     true,
	})

	uc := NewLoginUseCase(userRepo, tokenSvc, hashSvc, refreshTokenRepo)

	// tenantA's correct password against tenantA succeeds and resolves userA.
	output, err := uc.Execute(context.Background(), tenantA, authtypes.LoginRequest{
		Email:    sharedEmail,
		Password: "PasswordForA!",
	}, "", "")
	if err != nil {
		t.Fatalf("expected tenantA login to succeed, got error: %v", err)
	}
	if output.User.ID != userA {
		t.Errorf("expected tenantA login to resolve user %s, got %s", userA, output.User.ID)
	}

	// tenantA's correct password must NOT authenticate against tenantB's row.
	_, err = uc.Execute(context.Background(), tenantB, authtypes.LoginRequest{
		Email:    sharedEmail,
		Password: "PasswordForA!",
	}, "", "")
	if err != domainerrors.ErrInvalidCredentials {
		t.Errorf("expected tenantA's password to be rejected for tenantB, got %v", err)
	}

	// tenantB's correct password must NOT authenticate against tenantA's row.
	_, err = uc.Execute(context.Background(), tenantA, authtypes.LoginRequest{
		Email:    sharedEmail,
		Password: "PasswordForB!",
	}, "", "")
	if err != domainerrors.ErrInvalidCredentials {
		t.Errorf("expected tenantB's password to be rejected for tenantA, got %v", err)
	}

	// tenantB's correct password against tenantB succeeds and resolves userB.
	output, err = uc.Execute(context.Background(), tenantB, authtypes.LoginRequest{
		Email:    sharedEmail,
		Password: "PasswordForB!",
	}, "", "")
	if err != nil {
		t.Fatalf("expected tenantB login to succeed, got error: %v", err)
	}
	if output.User.ID != userB {
		t.Errorf("expected tenantB login to resolve user %s, got %s", userB, output.User.ID)
	}
}
