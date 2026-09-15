// Package container manages application dependency injection and wiring.
package container

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/your-org/your-project/backend/config"
	deliveryhttp "github.com/your-org/your-project/backend/internal/delivery/http"
	"github.com/your-org/your-project/backend/internal/delivery/http/handlers"
	"github.com/your-org/your-project/backend/internal/infrastructure/database"
	usecaseimpl "github.com/your-org/your-project/backend/internal/usecase/implementation"
	"gorm.io/gorm"
)

// Container holds wired dependencies and handlers for the application.
type Container struct {
	Config *config.Config
	Logger zerolog.Logger
	DB     *gorm.DB

	Handlers deliveryhttp.Handlers
}

// New wires the full dependency graph for the application.
func New(cfg *config.Config, db *gorm.DB, log zerolog.Logger) (*Container, error) {
	dbPinger := database.NewGormDatabasePinger(db)
	healthUseCase := usecaseimpl.NewHealthUseCase(dbPinger)
	healthHandler := handlers.NewHealthHandler(healthUseCase)

	return &Container{
		Config: cfg,
		Logger: log,
		DB:     db,
		Handlers: deliveryhttp.Handlers{
			Health: healthHandler,
		},
	}, nil
}

// Router constructs and returns the configured Gin engine.
func (c *Container) Router() *gin.Engine {
	return deliveryhttp.NewRouter(c.Logger, c.Config.CORS.AllowedOrigins, c.Handlers)
}
