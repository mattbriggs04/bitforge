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
	"github.com/mattbriggs04/bitforge/backend/internal/queue"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
	"github.com/mattbriggs04/bitforge/backend/internal/service"
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

	redisClient := queue.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("connect redis: %v", err)
	}

	problemsRepo := repository.NewProblemsRepository(postgres)
	submissionsRepo := repository.NewSubmissionsRepository(postgres)
	competitionsRepo := repository.NewCompetitionsRepository(postgres)
	usersRepo := repository.NewUsersRepository(postgres)
	submissionQueue := queue.NewRedisSubmissionQueue(redisClient, cfg.SubmissionQueue)

	problemService := service.NewProblemService(problemsRepo)
	submissionService := service.NewSubmissionService(
		problemsRepo,
		submissionsRepo,
		usersRepo,
		submissionQueue,
		cfg.DefaultUserHandle,
	)
	competitionService := service.NewCompetitionService(competitionsRepo, usersRepo, cfg.DefaultUserHandle)

	srv := httpapi.NewServer(problemService, submissionService, competitionService, cfg.DefaultUserHandle)

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
