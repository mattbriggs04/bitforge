package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/judge"
	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

type SubmissionRunnerService struct {
	submissions    *repository.SubmissionsRepository
	problems       *repository.ProblemsRepository
	judge          *judge.Service
	compileTimeout time.Duration
	runTimeout     time.Duration
}

func NewSubmissionRunnerService(
	submissions *repository.SubmissionsRepository,
	problems *repository.ProblemsRepository,
	judgeService *judge.Service,
	compileTimeout time.Duration,
	runTimeout time.Duration,
) *SubmissionRunnerService {
	return &SubmissionRunnerService{
		submissions:    submissions,
		problems:       problems,
		judge:          judgeService,
		compileTimeout: compileTimeout,
		runTimeout:     runTimeout,
	}
}

func (s *SubmissionRunnerService) ProcessSubmission(ctx context.Context, submissionID, token string) error {
	submission, err := s.submissions.GetForRunner(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("load submission: %w", err)
	}
	if submission == nil {
		return fmt.Errorf("submission %s not found", submissionID)
	}

	if err := s.submissions.MarkRunning(ctx, submissionID, token); err != nil {
		return fmt.Errorf("mark submission running: %w", err)
	}

	snapshot, err := s.submissions.LoadSnapshot(ctx, submissionID)
	if err != nil {
		return err
	}
	testCases := []model.JudgeTestCase{}
	for _, tc := range snapshot.Cases {
		if submission.Mode == "submit" || !tc.IsHidden || snapshot.Config.Runner == "answer_v1" {
			testCases = append(testCases, tc)
		}
	}
	if len(testCases) == 0 {
		return fmt.Errorf("no test cases available")
	}
	judgeConfig := &snapshot.Config

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
		Runner:         judgeConfig.Runner,
		Language:       submission.Language,
		SourceCode:     submission.SourceCode,
		Cases:          testCases,
		Config:         judgeConfig.Config,
		CompileTimeout: compileTimeout,
		RunTimeout:     runTimeout,
	})
	if err != nil {
		return fmt.Errorf("evaluate submission: %w", err)
	}

	if err := s.submissions.Complete(ctx, submissionID, token, result); err != nil {
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
