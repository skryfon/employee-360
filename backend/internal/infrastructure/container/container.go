// Package container manages application dependency injection and wiring.
package container

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/config"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
	"github.com/skryfon/employee360/backend/internal/infrastructure/eventing"
	infraservice "github.com/skryfon/employee360/backend/internal/infrastructure/service"
	"github.com/skryfon/employee360/backend/internal/usecase/implementation/ucshared"
)

// AppContainer holds wired configuration, infrastructure, and domain sub-containers.
type AppContainer struct {
	Config *config.Config
	Log    zerolog.Logger
	DB     *gorm.DB

	Health *HealthContainer
	Auth   *AuthContainer
}

// NewAppContainer wires the full dependency graph for the application.
func NewAppContainer(cfg *config.Config, db *gorm.DB, log zerolog.Logger) (*AppContainer, error) {
	healthContainer := NewHealthContainer(db)

	// Transactional outbox: the GORM transactor and River-backed publisher share
	// the in-flight transaction via context. A nil db (construction-only tests)
	// falls back to the no-op transactor/publisher inside NewAuthContainer.
	var (
		transactor ucshared.Transactor
		publisher  domainservice.EventPublisher
	)
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("container: resolve sql.DB: %w", err)
		}
		riverClient, err := eventing.NewInsertOnlyClient(sqlDB)
		if err != nil {
			return nil, fmt.Errorf("container: build river insert client: %w", err)
		}
		transactor = database.NewGormTransactor(db)
		publisher = eventing.NewRiverPublisher(riverClient, eventing.NewDispatcher())
	}

	authContainer, err := NewAuthContainer(cfg, db, log, transactor, publisher)
	if err != nil {
		return nil, err
	}

	return &AppContainer{
		Config: cfg,
		Log:    log,
		DB:     db,
		Health: healthContainer,
		Auth:   authContainer,
	}, nil
}

// Container is an alias for AppContainer.
type Container = AppContainer

// New is an alias for NewAppContainer.
func New(cfg *config.Config, db *gorm.DB, log zerolog.Logger) (*AppContainer, error) {
	return NewAppContainer(cfg, db, log)
}

// WorkerContainer holds the dependencies of the cmd/worker process.
type WorkerContainer struct {
	RiverClient *river.Client[*sql.Tx]
}

// NewWorkerContainer wires the River worker client with the SMTP EmailService.
func NewWorkerContainer(cfg *config.Config, db *gorm.DB) (*WorkerContainer, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("container: resolve sql.DB: %w", err)
	}
	emailSvc := infraservice.NewSMTPMailService(cfg.SMTP)
	slogger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	client, err := eventing.NewWorkerClient(sqlDB, emailSvc, slogger, 0)
	if err != nil {
		return nil, fmt.Errorf("container: build river worker client: %w", err)
	}
	return &WorkerContainer{RiverClient: client}, nil
}
