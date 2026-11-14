package judge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/sandbox"
)

type CAssertRunner struct{ Executor sandbox.Executor }

func NewCAssertRunner() *CAssertRunner {
	return &CAssertRunner{Executor: sandbox.NewDocker()}
}

func (r *CAssertRunner) Evaluate(ctx context.Context, req Request) (model.JudgeResult, error) {
	if len(req.Cases) == 0 {
		return model.JudgeResult{}, fmt.Errorf("no test cases configured for submission")
	}

	harness, err := buildHarness(req.SourceCode, req.Cases, req.Config)
	if err != nil {
		return model.JudgeResult{}, fmt.Errorf("build harness: %w", err)
	}

	p, compile, err := r.Executor.Prepare(ctx, "c", harness, req.CompileTimeout)
	if err != nil {
		return model.JudgeResult{}, err
	}
	if p == nil {
		return model.JudgeResult{Status: "completed", Verdict: "compile_error", CompileOutput: truncateOutput(compile.Stdout + compile.Stderr), TotalTests: len(req.Cases), Results: defaultErrorResults(req.Cases, "compilation failed or exceeded limits")}, nil
	}
	defer p.Close()
	out, err := p.Run(ctx, "", req.RunTimeout)
	if err != nil {
		return model.JudgeResult{}, err
	}
	runOutput := out.Stdout + out.Stderr
	formattedRuntimeOutput := truncateOutput(formatRuntimeOutput(runOutput))
	var runErr error
	if out.ExitCode != 0 {
		runErr = fmt.Errorf("exit %d", out.ExitCode)
	}
	if out.TimedOut || out.Truncated {
		return model.JudgeResult{Status: "completed", Verdict: "runtime_error", TotalTests: len(req.Cases), Results: defaultErrorResults(req.Cases, "execution exceeded limits")}, nil
	}

	parsed := parseHarnessOutput(runOutput, req.Cases)
	passed := 0
	for i := range parsed {
		if parsed[i].Status == "passed" {
			passed++
		}
	}

	total := len(req.Cases)
	verdict := "wrong_answer"
	errMessage := ""

	if parsedCount(parsed) < total {
		verdict = "runtime_error"
		errMessage = "program exited before all tests completed"
		for i := range parsed {
			if parsed[i].Status == "skipped" {
				parsed[i].Status = "error"
				parsed[i].Message = "execution interrupted"
			}
		}
	} else if passed == total {
		if runErr != nil {
			verdict = "runtime_error"
			errMessage = "program terminated unexpectedly"
		} else {
			verdict = "accepted"
		}
	} else {
		verdict = "wrong_answer"
	}

	for _, tc := range req.Cases {
		if tc.IsHidden {
			formattedRuntimeOutput = "Runtime output is available in Run Samples mode."
			break
		}
	}
	score := 0
	if total > 0 {
		score = int((float64(passed) / float64(total)) * 100)
	}

	return model.JudgeResult{
		Status:        "completed",
		Verdict:       verdict,
		Score:         score,
		TotalTests:    total,
		PassedTests:   passed,
		RuntimeOutput: formattedRuntimeOutput,
		ErrorMessage:  errMessage,
		Results:       parsed,
	}, nil
}

