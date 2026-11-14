package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/sandbox"
)

// Comparator executes in the trusted worker, never in the candidate program.
type Comparator struct {
	Kind     string  `json:"kind"`
	Absolute float64 `json:"absolute,omitempty"`
	Relative float64 `json:"relative,omitempty"`
}

func (c Comparator) Validate() error {
	if c.Kind != "exact" && c.Kind != "numeric" && c.Kind != "json" {
		return fmt.Errorf("unsupported comparator %q", c.Kind)
	}
	if c.Absolute < 0 || c.Relative < 0 || math.IsNaN(c.Absolute) || math.IsNaN(c.Relative) || math.IsInf(c.Absolute, 0) || math.IsInf(c.Relative, 0) {
		return fmt.Errorf("invalid tolerance")
	}
	return nil
}
func (c Comparator) Equal(actual, expected string) bool {
	if c.Validate() != nil {
		return false
	}
	if c.Kind == "exact" {
		return strings.TrimSpace(actual) == strings.TrimSpace(expected)
	}
	if c.Kind == "numeric" {
		return c.number(strings.TrimSpace(actual), strings.TrimSpace(expected))
	}
	var a, b any
	da := json.NewDecoder(strings.NewReader(actual))
	da.UseNumber()
	db := json.NewDecoder(strings.NewReader(expected))
	db.UseNumber()
	if da.Decode(&a) != nil || db.Decode(&b) != nil || !json.Valid([]byte(actual)) || !json.Valid([]byte(expected)) {
		return false
	}
	return c.value(a, b)
}
func (c Comparator) number(a, b string) bool {
	// Reject non-finite numbers and unbounded exponents before rational parsing.
	if len(a) > 128 || len(b) > 128 {
		return false
	}
	for _, v := range []string{a, b} {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return false
		}
	}
	// Rational parsing avoids rounding large integers into the same float64.
	x, ok := new(big.Rat).SetString(a)
	if !ok {
		return false
	}
	y, ok := new(big.Rat).SetString(b)
	if !ok {
		return false
	}
	if c.Absolute == 0 && c.Relative == 0 {
		return x.Cmp(y) == 0
	}
	diff := new(big.Rat).Abs(new(big.Rat).Sub(x, y))
	abs := new(big.Rat).SetFloat64(c.Absolute)
	rel := new(big.Rat).Mul(new(big.Rat).SetFloat64(c.Relative), new(big.Rat).Abs(y))
	if rel.Cmp(abs) > 0 {
		abs = rel
	}
	return diff.Cmp(abs) <= 0
}
func (c Comparator) value(a, b any) bool {
	switch x := a.(type) {
	case json.Number:
		y, ok := b.(json.Number)
		return ok && c.number(string(x), string(y))
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !c.value(x[i], y[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			z, ok := y[k]
			if !ok || !c.value(v, z) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}
func comparator(config map[string]any) (Comparator, error) {
	b, err := json.Marshal(config["comparator"])
	if err != nil {
		return Comparator{}, err
	}
	var c Comparator
	if err = json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, c.Validate()
}

type ProgramRunner struct{ Executor sandbox.Executor }

func (r *ProgramRunner) Evaluate(ctx context.Context, req Request) (model.JudgeResult, error) {
	c, err := comparator(req.Config)
	if err != nil {
		return model.JudgeResult{}, err
	}
	executor := r.Executor
	if docker, ok := executor.(*sandbox.Docker); ok {
		copy := *docker
		if image, ok := req.Config["sandbox_image"].(string); ok && image != "" {
			copy.Image = image
		}
		executor = &copy
	}
	p, out, err := executor.Prepare(ctx, req.Language, req.SourceCode, req.CompileTimeout)
	if err != nil {
		return model.JudgeResult{}, err
	}
	if p == nil {
		return model.JudgeResult{Status: "completed", Verdict: "compile_error", CompileOutput: out.Stdout + out.Stderr, TotalTests: len(req.Cases), Results: defaultErrorResults(req.Cases, "compilation failed or exceeded limits")}, nil
	}
	defer p.Close()
	result := model.JudgeResult{Status: "completed", Verdict: "wrong_answer", TotalTests: len(req.Cases)}
	totalWeight, passedWeight := 0, 0
	for _, tc := range req.Cases {
		input, ok := payloadString(tc.Payload, "input")
		if !ok {
			return result, fmt.Errorf("case has no input")
		}
		expected, ok := payloadString(tc.Payload, "expected")
		if !ok {
			return result, fmt.Errorf("case has no expected output")
		}
		start := time.Now()
		got, err := p.Run(ctx, input, req.RunTimeout)
		if err != nil {
			return result, err
		}
		item := model.SubmissionTestResult{CaseName: tc.Name, IsHidden: tc.IsHidden, SortOrder: tc.SortOrder, ExecutionMS: int(time.Since(start).Milliseconds()), Status: "failed", Message: "output did not match"}
		weight := max(tc.Weight, 1)
		totalWeight += weight
		if got.ExitCode != 0 || got.TimedOut || got.Truncated {
			item.Status = "error"
			item.Message = "execution failed or exceeded limits"
		} else if c.Equal(got.Stdout, expected) {
			item.Status = "passed"
			item.Message = "ok"
			result.PassedTests++
			passedWeight += weight
		}
		if !tc.IsHidden {
			result.RuntimeOutput = truncateOutput(result.RuntimeOutput + "\n" + tc.Name + ":\n" + got.Stdout + got.Stderr)
		}
		result.Results = append(result.Results, item)
	}
	if totalWeight > 0 {
		result.Score = passedWeight * 100 / totalWeight
	}
	if result.PassedTests == result.TotalTests && result.TotalTests > 0 {
		result.Verdict = "accepted"
	}
	return result, nil
}

func GradeAnswer(source string, cases []model.JudgeTestCase, config map[string]any) (model.JudgeResult, error) {
	c, err := comparator(config)
	if err != nil {
		return model.JudgeResult{}, err
	}
	result := model.JudgeResult{Status: "completed", Verdict: "wrong_answer", TotalTests: len(cases)}
	for _, tc := range cases {
		expected, ok := payloadString(tc.Payload, "expected")
		if !ok {
			return result, fmt.Errorf("missing answer")
		}
		passed := c.Equal(source, expected)
		// Multiple select answers represent sets; JSON order has no significance.
		if config["format"] == "multiple_select" {
			var a, b []string
			passed = json.Unmarshal([]byte(source), &a) == nil && json.Unmarshal([]byte(expected), &b) == nil && sameSet(a, b)
		}
		status, message := "failed", "answer did not match"
		if passed {
			status, message = "passed", "ok"
			result.PassedTests++
		}
		result.Results = append(result.Results, model.SubmissionTestResult{CaseName: tc.Name, Status: status, Message: message, IsHidden: true})
	}
	if result.TotalTests > 0 {
		result.Score = 100 * result.PassedTests / result.TotalTests
		if result.PassedTests == result.TotalTests {
			result.Verdict = "accepted"
		}
	}
	return result, nil
}
func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := map[string]bool{}
	for _, v := range a {
		if set[v] {
			return false
		}
		set[v] = true
	}
	for _, v := range b {
		if !set[v] {
			return false
		}
	}
	return true
}
