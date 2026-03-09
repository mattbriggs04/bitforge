package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/queue"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

type SubmissionService struct {
	problems          *repository.ProblemsRepository
	submissions       *repository.SubmissionsRepository
	users             *repository.UsersRepository
	submissionQueue   queue.SubmissionQueue
	defaultUserHandle string
}

type CreateSubmissionInput struct {
	ProblemSlug string
	Language    string
	Mode        string
	SourceCode  string
	UserHandle  string
}

type CreateSubmissionOutput struct {
	SubmissionID string `json:"submissionId"`
	Status       string `json:"status"`
}

func NewSubmissionService(
	problems *repository.ProblemsRepository,
	submissions *repository.SubmissionsRepository,
	users *repository.UsersRepository,
	submissionQueue queue.SubmissionQueue,
	defaultUserHandle string,
) *SubmissionService {
	return &SubmissionService{
		problems:          problems,
		submissions:       submissions,
		users:             users,
		submissionQueue:   submissionQueue,
		defaultUserHandle: defaultUserHandle,
	}
}

func (s *SubmissionService) Create(ctx context.Context, input CreateSubmissionInput) (*CreateSubmissionOutput, error) {
	problemSlug := strings.TrimSpace(input.ProblemSlug)
	if problemSlug == "" {
		return nil, errors.New("problemSlug is required")
	}
	language := strings.ToLower(strings.TrimSpace(input.Language))
	if language == "" {
		language = "c"
	}
	if language != "c" {
		return nil, fmt.Errorf("language %q is not supported yet", language)
	}
	mode := strings.ToLower(strings.TrimSpace(input.Mode))
	if mode == "" {
		mode = "submit"
	}
	if mode != "run" && mode != "submit" {
		return nil, errors.New("mode must be either run or submit")
	}
	source := strings.TrimSpace(input.SourceCode)
	if source == "" {
		return nil, errors.New("sourceCode is required")
	}
	if len(source) > 300_000 {
		return nil, errors.New("sourceCode exceeds maximum size")
	}

	problem, err := s.problems.GetBySlug(ctx, problemSlug)
	if err != nil {
		return nil, fmt.Errorf("load problem: %w", err)
	}
	if problem == nil {
		return nil, errors.New("problem not found")
	}

	starter, err := s.problems.GetStarterCode(ctx, problem.ID, language)
	if err != nil {
		return nil, fmt.Errorf("load language template: %w", err)
	}
	if starter == "" {
		return nil, fmt.Errorf("problem %q does not support language %q", problemSlug, language)
	}

	userHandle := strings.TrimSpace(input.UserHandle)
	if userHandle == "" {
		userHandle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureByHandle(ctx, userHandle)
	if err != nil {
		return nil, fmt.Errorf("ensure user: %w", err)
	}

	submissionID, err := s.submissions.Create(ctx, model.NewSubmission{
		UserID:     userID,
		ProblemID:  problem.ID,
		Language:   language,
		Mode:       mode,
		SourceCode: input.SourceCode,
	})
	if err != nil {
		return nil, fmt.Errorf("create submission: %w", err)
	}

	if err := s.submissionQueue.Enqueue(ctx, submissionID); err != nil {
		return nil, fmt.Errorf("queue submission: %w", err)
	}

	return &CreateSubmissionOutput{SubmissionID: submissionID, Status: "queued"}, nil
}

func (s *SubmissionService) GetByID(ctx context.Context, submissionID string) (*model.Submission, error) {
	submission, err := s.submissions.GetByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	if submission == nil {
		return nil, nil
	}

	hiddenCaseCount := 0
	for i := range submission.Results {
		if submission.Results[i].IsHidden {
			hiddenCaseCount++
			submission.Results[i].CaseName = fmt.Sprintf("hidden_case_%d", hiddenCaseCount)
			if submission.Results[i].Status == "failed" || submission.Results[i].Status == "error" {
				submission.Results[i].Message = "hidden test did not pass"
			}
		}
	}
	if submission.Results == nil {
		submission.Results = []model.SubmissionTestResult{}
	}

	return submission, nil
}
