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
	UserKey     string
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
		return nil, newInvalidError("problemSlug is required")
	}
	language := strings.ToLower(strings.TrimSpace(input.Language))
	if language == "" {
		language = "c"
	}
	if language != "c" {
		return nil, newUnsupportedError(fmt.Sprintf("language %q is not supported yet", language))
	}
	mode := strings.ToLower(strings.TrimSpace(input.Mode))
	if mode == "" {
		mode = "submit"
	}
	if mode != "run" && mode != "submit" {
		return nil, newInvalidError("mode must be either run or submit")
	}
	source := strings.TrimSpace(input.SourceCode)
	if source == "" {
		return nil, newInvalidError("sourceCode is required")
	}
	if len(source) > 300_000 {
		return nil, newInvalidError("sourceCode exceeds maximum size")
	}

	problem, err := s.problems.GetBySlug(ctx, problemSlug)
	if err != nil {
		return nil, fmt.Errorf("load problem: %w", err)
	}
	if problem == nil {
		return nil, newNotFoundError("problem not found")
	}

	starter, err := s.problems.GetStarterCode(ctx, problem.ID, language)
	if err != nil {
		return nil, fmt.Errorf("load language template: %w", err)
	}
	if starter == "" {
		return nil, newUnsupportedError(fmt.Sprintf("problem %q does not support language %q", problemSlug, language))
	}

	userHandle := strings.TrimSpace(input.UserHandle)
	if userHandle == "" {
		userHandle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureIdentity(ctx, input.UserKey, userHandle)
	if err != nil {
		if errors.Is(err, repository.ErrHandleTaken) {
			return nil, newConflictError("username is already taken")
		}
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
