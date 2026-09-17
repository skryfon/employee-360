//go:build integration

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

// migrateDir is cmd/migrate's location relative to this package
// (integration/), used as the working directory for "go run ."
// invocations below.
const migrateDir = "../cmd/migrate"

func checkLiveDBOrSkip(t *testing.T) {
	t.Helper()
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
}

func TestMigrate_LiveDatabaseZeroMigrationsClean(t *testing.T) {
	checkLiveDBOrSkip(t)

	// Test "up" with zero migrations (AC-4 & AC-5)
	cmdUp := exec.Command("go", "run", ".", "up")
	cmdUp.Dir = migrateDir
	var stdoutUp, stderrUp bytes.Buffer
	cmdUp.Stdout = &stdoutUp
	cmdUp.Stderr = &stderrUp

	if err := cmdUp.Run(); err != nil {
		t.Fatalf("expected migrate up to succeed with zero migrations, got err: %v, stderr: %s", err, stderrUp.String())
	}
	if !strings.Contains(stdoutUp.String(), "migrations applied successfully") {
		t.Errorf("expected stdout to contain 'migrations applied successfully', got: %q", stdoutUp.String())
	}

	// Test "status" with zero migrations
	cmdStatus := exec.Command("go", "run", ".", "status")
	cmdStatus.Dir = migrateDir
	var stdoutStatus, stderrStatus bytes.Buffer
	cmdStatus.Stdout = &stdoutStatus
	cmdStatus.Stderr = &stderrStatus

	if err := cmdStatus.Run(); err != nil {
		t.Fatalf("expected migrate status to succeed, got err: %v, stderr: %s", err, stderrStatus.String())
	}
	if !strings.Contains(stdoutStatus.String(), "no migrations applied yet") {
		t.Errorf("expected stdout to contain 'no migrations applied yet', got: %q", stdoutStatus.String())
	}
}

func TestMigrate_LiveDatabaseApplyAndRevert(t *testing.T) {
	checkLiveDBOrSkip(t)

	tempDir := t.TempDir()
	upContent := "CREATE TABLE IF NOT EXISTS _smoke_test_table (id serial primary key, name text);\n"
	downContent := "DROP TABLE IF EXISTS _smoke_test_table;\n"

	if err := os.WriteFile(filepath.Join(tempDir, "000001_smoke_test.up.sql"), []byte(upContent), 0644); err != nil {
		t.Fatalf("failed to write up migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "000001_smoke_test.down.sql"), []byte(downContent), 0644); err != nil {
		t.Fatalf("failed to write down migration: %v", err)
	}

	// 1. Run UP
	cmdUp := exec.Command("go", "run", ".", "-dir", tempDir, "up")
	cmdUp.Dir = migrateDir
	var stdoutUp, stderrUp bytes.Buffer
	cmdUp.Stdout = &stdoutUp
	cmdUp.Stderr = &stderrUp

	if err := cmdUp.Run(); err != nil {
		t.Fatalf("expected migrate up to succeed, got: %v, stderr: %s", err, stderrUp.String())
	}
	if !strings.Contains(stdoutUp.String(), "migrations applied successfully") {
		t.Errorf("expected stdout to contain 'migrations applied successfully', got: %q", stdoutUp.String())
	}

	// 2. Check STATUS
	cmdStatus := exec.Command("go", "run", ".", "-dir", tempDir, "status")
	cmdStatus.Dir = migrateDir
	var stdoutStatus, stderrStatus bytes.Buffer
	cmdStatus.Stdout = &stdoutStatus
	cmdStatus.Stderr = &stderrStatus

	if err := cmdStatus.Run(); err != nil {
		t.Fatalf("expected migrate status to succeed, got: %v, stderr: %s", err, stderrStatus.String())
	}
	if !strings.Contains(stdoutStatus.String(), "current version: 1") {
		t.Errorf("expected status to contain 'current version: 1', got: %q", stdoutStatus.String())
	}

	// 3. Run DOWN
	cmdDown := exec.Command("go", "run", ".", "-dir", tempDir, "down", "1")
	cmdDown.Dir = migrateDir
	var stdoutDown, stderrDown bytes.Buffer
	cmdDown.Stdout = &stdoutDown
	cmdDown.Stderr = &stderrDown

	if err := cmdDown.Run(); err != nil {
		t.Fatalf("expected migrate down to succeed, got: %v, stderr: %s", err, stderrDown.String())
	}
	if !strings.Contains(stdoutDown.String(), "rolled back 1 migration(s)") {
		t.Errorf("expected stdout to contain 'rolled back 1 migration(s)', got: %q", stdoutDown.String())
	}
}
