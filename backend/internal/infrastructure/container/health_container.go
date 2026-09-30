package container

import (
	"github.com/skryfon/employee360/backend/internal/delivery/http/handlers"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
	usecaseimpl "github.com/skryfon/employee360/backend/internal/usecase/implementation"
	usecaseinterface "github.com/skryfon/employee360/backend/internal/usecase/interface"
	"gorm.io/gorm"
)

// HealthContainer encapsulates dependencies, usecases, and handlers for health checks.
type HealthContainer struct {
	Handler *handlers.HealthHandler
	UseCase usecaseinterface.HealthUseCase
}

// NewHealthContainer initializes and wires the health check usecase and handler.
func NewHealthContainer(db *gorm.DB) *HealthContainer {
	dbPinger := database.NewGormDatabasePinger(db)
	healthUseCase := usecaseimpl.NewHealthUseCase(dbPinger)
	healthHandler := handlers.NewHealthHandler(healthUseCase)

	return &HealthContainer{
		Handler: healthHandler,
		UseCase: healthUseCase,
	}
}
