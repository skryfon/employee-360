package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/delivery/http/handlers"
	"github.com/skryfon/employee360/backend/internal/delivery/http/response"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/infrastructure/container"
	infraservice "github.com/skryfon/employee360/backend/internal/infrastructure/service"
	authtypes "github.com/skryfon/employee360/backend/internal/types/auth"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
	usecaseinterface "github.com/skryfon/employee360/backend/internal/usecase/interface"
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

func setupTestTokenService(t *testing.T) domainservice.TokenService {
	t.Helper()
	svc, err := infraservice.NewJWTService("super-secret-jwt-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}
	return svc
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
			Handler:      authHandler,
			TokenService: tokenService,
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

func (f *fakeInvite) Execute(c context.Context, req invtypes.InviteUserRequest) (*entity.UserInvitation, error) {
	f.gotTenant, _ = ctx.TenantIDFromContext(c)
	return &entity.UserInvitation{ID: uuid.New(), Email: req.Email, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type fakeAccept struct{ err error }

func (f *fakeAccept) Execute(context.Context, invtypes.AcceptInvitationRequest) error { return f.err }

type fakeResend struct{}

func (fakeResend) Execute(context.Context, uuid.UUID) (*entity.UserInvitation, error) {
	return &entity.UserInvitation{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type fakeRevoke struct{ err error }

func (f fakeRevoke) Execute(context.Context, uuid.UUID) error { return f.err }

type fakeList struct{}

func (fakeList) Execute(context.Context, int, int) ([]*entity.UserInvitation, int64, error) {
	return []*entity.UserInvitation{{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}}, 1, nil
}

func invitationEngine(t *testing.T, invite *fakeInvite, accept *fakeAccept, revoke fakeRevoke) (*gin.Engine, domainservice.TokenService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	jwtSvc := setupTestTokenService(t)
	cfg := &config.Config{CORS: config.CORSConfig{AllowedOrigins: []string{"*"}}}
	ctr := testContainer(t, cfg, nil, nil)
	ctr.Auth = &container.AuthContainer{
		TokenService:      jwtSvc,
		InvitationHandler: handlers.NewInvitationHandler(invite, accept, fakeResend{}, revoke, fakeList{}),
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

func TestInvitationRoutes_AllRegistered(t *testing.T) {
	engine, _ := invitationEngine(t, &fakeInvite{}, &fakeAccept{}, fakeRevoke{})
	registered := map[string]bool{}
	for _, r := range engine.Routes() {
		registered[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{
		"POST /api/v1/invitations/accept",
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