func buildHarness(userSource string, cases []model.JudgeTestCase, cfg map[string]any) (string, error) {
	prelude := configString(cfg, "prelude", "")
	builder := strings.Builder{}
	builder.WriteString("#include <stdbool.h>\n")
	builder.WriteString("#include <stddef.h>\n")
	builder.WriteString("#include <stdint.h>\n")
	builder.WriteString("#include <stdio.h>\n")
	builder.WriteString("#include <string.h>\n")
	builder.WriteString("#include <stdlib.h>\n")
	builder.WriteString("\n")
	if prelude != "" {
		builder.WriteString(prelude)
		if !strings.HasSuffix(prelude, "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("#line 1 \"solution.c\"\n")
	builder.WriteString(userSource)
	if !strings.HasSuffix(userSource, "\n") {
		builder.WriteString("\n")
	}
	builder.WriteString("\n#line 1 \"harness.c\"\n\n")
	builder.WriteString("int main(void) {\n")
	builder.WriteString("    int __passed = 0;\n")
	builder.WriteString("\n")

	for i, testCase := range cases {
		snippet, ok := payloadString(testCase.Payload, "code")
		if !ok || strings.TrimSpace(snippet) == "" {
			return "", fmt.Errorf("test case %q has no payload.code", testCase.Name)
		}
		safeName := sanitizeCaseName(testCase.Name)
		builder.WriteString("    {\n")
		builder.WriteString("        int case_passed = 0;\n")
		builder.WriteString(indentSnippet(snippet, 8))
		if !strings.HasSuffix(snippet, "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("        if (case_passed) {\n")
		builder.WriteString(fmt.Sprintf("            printf(\"CASE|%d|%s|PASS\\n\");\n", i, safeName))
		builder.WriteString("            __passed++;\n")
		builder.WriteString("        } else {\n")
		builder.WriteString(fmt.Sprintf("            printf(\"CASE|%d|%s|FAIL\\n\");\n", i, safeName))
		builder.WriteString("        }\n")
		builder.WriteString("    }\n\n")
	}

	builder.WriteString(fmt.Sprintf("    printf(\"SUMMARY|%%d|%d\\n\", __passed);\n", len(cases)))
	builder.WriteString(fmt.Sprintf("    return __passed == %d ? 0 : 1;\n", len(cases)))
	builder.WriteString("}\n")
	return builder.String(), nil
}

var harnessLineRegex = regexp.MustCompile(`^CASE\|(\d+)\|([^|]+)\|(PASS|FAIL)$`)
var harnessSummaryLineRegex = regexp.MustCompile(`^SUMMARY\|(\d+)\|(\d+)$`)
var submissionPathRegex = regexp.MustCompile(`(?:[A-Za-z]:)?[/\\][^:\n]*[/\\]submission\.c`)

func parseHarnessOutput(output string, cases []model.JudgeTestCase) []model.SubmissionTestResult {
	results := make([]model.SubmissionTestResult, len(cases))
	for i, testCase := range cases {
		results[i] = model.SubmissionTestResult{
			CaseName:    testCase.Name,
			IsHidden:    testCase.IsHidden,
			Status:      "skipped",
			Message:     "case did not run",
			ExecutionMS: 0,
			SortOrder:   testCase.SortOrder,
			TestCaseID:  testCase.ID,
		}
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := harnessLineRegex.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}
		idx, err := strconv.Atoi(matches[1])
		if err != nil || idx < 0 || idx >= len(results) {
			continue
		}
		status := "failed"
		msg := "assertion failed"
		if matches[3] == "PASS" {
			status = "passed"
			msg = "ok"
		}
		results[idx].Status = status
		results[idx].Message = msg
		results[idx].CaseName = matches[2]
	}

	return results
}

func defaultErrorResults(cases []model.JudgeTestCase, message string) []model.SubmissionTestResult {
	results := make([]model.SubmissionTestResult, 0, len(cases))
	for _, testCase := range cases {
		results = append(results, model.SubmissionTestResult{
			CaseName:   testCase.Name,
			IsHidden:   testCase.IsHidden,
			Status:     "error",
			Message:    message,
			SortOrder:  testCase.SortOrder,
			TestCaseID: testCase.ID,
		})
	}
	return results
}

func parsedCount(items []model.SubmissionTestResult) int {
	count := 0
	for _, item := range items {
		if item.Status == "passed" || item.Status == "failed" {
			count++
		}
	}
	return count
}

func payloadString(payload map[string]any, key string) (string, bool) {
	value, ok := payload[key]
	if !ok {
		return "", false
	}
	asString, ok := value.(string)
	return asString, ok
}

func indentSnippet(snippet string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(snippet, "\n")
	for i := range lines {
		if strings.TrimSpace(lines[i]) == "" {
			lines[i] = ""
			continue
		}
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func sanitizeCaseName(input string) string {
	input = strings.ReplaceAll(input, "|", "_")
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.TrimSpace(input)
	if input == "" {
		return "unnamed_case"
	}
	return input
}

func formatCompileOutput(output, sourcePath, workingDir string) string {
	formatted := output
	if sourcePath != "" {
		formatted = strings.ReplaceAll(formatted, sourcePath, "solution.c")
	}
	if workingDir != "" {
		prefix := filepath.Clean(workingDir) + string(os.PathSeparator)
		formatted = strings.ReplaceAll(formatted, prefix, "")
	}
	// Fallback for any remaining absolute path to generated source.
	formatted = submissionPathRegex.ReplaceAllString(formatted, "solution.c")
	return formatted
}

func formatRuntimeOutput(output string) string {
	lines := strings.Split(output, "\n")
	formatted := make([]string, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if matches := harnessLineRegex.FindStringSubmatch(line); len(matches) == 4 {
			formatted = append(formatted, fmt.Sprintf("CASE %s | %s | %s", matches[1], matches[2], matches[3]))
			continue
		}
		if matches := harnessSummaryLineRegex.FindStringSubmatch(line); len(matches) == 3 {
			formatted = append(formatted, fmt.Sprintf("SUMMARY: %s/%s", matches[1], matches[2]))
			continue
		}

		formatted = append(formatted, line)
	}
	return strings.Join(formatted, "\n")
}

func truncateOutput(output string) string {
	const maxBytes = 6000
	if len(output) <= maxBytes {
		return output
	}
	return output[:maxBytes] + "\n... output truncated ..."
}

func configString(cfg map[string]any, key, fallback string) string {
	if cfg == nil {
		return fallback
	}
	value, ok := cfg[key]
	if !ok {
		return fallback
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return fallback
	}
	return text
}

func configStringSlice(cfg map[string]any, key string) []string {
	if cfg == nil {
		return nil
	}
	value, ok := cfg[key]
	if !ok {
		return nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok || text == "" {
			continue
		}
		out = append(out, text)
	}
	return out
}
