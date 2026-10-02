package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/delivery/http/handlers"
	"github.com/skryfon/employee360/backend/internal/delivery/http/response"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/infrastructure/container"
	infraservice "github.com/skryfon/employee360/backend/internal/infrastructure/service"
	authtypes "github.com/skryfon/employee360/backend/internal/types/auth"
	dashtypes "github.com/skryfon/employee360/backend/internal/types/dashboard"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
	usecaseinterface "github.com/skryfon/employee360/backend/internal/usecase/interface"
	deptuc "github.com/skryfon/employee360/backend/internal/usecase/interface/department"
	"github.com/skryfon/employee360/backend/shared"
)

type fakeHealthUseCase struct{}

func (f *fakeHealthUseCase) Execute(ctx context.Context) usecaseinterface.HealthResult {
	return usecaseinterface.HealthResult{App: shared.AppName, Database: "ok"}
}

type fakeLoginUseCase struct{}

func (f *fakeLoginUseCase) Execute(ctx context.Context, req authtypes.LoginRequest, ipAddress, userAgent string) (*authtypes.LoginResponse, error) {
	return &authtypes.LoginResponse{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}, nil
}

type fakeTokenRefreshUseCase struct{}

func (f *fakeTokenRefreshUseCase) Execute(ctx context.Context, req authtypes.TokenRefreshRequest, ipAddress, userAgent string) (*authtypes.TokenRefreshResponse, error) {
	return &authtypes.TokenRefreshResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}, nil
}

type fakeLogoutUseCase struct{}

func (f *fakeLogoutUseCase) Execute(ctx context.Context, tenantID, userID uuid.UUID, req authtypes.LogoutRequest) error {
	return nil
}

type fakeForgotPasswordUseCase struct{}

func (f *fakeForgotPasswordUseCase) Execute(ctx context.Context, req authtypes.ForgotPasswordRequest) error {
	return nil
}

type fakeResetPasswordUseCase struct{}

func (f *fakeResetPasswordUseCase) Execute(ctx context.Context, req authtypes.ResetPasswordRequest) error {
	return nil
}

// fakeVerifier is an in-memory VerifyIdentityUseCase keyed by user id.
type fakeVerifier struct {
	users map[uuid.UUID]uuid.UUID // userID -> tenantID
	roles map[uuid.UUID][]string
	err   error
}

func (f *fakeVerifier) Execute(_ context.Context, tenantID, userID uuid.UUID) (*authtypes.VerifiedIdentity, error) {
	if f.err != nil {
		return nil, f.err
	}
	if tid, ok := f.users[userID]; !ok || tid != tenantID {
		return nil, domainerrors.ErrUnauthorized
	}
	return &authtypes.VerifiedIdentity{TenantID: tenantID, UserID: userID, Roles: f.roles[userID]}, nil
}

// registeringTokenService wraps a TokenService so each minted access token also
// registers a healthy identity (and its roles) in the fake verifier.
type registeringTokenService struct {
	domainservice.TokenService
	verifier *fakeVerifier
}

func (r *registeringTokenService) GenerateAccessToken(claims domainservice.AccessTokenClaims) (string, time.Time, error) {
	r.verifier.users[claims.UserID] = claims.TenantID
	r.verifier.roles[claims.UserID] = claims.Roles
	return r.TokenService.GenerateAccessToken(claims)
}

func verifierOf(svc domainservice.TokenService) *fakeVerifier {
	return svc.(*registeringTokenService).verifier
}

