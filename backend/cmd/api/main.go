package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/your-project/backend/config"
	"github.com/your-org/your-project/backend/internal/infrastructure/container"
	"github.com/your-org/your-project/backend/internal/infrastructure/database"
	"github.com/your-org/your-project/backend/internal/infrastructure/server"
	"github.com/your-org/your-project/backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.LogLevel)

	// Connect to database (fail fast if unreachable).
	db, err := database.Connect(cfg.Database)
	if err != nil {
		database.FailFast(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get underlying database connection: %v\n", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	log.Info().
		Str("database", cfg.Database.DBName).
		Str("host", cfg.Database.Host).
		Str("port", cfg.Database.Port).
		Msg("employee360 api: connected to database")

	// Wire dependencies.
	c, err := container.New(cfg, db, log)
	if err != nil {
		log.Error().Err(err).Msg("failed to wire dependencies")
		os.Exit(1)
	}

	srv := server.New(cfg.Server, c.Router())

	// Run the server in the background so the main goroutine is free to
	// wait for either a shutdown signal or a server startup failure.
	serverErrCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", srv.Addr()).Msg("employee360 api: starting server")
		serverErrCh <- srv.Start()
	}()

	// Wait for SIGINT/SIGTERM to trigger a graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrCh:
		if err != nil {
			log.Error().Err(err).Msg("employee360 api: server failed")
			os.Exit(1)
		}
	case <-ctx.Done():
		log.Info().Msg("employee360 api: shutdown signal received")
		stop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("employee360 api: graceful shutdown failed")
			os.Exit(1)
		}

		log.Info().Msg("employee360 api: shutdown complete")
	}
}
