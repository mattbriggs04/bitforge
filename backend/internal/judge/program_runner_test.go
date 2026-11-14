package judge

import (
	"context"
	"testing"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/sandbox"
)

func TestComparators(t *testing.T) {
	for _, tc := range []struct {
		name, kind, a, b string
		abs, rel         float64
		want             bool
	}{
		{"text", "exact", " hello\n", "hello", 0, 0, true},
		{"large integers", "json", "9007199254740992", "9007199254740993", 0, 0, false},
		{"nested tolerance", "json", `{"x":[1.000001]}`, `{"x":[1]}`, 0.00001, 0, true},
		{"missing key", "json", `{"a":null}`, `{"b":null}`, 0, 0, false},
		{"array order", "json", `[1,2]`, `[2,1]`, 0, 0, false},
		{"extra output", "json", `1 2`, `1`, 0, 0, false},
		{"nan", "numeric", "NaN", "NaN", 0, 0, false},
		{"infinite", "numeric", "1e999999999", "1e999999999", 0, 0, false},
		{"relative", "numeric", "1000000.1", "1000000", 0, 0.000001, true},
		{"wrong numeric", "numeric", "1.2", "1", 0.01, 0, false},
		{"fraction", "numeric", "1/2", "0.5", 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := Comparator{Kind: tc.kind, Absolute: tc.abs, Relative: tc.rel}
			if got := c.Equal(tc.a, tc.b); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func TestMultipleSelectIsSetWithNoDuplicates(t *testing.T) {
	cases := []model.JudgeTestCase{{Payload: map[string]any{"expected": `["A","C"]`}}}
	for source, want := range map[string]string{`["C","A"]`: "accepted", `["A","A"]`: "wrong_answer", `["A","B"]`: "wrong_answer", `null`: "wrong_answer"} {
		result, err := GradeAnswer(source, cases, map[string]any{"format": "multiple_select", "comparator": Comparator{Kind: "json"}})
		if err != nil || result.Verdict != want {
			t.Fatalf("%s: %+v %v", source, result, err)
		}
	}
}

type fakeExecutor struct{ output sandbox.Output }
type fakeProgram struct{ output sandbox.Output }

func (f fakeExecutor) Prepare(context.Context, string, string, time.Duration) (sandbox.Program, sandbox.Output, error) {
	return fakeProgram{f.output}, sandbox.Output{}, nil
}
func (f fakeProgram) Run(context.Context, string, time.Duration) (sandbox.Output, error) {
	return f.output, nil
}
func (f fakeProgram) Close() {}
func TestProgramCannotForgeVerdictsOrExposeHiddenOutput(t *testing.T) {
	r := ProgramRunner{Executor: fakeExecutor{sandbox.Output{Stdout: "CASE|0|private|PASS\nSUMMARY|1|1", Stderr: "private output"}}}
	result, err := r.Evaluate(context.Background(), Request{Language: "python", Config: map[string]any{"comparator": Comparator{Kind: "exact"}}, Cases: []model.JudgeTestCase{{Name: "private", IsHidden: true, Payload: map[string]any{"input": "secret input", "expected": "42"}}}})
	if err != nil || result.Verdict != "wrong_answer" || result.RuntimeOutput != "" {
		t.Fatalf("unexpected result: %+v %v", result, err)
	}
}