func setupTestTokenService(t *testing.T) domainservice.TokenService {
	t.Helper()
	svc, err := infraservice.NewJWTService("super-secret-jwt-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}
	return &registeringTokenService{TokenService: svc, verifier: &fakeVerifier{users: map[uuid.UUID]uuid.UUID{}, roles: map[uuid.UUID][]string{}}}
}

// testContainer builds a minimal *container.AppContainer directly from fakes,
// without touching the database, so SetupRouter can be exercised in isolation.
func testContainer(t *testing.T, cfg *config.Config, authHandler *handlers.AuthHandler, tokenService domainservice.TokenService) *container.AppContainer {
	t.Helper()

	ctr := &container.AppContainer{
		Config: cfg,
		Log:    zerolog.Nop(),
		Health: &container.HealthContainer{
			Handler: handlers.NewHealthHandler(&fakeHealthUseCase{}),
		},
	}
	if authHandler != nil {
		ctr.Auth = &container.AuthContainer{
			Handler:          authHandler,
			TokenService:     tokenService,
			IdentityVerifier: verifierOf(tokenService),
		}
	}
	return ctr
}

// TestSetupRouter_HealthRoute verifies middleware and the versioned health route.
func TestSetupRouter_HealthRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"https://app.employee360.example"}}}
	ctr := testContainer(t, cfg, nil, nil)
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "https://app.employee360.example")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/v1/health, got %d, body: %s", rec.Code, rec.Body.String())
	}

	if rec.Header().Get(shared.RequestIDHeader) == "" {
		t.Error("expected X-Request-ID header on the response")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.employee360.example" {
		t.Errorf("expected CORS Allow-Origin to reflect the allowed origin, got %q", got)
	}

	if !strings.Contains(rec.Body.String(), `"database"`) {
		t.Errorf("expected body to contain a database field, got: %s", rec.Body.String())
	}
}

