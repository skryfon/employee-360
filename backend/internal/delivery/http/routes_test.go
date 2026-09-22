package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/internal/delivery/http/handlers"
	usecaseinterface "github.com/skryfon/employee360/backend/internal/usecase/interface"
	"github.com/skryfon/employee360/backend/shared"
)

type fakeHealthUseCase struct{}

func (f *fakeHealthUseCase) Execute(ctx context.Context) usecaseinterface.HealthResult {
	return usecaseinterface.HealthResult{App: shared.AppName, Database: "ok"}
}

// TestNewRouter_HealthRoute verifies middleware and the versioned health route.
func TestNewRouter_HealthRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := NewRouter(zerolog.Nop(), []string{"https://app.employee360.example"}, Handlers{
		Health: handlers.NewHealthHandler(&fakeHealthUseCase{}),
	})

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

// TestNewRouter_HealthRoute_UnversionedEndpoints verifies /health and /healthz endpoints.
func TestNewRouter_HealthRoute_UnversionedEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := NewRouter(zerolog.Nop(), []string{"*"}, Handlers{
		Health: handlers.NewHealthHandler(&fakeHealthUseCase{}),
	})

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
