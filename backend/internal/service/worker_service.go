package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/judge"
	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/queue"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

type WorkerService struct {
	submissions     *repository.SubmissionsRepository
	problems        *repository.ProblemsRepository
	submissionQueue queue.SubmissionQueue
	judge           *judge.Service
	compiler        string
	compileTimeout  time.Duration
	runTimeout      time.Duration
	popTimeout      time.Duration
}

func NewWorkerService(
	submissions *repository.SubmissionsRepository,
	problems *repository.ProblemsRepository,
	submissionQueue queue.SubmissionQueue,
	judgeService *judge.Service,
	compiler string,
	compileTimeout time.Duration,
	runTimeout time.Duration,
	popTimeout time.Duration,
) *WorkerService {
	return &WorkerService{
		submissions:     submissions,
		problems:        problems,
		submissionQueue: submissionQueue,
		judge:           judgeService,
		compiler:        compiler,
		compileTimeout:  compileTimeout,
		runTimeout:      runTimeout,
		popTimeout:      popTimeout,
	}
}

func (s *WorkerService) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		submissionID, err := s.submissionQueue.Dequeue(ctx, s.popTimeout)
		if err != nil {
			if err == queue.ErrNoJob {
				continue
			}
			return fmt.Errorf("dequeue submission: %w", err)
		}

		if err := s.processSubmission(ctx, submissionID); err != nil {
			log.Printf("worker failed submission %s: %v", submissionID, err)
			_ = s.submissions.Fail(ctx, submissionID, err.Error())
		}
	}
}

func (s *WorkerService) processSubmission(ctx context.Context, submissionID string) error {
	submission, err := s.submissions.GetForWorker(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("load submission: %w", err)
	}
	if submission == nil {
		return fmt.Errorf("submission %s not found", submissionID)
	}

	if err := s.submissions.MarkRunning(ctx, submissionID); err != nil {
		return fmt.Errorf("mark submission running: %w", err)
	}

	includeHidden := submission.Mode == "submit"
	testCases, err := s.problems.GetTestCases(ctx, submission.ProblemID, includeHidden)
	if err != nil {
		return fmt.Errorf("load test cases: %w", err)
	}
	if len(testCases) == 0 {
		return fmt.Errorf("no test cases available")
	}

	judgeConfig, err := s.problems.GetJudgeConfig(ctx, submission.ProblemID)
	if err != nil {
		return fmt.Errorf("load judge config: %w", err)
	}
	if judgeConfig == nil {
		judgeConfig = &model.JudgeConfig{Runner: "c_assert_harness_v1", Config: map[string]any{}}
	}

	compileTimeout := s.compileTimeout
	runTimeout := s.runTimeout
	if value := configDurationMS(judgeConfig.Config, "compile_timeout_ms"); value > 0 {
		compileTimeout = value
	}
	if value := configDurationMS(judgeConfig.Config, "run_timeout_ms"); value > 0 {
		runTimeout = value
	}

	result, err := s.judge.Evaluate(ctx, judge.Request{
		SubmissionID:   submission.ID,
		Language:       submission.Language,
		SourceCode:     submission.SourceCode,
		Cases:          testCases,
		Config:         judgeConfig.Config,
		Compiler:       s.compiler,
		CompileTimeout: compileTimeout,
		RunTimeout:     runTimeout,
	})
	if err != nil {
		return fmt.Errorf("evaluate submission: %w", err)
	}

	if err := s.submissions.Complete(ctx, submissionID, result); err != nil {
		return fmt.Errorf("persist judge result: %w", err)
	}

	return nil
}

func configDurationMS(config map[string]any, key string) time.Duration {
	if config == nil {
		return 0
	}
	value, ok := config[key]
	if !ok {
		return 0
	}
	number, ok := value.(float64)
	if !ok || number <= 0 {
		return 0
	}
	return time.Duration(number) * time.Millisecond
}
