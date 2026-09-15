package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
