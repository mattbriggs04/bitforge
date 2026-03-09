package judge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
)

type Request struct {
	SubmissionID   string
	Language       string
	SourceCode     string
	Cases          []model.JudgeTestCase
	Config         map[string]any
	Compiler       string
	CompileTimeout time.Duration
	RunTimeout     time.Duration
}

type Engine interface {
	Evaluate(ctx context.Context, req Request) (model.JudgeResult, error)
}

type Service struct {
	c Engine
}

func NewService(c Engine) *Service {
	return &Service{c: c}
}

func (s *Service) Evaluate(ctx context.Context, req Request) (model.JudgeResult, error) {
	if strings.ToLower(req.Language) != "c" {
		return model.JudgeResult{}, fmt.Errorf("language %q is not supported yet", req.Language)
	}
	return s.c.Evaluate(ctx, req)
}