// TestSetupRouter_HealthRoute_UnversionedEndpoints verifies /health and /healthz endpoints.
func TestSetupRouter_HealthRoute_UnversionedEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)

	for _, path := range []string{"/health", "/healthz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, got %d", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"database"`) {
			t.Errorf("expected body for %s to contain a database field, got: %s", path, rec.Body.String())
		}
	}
}

// TestSetupRouter_AuthRoutes verifies route registration and authentication requirements for all 5 auth endpoints.
func TestSetupRouter_AuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)

	authHandler := handlers.NewAuthHandler(
		&fakeLoginUseCase{},
		&fakeTokenRefreshUseCase{},
		&fakeLogoutUseCase{},
		&fakeForgotPasswordUseCase{},
		&fakeResetPasswordUseCase{},
	)

	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, authHandler, jwtSvc)
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)

	t.Run("POST /api/v1/auth/login is reachable unauthenticated", func(t *testing.T) {
		body, _ := json.Marshal(authtypes.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp response.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success=true")
		}
	})

	t.Run("POST /api/v1/auth/refresh is reachable unauthenticated", func(t *testing.T) {
		body, _ := json.Marshal(authtypes.TokenRefreshRequest{
			RefreshToken: "refresh-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/auth/forgot-password is reachable unauthenticated", func(t *testing.T) {
		body, _ := json.Marshal(authtypes.ForgotPasswordRequest{
			Email: "test@example.com",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/auth/reset-password is reachable unauthenticated", func(t *testing.T) {
		body, _ := json.Marshal(authtypes.ResetPasswordRequest{
			Token:       "token",
			NewPassword: "newpassword123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/auth/logout rejects unauthenticated requests with 401", func(t *testing.T) {
		body, _ := json.Marshal(authtypes.LogoutRequest{
			RefreshToken: "refresh-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized for unauthenticated logout, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/auth/logout succeeds with valid access token", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		token, _, err := jwtSvc.GenerateAccessToken(domainservice.AccessTokenClaims{
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []string{entity.RoleEmployee},
			Email:    "emp@example.com",
		})
		if err != nil {
			t.Fatalf("failed to generate access token: %v", err)
		}

		body, _ := json.Marshal(authtypes.LogoutRequest{
			RefreshToken: "refresh-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for authenticated logout, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

type fakeInvite struct{ gotTenant string }

func (f *fakeInvite) Execute(c context.Context, tenantID, _ uuid.UUID, req invtypes.InviteUserRequest) (*entity.UserInvitation, error) {
	f.gotTenant = tenantID.String()
	return &entity.UserInvitation{ID: uuid.New(), Email: req.Email, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type fakeAccept struct{ err error }

func (f *fakeAccept) Execute(context.Context, invtypes.AcceptInvitationRequest) error { return f.err }

type fakeResend struct{}

func (fakeResend) Execute(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (*entity.UserInvitation, error) {
	return &entity.UserInvitation{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type fakeRevoke struct{ err error }

func (f fakeRevoke) Execute(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error { return f.err }

type fakeList struct{}

func (fakeList) Execute(context.Context, uuid.UUID, invtypes.ListInvitationsQuery) (*invtypes.ListInvitationsResult, error) {
	items := []*entity.InvitationListItem{{UserInvitation: entity.UserInvitation{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}}}
	return &invtypes.ListInvitationsResult{Items: items, Total: 1, Page: 1, PageSize: 20, TotalPages: 1}, nil
}

type fakeValidate struct{}

func (fakeValidate) Execute(context.Context, string) (*invtypes.ValidateInvitationResponse, error) {
	return &invtypes.ValidateInvitationResponse{Email: "x@y.com", Role: "employee"}, nil
}

func invitationEngine(t *testing.T, invite *fakeInvite, accept *fakeAccept, revoke fakeRevoke) (*gin.Engine, domainservice.TokenService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)
	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{
		TokenService:      jwtSvc,
		IdentityVerifier:  verifierOf(jwtSvc),
		InvitationHandler: handlers.NewInvitationHandler(invite, accept, fakeResend{}, revoke, fakeList{}, fakeValidate{}),
	}
	return SetupRouter(cfg, zerolog.Nop(), ctr), jwtSvc
}

func bearer(t *testing.T, svc domainservice.TokenService, tenant uuid.UUID, roles ...string) string {
	t.Helper()
	tok, _, err := svc.GenerateAccessToken(domainservice.AccessTokenClaims{UserID: uuid.New(), TenantID: tenant, Roles: roles, Email: "a@b.com"})
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + tok
}

func do(engine *gin.Engine, method, path, auth, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestInvitationRoutes_ManagementRequiresAdminToken(t *testing.T) {
	engine, jwtSvc := invitationEngine(t, &fakeInvite{}, &fakeAccept{}, fakeRevoke{})
	id := uuid.New().String()
	routes := []struct{ method, path, body string }{
		{http.MethodPost, "/api/v1/users/invitations", `{"email":"x@y.com","role_id":"` + id + `"}`},
		{http.MethodGet, "/api/v1/users/invitations", ""},
		{http.MethodPost, "/api/v1/users/invitations/" + id + "/resend", ""},
		{http.MethodDelete, "/api/v1/users/invitations/" + id, ""},
	}
	tenant := uuid.New()
	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			if rec := do(engine, r.method, r.path, "", r.body); rec.Code != http.StatusUnauthorized {
				t.Errorf("no token: want 401, got %d", rec.Code)
			}
			if rec := do(engine, r.method, r.path, "Bearer garbage", r.body); rec.Code != http.StatusUnauthorized {
				t.Errorf("bad token: want 401, got %d", rec.Code)
			}
			if rec := do(engine, r.method, r.path, bearer(t, jwtSvc, tenant, "employee"), r.body); rec.Code != http.StatusForbidden {
				t.Errorf("employee: want 403, got %d", rec.Code)
			}
			rec := do(engine, r.method, r.path, bearer(t, jwtSvc, tenant, "admin"), r.body)
			if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
				t.Errorf("admin: want 2xx, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestInvitationRoutes_TenantComesFromTokenNotClient(t *testing.T) {
	invite := &fakeInvite{}
	engine, jwtSvc := invitationEngine(t, invite, &fakeAccept{}, fakeRevoke{})
	tenant := uuid.New()
	other := uuid.New()
	body := `{"email":"x@y.com","role_id":"` + uuid.New().String() + `","tenant_id":"` + other.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/invitations", bytes.NewBufferString(body))
	req.Header.Set("Authorization", bearer(t, jwtSvc, tenant, "admin"))
	req.Header.Set("X-Tenant-ID", other.String())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if invite.gotTenant != tenant.String() {
		t.Errorf("usecase saw tenant %q, want token tenant %q", invite.gotTenant, tenant)
	}
}

