package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/config"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/httpapi"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
	"github.com/mattbriggs04/bitforge/backend/internal/service"
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
	if err := db.SeedMVP(ctx, conn); err != nil {
		log.Fatalf("seed problems: %v", err)
	}

	problemsRepo := repository.NewProblemsRepository(conn)
	submissionsRepo := repository.NewSubmissionsRepository(conn)
	usersRepo := repository.NewUsersRepository(conn)

	problemService := service.NewProblemService(problemsRepo)
	submissionService := service.NewSubmissionService(
		problemsRepo,
		submissionsRepo,
		usersRepo,
		cfg.DefaultUserHandle,
	)
	resetDatabase := func(ctx context.Context) error {
		return db.ResetDevelopmentDatabase(ctx, conn)
	}
	srv := httpapi.NewServer(
		problemService,
		submissionService,
		cfg.DefaultUserHandle,
		cfg.AppEnv != "production",
		resetDatabase,
	)

	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("http server shutdown error: %v", err)
		}
	}()

	log.Printf("API listening on :%s", cfg.HTTPPort)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api server failed: %v", err)
	}
}
