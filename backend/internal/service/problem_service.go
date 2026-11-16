package service

import (
	"context"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

type ProblemService struct {
	problems *repository.ProblemsRepository
}

func NewProblemService(problems *repository.ProblemsRepository) *ProblemService {
	return &ProblemService{problems: problems}
}

func (s *ProblemService) List(ctx context.Context, filter model.ProblemFilter) ([]model.ProblemSummary, error) {
	return s.problems.ListPublished(ctx, filter)
}

func (s *ProblemService) GetBySlug(ctx context.Context, slug string) (*model.ProblemDetail, error) {
	return s.problems.GetPublishedBySlug(ctx, slug)
}

func (s *ProblemService) ListInterviews(ctx context.Context) ([]model.InterviewSummary, error) {
	return s.problems.ListInterviews(ctx, "")
}
func (s *ProblemService) GetInterview(ctx context.Context, slug string) (*model.InterviewDetail, error) {
	return s.problems.GetInterview(ctx, slug)
}