func TestInvitationRoutes_AcceptIsUnauthenticated(t *testing.T) {
	engine, _ := invitationEngine(t, &fakeInvite{}, &fakeAccept{}, fakeRevoke{})
	body := `{"token":"abc","password":"password123"}`
	if rec := do(engine, http.MethodPost, "/api/v1/invitations/accept", "", body); rec.Code != http.StatusOK {
		t.Fatalf("want 200 without token, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvitationRoutes_ValidateIsUnauthenticated(t *testing.T) {
	engine, _ := invitationEngine(t, &fakeInvite{}, &fakeAccept{}, fakeRevoke{})
	rec := do(engine, http.MethodGet, "/api/v1/invitations/validate?token=abc", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 without token, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvitationRoutes_AllRegistered(t *testing.T) {
	engine, _ := invitationEngine(t, &fakeInvite{}, &fakeAccept{}, fakeRevoke{})
	registered := map[string]bool{}
	for _, r := range engine.Routes() {
		registered[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{
		"POST /api/v1/invitations/accept",
		"GET /api/v1/invitations/validate",
		"POST /api/v1/users/invitations",
		"GET /api/v1/users/invitations",
		"POST /api/v1/users/invitations/:id/resend",
		"DELETE /api/v1/users/invitations/:id",
	} {
		if !registered[want] {
			t.Errorf("route not registered: %s", want)
		}
	}
}

func TestInvitationRoutes_AcceptRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		CORS:      config.CORSConfig{AllowedOrigins: []string{"*"}},
		RateLimit: config.RateLimitConfig{Enabled: true, RequestsPerSecond: 0.001, Burst: 2},
	}
	rlSvc := setupTestTokenService(t)
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{
		TokenService:      rlSvc,
		IdentityVerifier:  verifierOf(rlSvc),
		InvitationHandler: handlers.NewInvitationHandler(&fakeInvite{}, &fakeAccept{}, fakeResend{}, fakeRevoke{}, fakeList{}, fakeValidate{}),
	}
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)

	body := `{"token":"t","password":"Sup3rSecret!pw"}`
	for i := 0; i < 2; i++ {
		if rec := do(engine, http.MethodPost, "/api/v1/invitations/accept", "", body); rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d: %s", i, rec.Code, rec.Body.String())
		}
	}
	rec := do(engine, http.MethodPost, "/api/v1/invitations/accept", "", body)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
	// Health stays exempt.
	for i := 0; i < 5; i++ {
		if rec := do(engine, http.MethodGet, "/healthz", "", ""); rec.Code == http.StatusTooManyRequests {
			t.Fatal("health route must be exempt from rate limiting")
		}
	}
}

type fakeListRoles struct{}

func (fakeListRoles) Execute(context.Context, uuid.UUID) ([]*entity.Role, error) {
	return []*entity.Role{{ID: uuid.New(), Name: "employee"}}, nil
}

func TestRoleRoutes_ListRequiresAdminToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)
	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{TokenService: jwtSvc, IdentityVerifier: verifierOf(jwtSvc), RoleHandler: handlers.NewRoleHandler(fakeListRoles{})}
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)

	const path = "/api/v1/roles"
	tenant := uuid.New()
	tests := []struct {
		name, auth string
		want       int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"employee", bearer(t, jwtSvc, tenant, "employee"), http.StatusForbidden},
		{"admin", bearer(t, jwtSvc, tenant, "admin"), http.StatusOK},
		{"super_admin", bearer(t, jwtSvc, tenant, "super_admin"), http.StatusOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := do(engine, http.MethodGet, path, tc.auth, ""); rec.Code != tc.want {
				t.Fatalf("want %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestInvitationRoutes_ValidateRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		CORS:      config.CORSConfig{AllowedOrigins: []string{"*"}},
		RateLimit: config.RateLimitConfig{Enabled: true, RequestsPerSecond: 0.001, Burst: 2},
	}
	rlSvc := setupTestTokenService(t)
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{
		TokenService:      rlSvc,
		IdentityVerifier:  verifierOf(rlSvc),
		InvitationHandler: handlers.NewInvitationHandler(&fakeInvite{}, &fakeAccept{}, fakeResend{}, fakeRevoke{}, fakeList{}, fakeValidate{}),
	}
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)

	for i := 0; i < 2; i++ {
		if rec := do(engine, http.MethodGet, "/api/v1/invitations/validate?token=t", "", ""); rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d: %s", i, rec.Code, rec.Body.String())
		}
	}
	rec := do(engine, http.MethodGet, "/api/v1/invitations/validate?token=t", "", "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

// A validly signed token whose user/tenant is no longer usable must be rejected
// by the real route wiring (Auth middleware + verifier), and a verifier failure
// must fail closed.
func TestProtectedRoutes_RejectUnverifiableIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)
	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{TokenService: jwtSvc, IdentityVerifier: verifierOf(jwtSvc), RoleHandler: handlers.NewRoleHandler(fakeListRoles{})}
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)

	tenant := uuid.New()
	auth := bearer(t, jwtSvc, tenant, entity.RoleAdmin)
	if rec := do(engine, http.MethodGet, "/api/v1/roles", auth, ""); rec.Code != http.StatusOK {
		t.Fatalf("healthy identity: got %d: %s", rec.Code, rec.Body.String())
	}

	v := verifierOf(jwtSvc)
	for id := range v.users { // user deleted/deactivated or tenant removed
		delete(v.users, id)
	}
	if rec := do(engine, http.MethodGet, "/api/v1/roles", auth, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked identity: got %d, want 401", rec.Code)
	}

	v.err = errors.New("db down")
	if rec := do(engine, http.MethodGet, "/api/v1/roles", auth, ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("verifier error: got %d, want 503", rec.Code)
	}
}

