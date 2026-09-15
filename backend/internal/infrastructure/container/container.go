// Package container is the dependency-injection wiring point: it
// constructs repositories, domain services, usecases, and handlers, and
// connects them together. cmd/api/main.go depends only on this package
// (plus config/logger/database bootstrapping), never on infrastructure
// internals directly.
package container

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/your-org/your-project/backend/config"
	deliveryhttp "github.com/your-org/your-project/backend/internal/delivery/http"
	"github.com/your-org/your-project/backend/internal/delivery/http/handlers"
	"gorm.io/gorm"
)

// Container holds every wired dependency the application needs to run.
// As later cycles add repositories, services, and usecases, they get
// constructed here and threaded into the relevant handlers.
type Container struct {
	Config *config.Config
	Logger zerolog.Logger
	DB     *gorm.DB

	Handlers deliveryhttp.Handlers
}

// New wires the full dependency graph from an already-loaded config, an
// already-connected GORM database, and an already-constructed logger.
// It does not open any new connections.
func New(cfg *config.Config, db *gorm.DB, log zerolog.Logger) (*Container, error) {
	healthHandler := handlers.NewHealthHandler()

	return &Container{
		Config: cfg,
		Logger: log,
		DB:     db,
		Handlers: deliveryhttp.Handlers{
			Health: healthHandler,
		},
	}, nil
}

// Router builds the fully configured Gin engine (middleware chain +
// routes) for this container's wired handlers.
func (c *Container) Router() *gin.Engine {
	return deliveryhttp.NewRouter(c.Logger, c.Handlers)
}
