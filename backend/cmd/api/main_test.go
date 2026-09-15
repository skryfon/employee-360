package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/your-org/your-project/backend/config"
	"github.com/your-org/your-project/backend/internal/infrastructure/database"
)

func TestAPI_LiveDatabaseSuccess(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatalf("failed to load config in CI: %v", err)
		}
		t.Skipf("skipping live database test: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatalf("PostgreSQL must be reachable in CI: %v", err)
		}
		t.Skipf("skipping live database test (PostgreSQL unreachable: %v)", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = "."
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected api to succeed against live database, got err: %v, stderr: %s", err, stderr.String())
	}

	if !strings.Contains(stdout.String(), "employee360 api: connected to database") {
		t.Errorf("expected stdout to contain connection confirmation, got: %q", stdout.String())
	}
}

func TestAPI_DatabaseUnreachableFailsFast(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(),
		"DATABASE_HOST=127.0.0.1",
		"DATABASE_PORT=59999",
		"DATABASE_CONN_TIMEOUT=500ms",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected api to fail when database is unreachable, but it exited 0")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got: %v", err)
	}

	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "database unreachable") {
		t.Errorf("expected stderr to contain 'database unreachable', got: %q", stderrStr)
	}
}
