package container

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/config"
)

// TestNew_ConstructsWithoutTouchingDB verifies the container can be built
// from a nil *gorm.DB: the database pinger wraps db lazily and is only
// dereferenced when Ping is actually called (e.g. by a request hitting the
// health route), not during construction.
func TestNew_ConstructsWithoutTouchingDB(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: []string{"*"}},
	}

	c, err := New(cfg, nil, zerolog.Nop())
	if err != nil {
		t.Fatalf("expected New() to succeed, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected a non-nil Container")
	}
	if c.Handlers.Health == nil {
		t.Fatal("expected Handlers.Health to be wired")
	}
}

// TestContainer_Router_RegistersHealthRoute verifies the /api/v1/health
// route is registered on the engine returned by Router(), without ever
// invoking the route (which would dereference the nil DB via the pinger).
func TestContainer_Router_RegistersHealthRoute(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: []string{"*"}},
	}

	c, err := New(cfg, nil, zerolog.Nop())
	if err != nil {
		t.Fatalf("expected New() to succeed, got: %v", err)
	}

	engine := c.Router()

	found := false
	for _, route := range engine.Routes() {
		if route.Method == "GET" && route.Path == "/api/v1/health" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected GET /api/v1/health to be registered, got routes: %+v", engine.Routes())
	}
}
