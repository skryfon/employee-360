// Package usecaseimpl contains application usecase implementations.
package usecaseimpl

import (
	"context"

	"github.com/skryfon/employee360/backend/internal/domain/service"
	usecaseinterface "github.com/skryfon/employee360/backend/internal/usecase/interface"
	"github.com/skryfon/employee360/backend/shared"
)

// HealthUseCaseImpl implements usecaseinterface.HealthUseCase.
type HealthUseCaseImpl struct {
	pinger service.DatabasePinger
}

var _ usecaseinterface.HealthUseCase = (*HealthUseCaseImpl)(nil)

// NewHealthUseCase constructs a HealthUseCaseImpl with the given database pinger.
func NewHealthUseCase(pinger service.DatabasePinger) *HealthUseCaseImpl {
	return &HealthUseCaseImpl{pinger: pinger}
}

// Execute checks application and database health status.
func (u *HealthUseCaseImpl) Execute(ctx context.Context) usecaseinterface.HealthResult {
	database := "ok"
	if err := u.pinger.Ping(ctx); err != nil {
		database = "unreachable"
	}

	return usecaseinterface.HealthResult{
		App:      shared.AppName,
		Database: database,
	}
}
