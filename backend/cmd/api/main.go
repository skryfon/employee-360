package main

import (
	"fmt"
	"os"

	"github.com/your-org/your-project/backend/config"
	"github.com/your-org/your-project/backend/internal/infrastructure/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Connect to database (fail fast if unreachable)
	if err := database.EnsureDatabaseExists(cfg.Database); err != nil {
		database.FailFast(err)
	}

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

	fmt.Printf("employee360 api: connected to database %s on %s:%s\n", cfg.Database.DBName, cfg.Database.Host, cfg.Database.Port)
}
