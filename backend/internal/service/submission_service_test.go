package service_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/judge"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
	"github.com/mattbriggs04/bitforge/backend/internal/service"
)

func TestAnswerWorkflowAndPrivateExplanation(t *testing.T) {
	ctx := context.Background()
	conn, err := db.OpenDatabase(ctx, "file:"+filepath.Join(t.TempDir(), "practice.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err = db.RunMigrations(ctx, conn); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../../generation/examples/pthread-counter-interleaving.json")
	if err != nil {
		t.Fatal(err)
	}
	var b authoring.Bundle
	if err = authoring.Decode(data, &b); err != nil {
		t.Fatal(err)
	}
	report := authoring.Report{BundleHash: authoring.Hash(b), Image: "test-image", Checks: []string{"fixture"}, Cases: []authoring.Case{{Name: "answer", Expected: b.Answer, Hidden: true}}}
	if err = db.ImportGenerated(ctx, conn, b, report); err != nil {
		t.Fatal(err)
	}
	problems := repository.NewProblemsRepository(conn)
	submissions := repository.NewSubmissionsRepository(conn)
	users := repository.NewUsersRepository(conn)
	svc := service.NewSubmissionService(problems, submissions, users, "demo")
	detail, err := problems.GetPublishedBySlug(ctx, b.Slug)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Samples) != 0 || len(detail.Assets) != 0 || strings.Contains(stringMustJSON(t, detail), "rationale") {
		t.Fatal("private material exposed")
	}
	attempt, err := svc.Create(ctx, service.CreateSubmissionInput{ProblemSlug: b.Slug, Language: "answer", Mode: "submit", SourceCode: `"B"`, UserHandle: "student", UserKey: "student-key"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Explanation(ctx, b.Slug, attempt.SubmissionID, "student-key", "student"); err == nil {
		t.Fatal("explanation revealed before attempt completed")
	}
	job, err := submissions.Claim(ctx)
	if err != nil || job == nil {
		t.Fatal(job, err)
	}
	runner := service.NewSubmissionRunnerService(submissions, problems, judge.NewService(judge.NewCAssertRunner()), 0, 0)
	if err = runner.ProcessSubmission(ctx, job.SubmissionID, job.Token); err != nil {
		t.Fatal(err)
	}
	result, err := svc.GetByID(ctx, attempt.SubmissionID)
	if err != nil || result.Verdict != "accepted" {
		t.Fatal(result, err)
	}
	text, err := svc.Explanation(ctx, b.Slug, attempt.SubmissionID, "student-key", "student")
	if err != nil || !strings.Contains(text, "lost-update") {
		t.Fatal(text, err)
	}
	if _, err = svc.Explanation(ctx, b.Slug, attempt.SubmissionID, "other-key", "other-student"); err == nil {
		t.Fatal("other user got private explanation")
	}
}

func stringMustJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
