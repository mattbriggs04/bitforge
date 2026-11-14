package judge

import (
	"context"
	"fmt"
	"github.com/mattbriggs04/bitforge/backend/internal/sandbox"
	"strings"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
)

type Request struct {
	SubmissionID   string
	Runner         string
	Language       string
	SourceCode     string
	Cases          []model.JudgeTestCase
	Config         map[string]any
	CompileTimeout time.Duration
	RunTimeout     time.Duration
}

type Engine interface {
	Evaluate(ctx context.Context, req Request) (model.JudgeResult, error)
}

type Service struct {
	c        Engine
	programs Engine
}

func NewService(c Engine) *Service {
	return &Service{c: c, programs: &ProgramRunner{Executor: sandbox.NewDocker()}}
}

func (s *Service) Evaluate(ctx context.Context, req Request) (model.JudgeResult, error) {
	switch req.Runner {
	case "answer_v1":
		return GradeAnswer(req.SourceCode, req.Cases, req.Config)
	case "stdio_v1":
		return s.programs.Evaluate(ctx, req)
	case "", "c_assert_harness_v1":
		if strings.ToLower(req.Language) != "c" {
			return model.JudgeResult{}, fmt.Errorf("legacy runner only supports C")
		}
		return s.c.Evaluate(ctx, req)
	default:
		return model.JudgeResult{}, fmt.Errorf("unsupported runner %q", req.Runner)
	}
}
