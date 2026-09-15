package database

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/your-org/your-project/backend/config"
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

	db, err := Connect(cfg)
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

func TestConnect_UnreachableDatabaseFailsFast(t *testing.T) {
	// Point to a non-existent port with a short timeout
	cfg := config.DatabaseConfig{
		Host:        "127.0.0.1",
		Port:        "59999",
		User:        "postgres",
		Password:    "postgres",
		DBName:      "nonexistent",
		SSLMode:     "disable",
		ConnTimeout: 1 * time.Second,
	}

	start := time.Now()
	db, err := Connect(cfg)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("expected Connect to fail on unreachable host, but got nil error")
	}
	if db != nil {
		t.Fatal("expected returned db to be nil on failure")
	}

	// Verify bounded timeout (no retry loop)
	if duration > 3*time.Second {
		t.Errorf("connection attempt took %v, expected fail-fast within ~1s", duration)
	}
}

// TestMustConnect_ProcessHelper is executed as a subprocess by TestMustConnect_ExitsNonZeroWithStderr.
func TestMustConnect_ProcessHelper(t *testing.T) {
	if os.Getenv("BE_FAIL_FAST_PROCESS") != "1" {
		return
	}

	cfg := config.DatabaseConfig{
		Host:        "127.0.0.1",
		Port:        "59999",
		User:        "postgres",
		Password:    "postgres",
		DBName:      "employee360",
		SSLMode:     "disable",
		ConnTimeout: 500 * time.Millisecond,
	}

	// This must fail fast, log to stderr containing "database unreachable", and exit code 1
	MustConnect(cfg)
}

func TestMustConnect_ExitsNonZeroWithStderr(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestMustConnect_ProcessHelper")
	cmd.Env = append(os.Environ(), "BE_FAIL_FAST_PROCESS=1")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected process to exit with non-zero status code")
	}

	exitError, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}

	if exitError.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitError.ExitCode())
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "database unreachable") {
		t.Errorf("expected stderr to contain 'database unreachable', got: %q", stderrStr)
	}
}
