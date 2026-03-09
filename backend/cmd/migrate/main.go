package main

import (
	"context"
	"log"

	"github.com/mattbriggs04/bitforge/backend/internal/config"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	postgres, err := db.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer postgres.Close()

	if err := db.RunMigrations(ctx, postgres); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Printf("migrations applied")
}