// Pre-auth routes must not invoke the verifier at all.
func TestPreAuthRoutes_DoNotVerifyIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)
	verifierOf(jwtSvc).err = errors.New("verifier must not be called")
	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{
		TokenService:      jwtSvc,
		IdentityVerifier:  verifierOf(jwtSvc),
		InvitationHandler: handlers.NewInvitationHandler(&fakeInvite{}, &fakeAccept{}, fakeResend{}, fakeRevoke{}, fakeList{}, fakeValidate{}),
	}
	engine := SetupRouter(cfg, zerolog.Nop(), ctr)
	if rec := do(engine, http.MethodGet, "/api/v1/invitations/validate?token=t", "", ""); rec.Code != http.StatusOK {
		t.Fatalf("validate: got %d: %s", rec.Code, rec.Body.String())
	}
}

type fakeCreateDept struct {
	gotTenant string
	gotActor  string
}

func (f *fakeCreateDept) Execute(c context.Context, tenantID, actorID uuid.UUID, in deptuc.CreateDepartmentInput) (*entity.Department, error) {
	f.gotTenant, f.gotActor = tenantID.String(), actorID.String()
	return &entity.Department{ID: uuid.New(), TenantID: tenantID, Name: in.Name, Description: in.Description}, nil
}

type fakeGetDept struct {
	ownerTenant uuid.UUID
	deptID      uuid.UUID
}

func (f *fakeGetDept) Execute(c context.Context, tenantID uuid.UUID, in deptuc.GetDepartmentInput) (*entity.Department, error) {
	if in.ID == f.deptID && tenantID != f.ownerTenant {
		return nil, domainerrors.ErrDepartmentNotFound
	}
	return &entity.Department{ID: in.ID, TenantID: tenantID, Name: "Engineering"}, nil
}

