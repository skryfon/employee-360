// Package container manages application dependency injection and wiring.
package container

import (
	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/config"
	"gorm.io/gorm"
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

	authContainer, err := NewAuthContainer(cfg, db, log, nil, nil)
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
