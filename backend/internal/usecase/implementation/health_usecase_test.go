package usecaseimpl

import (
	"context"
	"errors"
	"testing"
)

// fakeDatabasePinger implements service.DatabasePinger for tests.
type fakeDatabasePinger struct {
	err error
}

func (f *fakeDatabasePinger) Ping(ctx context.Context) error {
	return f.err
}

func TestHealthUseCase_Execute_DatabaseOK(t *testing.T) {
	uc := NewHealthUseCase(&fakeDatabasePinger{})

	result := uc.Execute(context.Background())

	if result.Database != "ok" {
		t.Errorf("expected Database 'ok', got %q", result.Database)
	}
	if result.App == "" {
		t.Error("expected App to be populated")
	}
}

func TestHealthUseCase_Execute_DatabaseUnreachable(t *testing.T) {
	uc := NewHealthUseCase(&fakeDatabasePinger{err: errors.New("connection refused")})

	result := uc.Execute(context.Background())

	if result.Database != "unreachable" {
		t.Errorf("expected Database 'unreachable', got %q", result.Database)
	}
}
