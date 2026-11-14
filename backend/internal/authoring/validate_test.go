package authoring

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/sandbox"
)

func example(t *testing.T, name string) Bundle {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../../../generation/examples", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var b Bundle
	if err = Decode(data, &b); err != nil {
		t.Fatal(err)
	}
	return b
}
func TestExamplesAndStrictDecoding(t *testing.T) {
	for _, name := range []string{"stable-log-sum-exp", "adc-quantization-error", "pthread-counter-interleaving"} {
		b := example(t, name)
		if err := b.Check(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	var b Bundle
	if Decode([]byte(`{"schemaVersion":1,"published":true}`), &b) == nil {
		t.Fatal("unrecognized publication bypass accepted")
	}
}
func TestPromptSelectionIsRepeatable(t *testing.T) {
	c := Config{SchemaVersion: 1, Seed: 41, Category: "random", Format: "random", Language: "random", Difficulty: "random", TargetMinutes: 25}
	a, err := c.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Resolve()
	if err != nil || Hash(a) != Hash(b) {
		t.Fatal(a, b, err)
	}
	c.Category = "unknown"
	if _, err = c.Resolve(); err == nil {
		t.Fatal("invalid category accepted")
	}
}

type pipelineExecutor struct {
	survivingMutant bool
	badOracle       bool
}
type pipelineProgram struct {
	source   string
	executor pipelineExecutor
}

func (e pipelineExecutor) Prepare(_ context.Context, _ string, source string, _ time.Duration) (sandbox.Program, sandbox.Output, error) {
	return pipelineProgram{source, e}, sandbox.Output{}, nil
}
func (p pipelineProgram) Close() {}
func (p pipelineProgram) Run(_ context.Context, input string, _ time.Duration) (sandbox.Output, error) {
	if p.source == "generator" {
		var cases []map[string]string
		for i := 0; i < 20; i++ {
			cases = append(cases, map[string]string{"name": string(rune('a' + i)), "input": "1"})
		}
		data, _ := json.Marshal(cases)
		return sandbox.Output{Stdout: string(data)}, nil
	}
	if p.source == "answer-set" {
		return sandbox.Output{Stdout: `["B","A"]`}, nil
	}
	output := strings.TrimSpace(input)
	if p.source == "independent" && p.executor.badOracle {
		output = "99"
	}
	if strings.HasPrefix(p.source, "mutant") && !p.executor.survivingMutant {
		output = "wrong"
	}
	return sandbox.Output{Stdout: output}, nil
}
func TestValidationRejectsOracleDisagreementAndSurvivingMutants(t *testing.T) {
	b := example(t, "stable-log-sum-exp")
	b.Comparator.Kind = "exact"
	b.Comparator.Absolute = 0
	b.Comparator.Relative = 0
	b.Cases = []Case{{Name: "anchor-visible", Input: "1", Expected: "1"}, {Name: "anchor-hidden", Input: "2", Expected: "2", Hidden: true}}
	b.References = []Program{{Language: "python", Source: "reference"}}
	b.Templates = map[string]string{"python": "starter"}
	b.Independent = Program{Language: "python", Source: "independent"}
	b.Generator = &Program{Language: "python", Source: "generator"}
	b.Mutants = []Program{{Language: "python", Source: "mutant-one"}, {Language: "python", Source: "mutant-two"}}
	report, err := Validate(context.Background(), b, pipelineExecutor{}, "test-image")
	if err != nil || len(report.Cases) != 22 || report.MutantsRejected != 2 {
		t.Fatal(report, err)
	}
	if _, err = Validate(context.Background(), b, pipelineExecutor{badOracle: true}, "test-image"); err == nil || !strings.Contains(err.Error(), "anchor") {
		t.Fatalf("bad oracle accepted: %v", err)
	}
	if _, err = Validate(context.Background(), b, pipelineExecutor{survivingMutant: true}, "test-image"); err == nil || !strings.Contains(err.Error(), "mutant") {
		t.Fatalf("surviving mutant accepted: %v", err)
	}
}

func TestMultipleSelectVerifierMayReturnDifferentOrder(t *testing.T) {
	b := example(t, "pthread-counter-interleaving")
	b.Format = "multiple_select"
	b.Answer = `["A","B"]`
	b.Independent = Program{Language: "python", Source: "answer-set"}
	if _, err := Validate(context.Background(), b, pipelineExecutor{}, "test-image"); err != nil {
		t.Fatal(err)
	}
}
