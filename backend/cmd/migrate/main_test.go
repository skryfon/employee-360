package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/your-project/backend/config"
	"github.com/your-org/your-project/backend/internal/infrastructure/database"
)

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
	cmdUp.Dir = "."
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
	cmdStatus.Dir = "."
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
	cmdUp.Dir = "."
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
	cmdStatus.Dir = "."
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
	cmdDown.Dir = "."
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

func TestMigrate_DatabaseUnreachableFailsFast(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "up")
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
		t.Fatal("expected migrate to fail when database is unreachable, but it exited 0")
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

func TestFindMigrationsDir(t *testing.T) {
	dir := findMigrationsDir()
	if dir == "" {
		t.Fatal("expected findMigrationsDir to return a non-empty string")
	}
}

func TestRunCreate_GeneratesMigrationPair(t *testing.T) {
	tempDir := t.TempDir()

	runCreate(tempDir, "create_users_table")

	upFile := filepath.Join(tempDir, "000001_create_users_table.up.sql")
	downFile := filepath.Join(tempDir, "000001_create_users_table.down.sql")

	if _, err := os.Stat(upFile); err != nil {
		t.Fatalf("expected up migration file %s to exist: %v", upFile, err)
	}
	if _, err := os.Stat(downFile); err != nil {
		t.Fatalf("expected down migration file %s to exist: %v", downFile, err)
	}

	// Create second migration and verify sequence increments to 000002
	runCreate(tempDir, "add_user_roles")

	upFile2 := filepath.Join(tempDir, "000002_add_user_roles.up.sql")
	downFile2 := filepath.Join(tempDir, "000002_add_user_roles.down.sql")

	if _, err := os.Stat(upFile2); err != nil {
		t.Fatalf("expected second up migration file %s to exist: %v", upFile2, err)
	}
	if _, err := os.Stat(downFile2); err != nil {
		t.Fatalf("expected second down migration file %s to exist: %v", downFile2, err)
	}
}
