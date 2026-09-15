package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_Defaults(t *testing.T) {
	// Ensure clean env for testing defaults
	os.Unsetenv("DATABASE_HOST")
	os.Unsetenv("DATABASE_PORT")
	os.Unsetenv("DATABASE_USER")
	os.Unsetenv("DATABASE_PASSWORD")
	os.Unsetenv("DATABASE_NAME")
	os.Unsetenv("DATABASE_DBNAME")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("PORT")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("ENVIRONMENT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected Load() to succeed with defaults, got: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("expected Server.Port '8080', got %q", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected Server.Host '0.0.0.0', got %q", cfg.Server.Host)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host 'localhost', got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("expected Database.Port '5432', got %q", cfg.Database.Port)
	}
	if cfg.Database.User != "postgres" {
		t.Errorf("expected Database.User 'postgres', got %q", cfg.Database.User)
	}
	if cfg.Database.Password != "postgres" {
		t.Errorf("expected Database.Password 'postgres', got %q", cfg.Database.Password)
	}
	if cfg.Database.DBName != "employee360" {
		t.Errorf("expected Database.DBName 'employee360', got %q", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("expected Database.SSLMode 'disable', got %q", cfg.Database.SSLMode)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("expected Database.MaxOpenConns 25, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 5 {
		t.Errorf("expected Database.MaxIdleConns 5, got %d", cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnTimeout != 5*time.Second {
		t.Errorf("expected Database.ConnTimeout 5s, got %v", cfg.Database.ConnTimeout)
	}
}

func TestConfig_EnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_HOST", "db.example.com")
	t.Setenv("DATABASE_PORT", "5433")
	t.Setenv("DATABASE_USER", "custom_user")
	t.Setenv("DATABASE_PASSWORD", "custom_secret")
	t.Setenv("DATABASE_NAME", "custom_employee360")
	t.Setenv("DATABASE_SSLMODE", "require")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("APP_ENV", "staging")
	t.Setenv("JWT_SECRET", "custom-32-byte-secure-jwt-secret-key-360")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected Load() to succeed with env vars, got: %v", err)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected Database.Host 'db.example.com', got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != "5433" {
		t.Errorf("expected Database.Port '5433', got %q", cfg.Database.Port)
	}
	if cfg.Database.User != "custom_user" {
		t.Errorf("expected Database.User 'custom_user', got %q", cfg.Database.User)
	}
	if cfg.Database.Password != "custom_secret" {
		t.Errorf("expected Database.Password 'custom_secret', got %q", cfg.Database.Password)
	}
	if cfg.Database.DBName != "custom_employee360" {
		t.Errorf("expected Database.DBName 'custom_employee360', got %q", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("expected Database.SSLMode 'require', got %q", cfg.Database.SSLMode)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("expected Server.Port '9090', got %q", cfg.Server.Port)
	}
	if cfg.App.Environment != "staging" {
		t.Errorf("expected App.Environment 'staging', got %q", cfg.App.Environment)
	}
}

func TestConfig_JWTSecretValidationInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", DefaultJWTSecret)

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to fail when default JWT secret is used in production")
	}

	// Now set a strong custom secret
	t.Setenv("JWT_SECRET", "a-secure-production-jwt-secret-with-over-32-chars")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected Load() to succeed with strong secret in production: %v", err)
	}
	if cfg.JWT.Secret != "a-secure-production-jwt-secret-with-over-32-chars" {
		t.Errorf("unexpected secret: %q", cfg.JWT.Secret)
	}
}

func TestConfig_DSN_And_URL(t *testing.T) {
	dbCfg := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpassword",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	expectedDSN := "host=localhost port=5432 user=testuser password=testpassword dbname=testdb sslmode=disable"
	if dbCfg.DSN() != expectedDSN {
		t.Errorf("expected DSN %q, got %q", expectedDSN, dbCfg.DSN())
	}

	expectedURL := "postgres://testuser:testpassword@localhost:5432/testdb?sslmode=disable"
	if dbCfg.URL() != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, dbCfg.URL())
	}
}

func TestConfig_YAMLFileLoading(t *testing.T) {
	tempDir := t.TempDir()
	yamlContent := `
server:
  port: "3000"
  host: "127.0.0.1"

database:
  host: "pg.internal"
  port: "5432"
  user: "yamluser"
  password: "yamlpassword"
  dbname: "yamldb"
  sslmode: "disable"
`
	configFilePath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configFilePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp yaml file: %v", err)
	}

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("failed to load config from yaml directory: %v", err)
	}

	if cfg.Server.Port != "3000" {
		t.Errorf("expected Server.Port '3000', got %q", cfg.Server.Port)
	}
	if cfg.Database.User != "yamluser" {
		t.Errorf("expected Database.User 'yamluser', got %q", cfg.Database.User)
	}
	if cfg.Database.DBName != "yamldb" {
		t.Errorf("expected Database.DBName 'yamldb', got %q", cfg.Database.DBName)
	}
}
