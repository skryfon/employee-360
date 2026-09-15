package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/your-org/your-project/backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DefaultConnTimeout is the default bounded timeout for verifying PostgreSQL connectivity.
const DefaultConnTimeout = 5 * time.Second

// Connect initializes a GORM PostgreSQL connection pool and verifies database reachability
// within a bounded timeout without retrying. Returns an error if the database is unreachable.
func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sql.DB from gorm: %w", err)
	}

	configureConnectionPool(sqlDB, cfg)

	// Bounded connectivity check (fail fast, no retry loop)
	timeout := cfg.ConnTimeout
	if timeout <= 0 {
		timeout = DefaultConnTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return db, nil
}

// EnsureDatabaseExists checks if the configured database exists. If it does not exist (SQLSTATE 3D000),
// it connects to the default maintenance database ("postgres") and creates the database automatically.
func EnsureDatabaseExists(cfg config.DatabaseConfig) error {
	// First check if direct connection succeeds
	db, err := Connect(cfg)
	if err == nil {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		return nil
	}

	errStr := err.Error()
	if !isDatabaseNotExistError(errStr) {
		// Not a "database does not exist" error; return original error (e.g. host unreachable, auth error)
		return err
	}

	// Database does not exist — connect to maintenance DB "postgres" to create it
	maintCfg := cfg
	maintCfg.DBName = "postgres"

	maintDB, maintErr := Connect(maintCfg)
	if maintErr != nil {
		// If maintenance DB is also unreachable, return original error
		return err
	}

	sqlDB, err := maintDB.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// Create database safely quoting identifier
	escapedDBName := strings.ReplaceAll(cfg.DBName, "\"", "\"\"")
	createQuery := fmt.Sprintf("CREATE DATABASE \"%s\"", escapedDBName)

	if err := maintDB.Exec(createQuery).Error; err != nil {
		return fmt.Errorf("failed to create database %q: %w", cfg.DBName, err)
	}

	return nil
}

func isDatabaseNotExistError(errStr string) bool {
	return strings.Contains(errStr, "3D000") ||
		strings.Contains(errStr, "does not exist") ||
		strings.Contains(errStr, "database \"")
}

// MustConnect establishes a GORM PostgreSQL connection or fails fast by logging
// "database unreachable: <error>" to stderr and exiting with exit code 1.
func MustConnect(cfg config.DatabaseConfig) *gorm.DB {
	_ = EnsureDatabaseExists(cfg)
	db, err := Connect(cfg)
	if err != nil {
		FailFast(err)
	}
	return db
}

// FailFast writes a single log line containing "database unreachable: <err>" to stderr
// and exits the process immediately with non-zero exit code 1.
func FailFast(err error) {
	fmt.Fprintf(os.Stderr, "database unreachable: %v\n", err)
	os.Exit(1)
}

func configureConnectionPool(sqlDB *sql.DB, cfg config.DatabaseConfig) {
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
}
