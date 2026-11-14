package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

type SubmissionService struct {
	problems          *repository.ProblemsRepository
	submissions       *repository.SubmissionsRepository
	users             *repository.UsersRepository
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
	defaultUserHandle string,
) *SubmissionService {
	return &SubmissionService{
		problems:          problems,
		submissions:       submissions,
		users:             users,
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
	if language != "c" && language != "cpp" && language != "python" && language != "rust" && language != "answer" {
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

	if language == "answer" {
		cfg, e := s.problems.GetJudgeConfig(ctx, problem.ID)
		if e != nil {
			return nil, e
		}
		if cfg == nil || cfg.Runner != "answer_v1" {
			return nil, newUnsupportedError("problem requires source code")
		}
	}

	starter, err := s.problems.GetStarterCode(ctx, problem.ID, language)
	if err != nil {
		return nil, fmt.Errorf("load language template: %w", err)
	}
	if starter == "" && language != "answer" {
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

func (s *SubmissionService) CompletedProblemIDs(ctx context.Context, userHandle, userKey string) (map[string]bool, error) {
	handle := strings.TrimSpace(userHandle)
	if handle == "" {
		handle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureIdentity(ctx, userKey, handle)
	if err != nil {
		if errors.Is(err, repository.ErrHandleTaken) {
			return nil, newConflictError("username is already taken")
		}
		return nil, fmt.Errorf("ensure user: %w", err)
	}
	return s.submissions.AcceptedProblemIDs(ctx, userID)
}

// Explanation is returned only for a completed submit owned by this local identity.
func (s *SubmissionService) Explanation(ctx context.Context, slug, id, key, handle string) (string, error) {
	attempt, err := s.submissions.GetForRunner(ctx, id)
	if err != nil {
		return "", err
	}
	if attempt == nil || attempt.ProblemSlug != slug {
		return "", newNotFoundError("attempt not found")
	}
	if handle == "" {
		handle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureIdentity(ctx, key, handle)
	if err != nil {
		return "", err
	}
	if userID != attempt.UserID || attempt.Mode != "submit" || attempt.Status != "completed" {
		return "", &AppError{Kind: ErrorKindForbidden, Message: "complete your own submission before revealing the explanation"}
	}
	snapshot, err := s.submissions.LoadSnapshot(ctx, id)
	if err != nil {
		return "", err
	}
	text, _ := snapshot.Config.Config["explanation"].(string)
	return text, nil
}
