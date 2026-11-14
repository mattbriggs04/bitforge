package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/config"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/judge"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
	"github.com/mattbriggs04/bitforge/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	conn, err := db.OpenDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	if err = db.RunMigrations(ctx, conn); err != nil {
		log.Fatal(err)
	}
	submissions := repository.NewSubmissionsRepository(conn)
	runner := service.NewSubmissionRunnerService(submissions, repository.NewProblemsRepository(conn), judge.NewService(judge.NewCAssertRunner()), cfg.CompileTimeout, cfg.RunTimeout)
	workers, _ := strconv.Atoi(os.Getenv("JUDGE_WORKERS"))
	if workers < 1 {
		workers = 2
	}
	if workers > 8 {
		workers = 8
	}
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				job, err := submissions.Claim(ctx)
				if err != nil {
					log.Printf("claim: %v", err)
				}
				if job != nil {
					jobCtx, cancel := context.WithTimeout(ctx, 8*time.Minute)
					err = runner.ProcessSubmission(jobCtx, job.SubmissionID, job.Token)
					cancel()
					if err != nil {
						log.Printf("judge %s: %v", job.SubmissionID, err)
						retryCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						if e := submissions.Retry(retryCtx, *job, err); e != nil {
							log.Printf("retry: %v", e)
						}
						cancel()
					}
					continue
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(400 * time.Millisecond):
				}
			}
		}()
	}
	log.Printf("judge worker ready (%d slots)", workers)
	wg.Wait()
}
