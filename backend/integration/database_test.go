//go:build integration

package integration

import (
	"os"
	"testing"
	"time"

	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

func getTestDatabaseConfig() config.DatabaseConfig {
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DATABASE_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DATABASE_USER")
	if user == "" {
		user = "postgres"
	}
	pass := os.Getenv("DATABASE_PASSWORD")
	if pass == "" {
		pass = "postgres"
	}
	dbname := os.Getenv("DATABASE_NAME")
	if dbname == "" {
		dbname = "employee360"
	}
	sslmode := os.Getenv("DATABASE_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return config.DatabaseConfig{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        pass,
		DBName:          dbname,
		SSLMode:         sslmode,
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnTimeout:     2 * time.Second,
	}
}

func TestConnect_SuccessLiveDatabase(t *testing.T) {
	cfg := getTestDatabaseConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatalf("expected PostgreSQL to be reachable in CI: %v", err)
		}
		t.Skipf("skipping live database test (PostgreSQL not reachable: %v)", err)
	}
	if db == nil {
		t.Fatal("expected non-nil db on success")
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping failed on live database: %v", err)
	}
}
