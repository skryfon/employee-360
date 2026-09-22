// Package usecaseinterface defines usecase interfaces (ports) for the application.
package usecaseinterface

import "context"

// HealthResult contains the status of the application and its dependencies.
type HealthResult struct {
	App      string
	Database string
}

// HealthUseCase defines the interface for checking system health.
type HealthUseCase interface {
	Execute(ctx context.Context) HealthResult
}
