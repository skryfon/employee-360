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

// MockUserRepository implements repository.UserRepository for testing.
type mockUserRepository struct {
	usersByID    map[uuid.UUID]*entity.User
	usersByEmail map[string]*entity.User
	updatedUsers []*entity.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByID:    make(map[uuid.UUID]*entity.User),
		usersByEmail: make(map[string]*entity.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.usersByID[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domainerrors.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, domainerrors.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepository) GetByEmailWithRoles(ctx context.Context, email string) (*entity.User, error) {
	return m.GetByEmail(ctx, email)
}

func (m *mockUserRepository) Update(ctx context.Context, user *entity.User) error {
	m.updatedUsers = append(m.updatedUsers, user)
	m.usersByID[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.usersByID, id)
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, int64, error) {
	var list []*entity.User
	for _, u := range m.usersByID {
		list = append(list, u)
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

	output, err := uc.Execute(context.Background(), authusecase.LoginInput{
		Email:     "alice@example.com",
		Password:  "Secret123!",
		IPAddress: "127.0.0.1",
		UserAgent: "TestBrowser/1.0",
	})

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
			_, err := uc.Execute(context.Background(), authusecase.LoginInput{
				Email:    tc.email,
				Password: tc.password,
			})
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

	_, err := uc.Execute(context.Background(), authusecase.LoginInput{
		Email:    "inactive@example.com",
		Password: "Secret123!",
	})

	if err != domainerrors.ErrUserInactive {
		t.Errorf("expected ErrUserInactive, got %v", err)
	}
}