type fakeListDept struct{}

func (fakeListDept) Execute(c context.Context, tenantID uuid.UUID, in deptuc.ListDepartmentsInput) (*deptuc.ListDepartmentsOutput, error) {
	return &deptuc.ListDepartmentsOutput{
		Departments: []*entity.Department{{ID: uuid.New(), TenantID: tenantID, Name: "HR"}},
		Total:       1,
		Page:        1,
		PageSize:    20,
	}, nil
}

type fakeUpdateDept struct {
	ownerTenant uuid.UUID
	deptID      uuid.UUID
}

func (f *fakeUpdateDept) Execute(c context.Context, tenantID, actorID uuid.UUID, in deptuc.UpdateDepartmentInput) (*entity.Department, error) {
	if in.ID == f.deptID && tenantID != f.ownerTenant {
		return nil, domainerrors.ErrDepartmentNotFound
	}
	return &entity.Department{ID: in.ID, TenantID: tenantID, Name: in.Name, Description: in.Description}, nil
}

type fakeDeleteDept struct {
	ownerTenant uuid.UUID
	deptID      uuid.UUID
}

func (f *fakeDeleteDept) Execute(c context.Context, tenantID, actorID uuid.UUID, in deptuc.DeleteDepartmentInput) error {
	if in.ID == f.deptID && tenantID != f.ownerTenant {
		return domainerrors.ErrDepartmentNotFound
	}
	return nil
}

func departmentEngine(
	t *testing.T,
	create deptuc.CreateDepartmentUseCase,
	get deptuc.GetDepartmentUseCase,
	list deptuc.ListDepartmentsUseCase,
	update deptuc.UpdateDepartmentUseCase,
	del deptuc.DeleteDepartmentUseCase,
) (*gin.Engine, domainservice.TokenService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)
	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{
		TokenService:     jwtSvc,
		IdentityVerifier: verifierOf(jwtSvc),
	}
	ctr.Department = &container.DepartmentContainer{
		Handler: handlers.NewDepartmentHandler(create, get, list, update, del),
	}
	return SetupRouter(cfg, zerolog.Nop(), ctr), jwtSvc
}

type fakeAdminDash struct{ gotTenant uuid.UUID }

func (f *fakeAdminDash) Execute(_ context.Context, tenantID uuid.UUID) (*dashtypes.AdminDashboardResult, error) {
	f.gotTenant = tenantID
	return &dashtypes.AdminDashboardResult{
		Counts:            &entity.TenantDashboardCounts{UsersTotal: 4, UsersActive: 3, UsersPendingInvited: 1, Departments: 2, Positions: 5, Invitations: entity.InvitationStatusCounts{Pending: 1}},
		RecentInvitations: []*entity.InvitationListItem{{UserInvitation: entity.UserInvitation{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}}},
	}, nil
}

type fakeSuperDash struct{ gotTenant uuid.UUID }

func (f *fakeSuperDash) Execute(_ context.Context, tenantID uuid.UUID) (*dashtypes.SuperAdminDashboardResult, error) {
	f.gotTenant = tenantID
	return &dashtypes.SuperAdminDashboardResult{
		AdminDashboardResult: dashtypes.AdminDashboardResult{Counts: &entity.TenantDashboardCounts{UsersTotal: 9, UsersActive: 7}},
		Tenant:               &entity.Tenant{ID: tenantID, Name: "Acme", IsActive: true},
		UsersByRole:          []entity.RoleUserCount{{Role: "admin", Total: 1, Active: 1}},
	}, nil
}

func dashboardEngine(t *testing.T, admin *fakeAdminDash, super *fakeSuperDash) (*gin.Engine, domainservice.TokenService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)
	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{
		TokenService:     jwtSvc,
		IdentityVerifier: verifierOf(jwtSvc),
		DashboardHandler: handlers.NewDashboardHandler(admin, super),
	}
	return SetupRouter(cfg, zerolog.Nop(), ctr), jwtSvc
}

