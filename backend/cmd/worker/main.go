package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattbriggs04/bitforge/backend/internal/config"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/judge"
	"github.com/mattbriggs04/bitforge/backend/internal/queue"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
	"github.com/mattbriggs04/bitforge/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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

	submissionsRepo := repository.NewSubmissionsRepository(postgres)
	problemsRepo := repository.NewProblemsRepository(postgres)
	submissionQueue := queue.NewRedisSubmissionQueue(redisClient, cfg.SubmissionQueue)

	judgeService := judge.NewService(judge.NewCAssertRunner())
	worker := service.NewWorkerService(
		submissionsRepo,
		problemsRepo,
		submissionQueue,
		judgeService,
		cfg.CCompiler,
		cfg.CompileTimeout,
		cfg.RunTimeout,
		cfg.QueuePopTimeout,
	)

	log.Printf("worker started. queue=%s", cfg.SubmissionQueue)
	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("worker stopped with error: %v", err)
		os.Exit(1)
	}
	log.Printf("worker stopped")
}
