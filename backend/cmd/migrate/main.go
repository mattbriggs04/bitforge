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
	conn, err := db.OpenDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer conn.Close()

	if err := db.RunMigrations(ctx, conn); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Printf("migrations applied")
}
