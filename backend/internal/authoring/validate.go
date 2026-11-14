package authoring

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/judge"
	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/sandbox"
)

type Report struct {
	BundleHash      string    `json:"bundleHash"`
	Image           string    `json:"image"`
	ValidatedAt     time.Time `json:"validatedAt"`
	Cases           []Case    `json:"cases"`
	MutantsRejected int       `json:"mutantsRejected"`
	Checks          []string  `json:"checks"`
}

func prepare(ctx context.Context, e sandbox.Executor, p Program) (sandbox.Program, error) {
	program, out, err := e.Prepare(ctx, p.Language, p.Source, 20*time.Second)
	if err != nil {
		return nil, err
	}
	if program == nil {
		return nil, fmt.Errorf("%s compilation failed: %s%s", p.Language, out.Stdout, out.Stderr)
	}
	return program, nil
}
func run(ctx context.Context, p sandbox.Program, input string) (string, error) {
	out, err := p.Run(ctx, input, 3*time.Second)
	if err != nil {
		return "", err
	}
	if out.ExitCode != 0 || out.TimedOut || out.Truncated {
		return "", fmt.Errorf("program failed or exceeded limits: %s", out.Stderr)
	}
	return strings.TrimSpace(out.Stdout), nil
}
func Validate(ctx context.Context, b Bundle, e sandbox.Executor, image string) (Report, error) {
	report := Report{BundleHash: Hash(b), Image: image, ValidatedAt: time.Now().UTC(), Checks: []string{}}
	if err := b.Check(); err != nil {
		return report, err
	}
	independent, err := prepare(ctx, e, b.Independent)
	if err != nil {
		return report, err
	}
	defer independent.Close()
	if b.Format != "coding" {
		got, err := run(ctx, independent, "")
		if err != nil {
			return report, err
		}
		grade, gradeErr := judge.GradeAnswer(got, []model.JudgeTestCase{{Payload: map[string]any{"expected": b.Answer}}}, map[string]any{"format": b.Format, "comparator": b.Comparator})
		if gradeErr != nil || grade.Verdict != "accepted" {
			return report, fmt.Errorf("independent verifier disagrees with answer")
		}
		// Choice rationale quality remains an author/reviewer responsibility.
		report.Checks = append(report.Checks, "schema and answer contract", "executable independent verifier")
		report.Cases = []Case{{Name: "answer", Expected: b.Answer, Hidden: true}}
		return report, nil
	}
	generator, err := prepare(ctx, e, *b.Generator)
	if err != nil {
		return report, err
	}
	defer generator.Close()
	input := fmt.Sprintf("%d\n", b.Seed)
	generated, err := run(ctx, generator, input)
	if err != nil {
		return report, err
	}
	again, err := run(ctx, generator, input)
	if err != nil || again != generated {
		return report, fmt.Errorf("generator must be deterministic for the supplied seed")
	}
	var extra []struct {
		Name  string `json:"name"`
		Input string `json:"input"`
	}
	if err = Decode([]byte(generated), &extra); err != nil || len(extra) < 20 || len(extra) > 100 {
		return report, fmt.Errorf("generator must emit a JSON array of 20..100 named input objects: %v", err)
	}
	report.Cases = append(report.Cases, b.Cases...)
	names := map[string]bool{}
	for _, c := range b.Cases {
		names[c.Name] = true
	}
	for _, c := range extra {
		if c.Name == "" || names[c.Name] || len(c.Input) > 16000 {
			return report, fmt.Errorf("invalid generated case %q", c.Name)
		}
		names[c.Name] = true
		report.Cases = append(report.Cases, Case{Name: c.Name, Input: c.Input, Hidden: true})
	}
	// Verify anchors against an independent implementation; derive all other outputs
	// from it before comparing every supported reference against these frozen cases.
	for i := range report.Cases {
		c := &report.Cases[i]
		expected, err := run(ctx, independent, c.Input)
		if err != nil {
			return report, fmt.Errorf("independent case %s: %w", c.Name, err)
		}
		if !b.Comparator.Equal(expected, expected) {
			return report, fmt.Errorf("invalid verifier output for %s", c.Name)
		}
		if i < len(b.Cases) && !b.Comparator.Equal(expected, c.Expected) {
			return report, fmt.Errorf("hand-derived anchor %s disagrees with verifier", c.Name)
		}
		c.Expected = expected
	}
	for _, ref := range b.References {
		p, err := prepare(ctx, e, ref)
		if err != nil {
			return report, err
		}
		for _, c := range report.Cases {
			got, e := run(ctx, p, c.Input)
			if e != nil || !b.Comparator.Equal(got, c.Expected) {
				p.Close()
				return report, fmt.Errorf("%s reference disagrees on %s: %v", ref.Language, c.Name, e)
			}
		}
		p.Close()
		// Starters must at least compile; malformed generated scaffolds never publish.
		p, err = prepare(ctx, e, Program{Language: ref.Language, Source: b.Templates[ref.Language]})
		if err != nil {
			return report, fmt.Errorf("starter: %w", err)
		}
		p.Close()
	}
	for _, mutant := range b.Mutants {
		p, err := prepare(ctx, e, mutant)
		if err != nil {
			return report, fmt.Errorf("mutants must compile: %w", err)
		}
		rejected := false
		for _, c := range report.Cases {
			got, e := run(ctx, p, c.Input)
			if e != nil {
				p.Close()
				return report, e
			}
			if !b.Comparator.Equal(got, c.Expected) {
				rejected = true
				break
			}
		}
		p.Close()
		if !rejected {
			return report, fmt.Errorf("test suite failed to reject a mutant")
		}
		report.MutantsRejected++
	}
	report.Checks = append(report.Checks, "schema and contract", "deterministic input generator", "hand-derived anchors", "independent oracle agreement across languages", "starter compilation", "known wrong solutions rejected")
	return report, nil
}

func (r Report) JSON() []byte { b, _ := json.MarshalIndent(r, "", "  "); return append(b, '\n') }