func TestDepartmentRoutes_AuthAndRoles(t *testing.T) {
	deptID := uuid.New()
	tenant := uuid.New()

	engine, jwtSvc := departmentEngine(
		t,
		&fakeCreateDept{},
		&fakeGetDept{ownerTenant: tenant, deptID: deptID},
		fakeListDept{},
		&fakeUpdateDept{ownerTenant: tenant, deptID: deptID},
		&fakeDeleteDept{ownerTenant: tenant, deptID: deptID},
	)

	routes := []struct {
		method, path, body string
		wantSuccessCode    int
	}{
		{http.MethodPost, "/api/v1/departments", `{"name":"Engineering","description":"Platform"}`, http.StatusCreated},
		{http.MethodGet, "/api/v1/departments", "", http.StatusOK},
		{http.MethodGet, "/api/v1/departments/" + deptID.String(), "", http.StatusOK},
		{http.MethodPut, "/api/v1/departments/" + deptID.String(), `{"name":"Engineering 2","description":"Platform"}`, http.StatusOK},
		{http.MethodDelete, "/api/v1/departments/" + deptID.String(), "", http.StatusOK},
	}

	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			// Unauthenticated -> 401
			if rec := do(engine, r.method, r.path, "", r.body); rec.Code != http.StatusUnauthorized {
				t.Errorf("no token: want 401, got %d", rec.Code)
			}
			// Bad token -> 401
			if rec := do(engine, r.method, r.path, "Bearer invalid-token", r.body); rec.Code != http.StatusUnauthorized {
				t.Errorf("bad token: want 401, got %d", rec.Code)
			}
			// Employee -> 403
			if rec := do(engine, r.method, r.path, bearer(t, jwtSvc, tenant, "employee"), r.body); rec.Code != http.StatusForbidden {
				t.Errorf("employee: want 403, got %d", rec.Code)
			}
			// Admin -> 200/201
			rec := do(engine, r.method, r.path, bearer(t, jwtSvc, tenant, "admin"), r.body)
			if rec.Code != r.wantSuccessCode {
				t.Errorf("admin: want %d, got %d: %s", r.wantSuccessCode, rec.Code, rec.Body.String())
			}
			// Super Admin -> 200/201 (scoped to tenant)
			rec = do(engine, r.method, r.path, bearer(t, jwtSvc, tenant, "super_admin"), r.body)
			if rec.Code != r.wantSuccessCode {
				t.Errorf("super_admin: want %d, got %d: %s", r.wantSuccessCode, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDepartmentRoutes_CrossTenant404(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	deptID := uuid.New()

	engine, jwtSvc := departmentEngine(
		t,
		&fakeCreateDept{},
		&fakeGetDept{ownerTenant: tenantA, deptID: deptID},
		fakeListDept{},
		&fakeUpdateDept{ownerTenant: tenantA, deptID: deptID},
		&fakeDeleteDept{ownerTenant: tenantA, deptID: deptID},
	)

	routes := []struct {
		method, path, body string
	}{
		{http.MethodGet, "/api/v1/departments/" + deptID.String(), ""},
		{http.MethodPut, "/api/v1/departments/" + deptID.String(), `{"name":"Hacked","description":"Cross-tenant"}`},
		{http.MethodDelete, "/api/v1/departments/" + deptID.String(), ""},
	}

	for _, r := range routes {
		t.Run("tenantB accessing tenantA department "+r.method, func(t *testing.T) {
			rec := do(engine, r.method, r.path, bearer(t, jwtSvc, tenantB, "admin"), r.body)
			if rec.Code != http.StatusNotFound {
				t.Errorf("cross-tenant admin: want 404, got %d: %s", rec.Code, rec.Body.String())
			}

			// Super Admin in tenantB also receives 404 (no cross-tenant bypass)
			rec = do(engine, r.method, r.path, bearer(t, jwtSvc, tenantB, "super_admin"), r.body)
			if rec.Code != http.StatusNotFound {
				t.Errorf("cross-tenant super_admin: want 404, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDepartmentRoutes_TenantFromTokenNotClient(t *testing.T) {
	create := &fakeCreateDept{}
	engine, jwtSvc := departmentEngine(
		t,
		create,
		&fakeGetDept{},
		fakeListDept{},
		&fakeUpdateDept{},
		&fakeDeleteDept{},
	)

	tenantA := uuid.New()
	spoofedTenant := uuid.New()

	body := `{"name":"Engineering","description":"Test","tenant_id":"` + spoofedTenant.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/departments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer(t, jwtSvc, tenantA, "admin"))
	req.Header.Set("X-Tenant-ID", spoofedTenant.String())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if create.gotTenant != tenantA.String() {
		t.Errorf("usecase saw tenant %q, want token tenant %q", create.gotTenant, tenantA)
	}
}

func TestDepartmentRoutes_AllRegistered(t *testing.T) {
	engine, _ := departmentEngine(t, &fakeCreateDept{}, &fakeGetDept{}, fakeListDept{}, &fakeUpdateDept{}, &fakeDeleteDept{})
	registered := map[string]bool{}
	for _, r := range engine.Routes() {
		registered[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{
		"POST /api/v1/departments",
		"GET /api/v1/departments",
		"GET /api/v1/departments/:id",
		"PUT /api/v1/departments/:id",
		"DELETE /api/v1/departments/:id",
	} {
		if !registered[want] {
			t.Errorf("route not registered: %s", want)
		}
	}
}

func TestDashboardRoutes_RoleGuards(t *testing.T) {
	admin, super := &fakeAdminDash{}, &fakeSuperDash{}
	engine, jwtSvc := dashboardEngine(t, admin, super)
	tenant := uuid.New()
	const adminPath, superPath = "/api/v1/dashboard/admin", "/api/v1/dashboard/super-admin"

	for _, p := range []string{adminPath, superPath} {
		if rec := do(engine, http.MethodGet, p, "", ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s no token: want 401, got %d", p, rec.Code)
		}
	}
	// admin: 200 on admin route, 403 on super-admin route.
	if rec := do(engine, http.MethodGet, superPath, bearer(t, jwtSvc, tenant, entity.RoleAdmin), ""); rec.Code != http.StatusForbidden {
		t.Errorf("admin on super-admin route: want 403, got %d", rec.Code)
	}
	if rec := do(engine, http.MethodGet, adminPath, bearer(t, jwtSvc, tenant, entity.RoleAdmin), ""); rec.Code != http.StatusOK {
		t.Fatalf("admin on admin route: want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if admin.gotTenant != tenant {
		t.Errorf("usecase got tenant %s, want auth-context tenant %s", admin.gotTenant, tenant)
	}
	// employee: 403 on both.
	for _, p := range []string{adminPath, superPath} {
		if rec := do(engine, http.MethodGet, p, bearer(t, jwtSvc, tenant, entity.RoleEmployee), ""); rec.Code != http.StatusForbidden {
			t.Errorf("employee on %s: want 403, got %d", p, rec.Code)
		}
	}
	// super_admin: 200 on super-admin route, 403 on tenant admin route.
	rec := do(engine, http.MethodGet, superPath, bearer(t, jwtSvc, tenant, entity.RoleSuperAdmin), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("super_admin on super-admin route: want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var env struct {
		Data dashtypes.SuperAdminDashboardResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Tenant.Name != "Acme" || env.Data.Users.Total != 9 || env.Data.Users.Inactive != 2 || len(env.Data.UsersByRole) != 1 || super.gotTenant != tenant {
		t.Errorf("unexpected payload: %s", rec.Body.String())
	}
	if rec := do(engine, http.MethodGet, adminPath, bearer(t, jwtSvc, tenant, entity.RoleSuperAdmin), ""); rec.Code != http.StatusForbidden {
		t.Errorf("super_admin on admin route: want 403, got %d", rec.Code)
	}
}
