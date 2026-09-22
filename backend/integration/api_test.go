//go:build integration

package integration

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

func isCI() bool {
	return os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"
}

// buildAPIBinary compiles cmd/api into a temp binary from the backend
// module root (this package lives one level below it, at integration/).
// Running a real binary (rather than "go run .") makes SIGTERM delivery
// deterministic, which the graceful-shutdown assertions below depend on.
func buildAPIBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "api-test-bin")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/api")
	cmd.Dir = ".."

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build api binary: %v, stderr: %s", err, stderr.String())
	}
	return binPath
}

// TestAPI_LiveDatabaseSuccess starts the real api binary against a live
// database, waits for the health endpoint to come up, verifies it
// reports 200, then sends SIGTERM and verifies the process shuts down
// gracefully (exit code 0) within the configured shutdown timeout.
func TestAPI_LiveDatabaseSuccess(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		if isCI() {
			t.Fatalf("failed to load config in CI: %v", err)
		}
		t.Skipf("skipping live database test: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		if isCI() {
			t.Fatalf("PostgreSQL must be reachable in CI: %v", err)
		}
		t.Skipf("skipping live database test (PostgreSQL unreachable: %v)", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	binPath := buildAPIBinary(t)

	const port = "18089"
	cmd := exec.Command(binPath)
	cmd.Env = append(os.Environ(),
		"SERVER_HOST=127.0.0.1",
		"SERVER_PORT="+port,
		"SERVER_SHUTDOWN_TIMEOUT=5s",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start api binary: %v", err)
	}

	healthURL := "http://127.0.0.1:" + port + "/api/v1/health"
	var getErr error
	var resp *http.Response
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		resp, getErr = http.Get(healthURL)
		if getErr == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if getErr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("health endpoint never became reachable: %v, stderr: %s", getErr, stderr.String())
	}
	body := make([]byte, 4096)
	n, _ := resp.Body.Read(body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from health endpoint, got %d, body: %s", resp.StatusCode, body[:n])
	}
	if !strings.Contains(string(body[:n]), `"database":"ok"`) {
		t.Errorf("expected health response body to report database ok (live Postgres is reachable at this point), got: %s", body[:n])
	}
	if !strings.Contains(string(body[:n]), `"status":"ok"`) {
		t.Errorf("expected health response body to report ok status, got: %s", body[:n])
	}

	// Trigger graceful shutdown.
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	waitErrCh := make(chan error, 1)
	go func() { waitErrCh <- cmd.Wait() }()

	select {
	case err := <-waitErrCh:
		if err != nil {
			t.Fatalf("expected graceful exit (code 0) after SIGTERM, got err: %v, stderr: %s", err, stderr.String())
		}
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("server did not shut down within timeout after SIGTERM, stderr: %s", stderr.String())
	}

	if !strings.Contains(stdout.String(), "employee360 api: connected to database") {
		t.Errorf("expected stdout to contain connection confirmation, got: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "shutdown complete") {
		t.Errorf("expected stdout to contain graceful shutdown confirmation, got: %q", stdout.String())
	}
}
