package repository_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/judge"
	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

func TestDurableJobsFreezeCasesAndFenceStaleWorkers(t *testing.T) {
	ctx := context.Background()
	conn, err := db.OpenDatabase(ctx, "file:"+filepath.Join(t.TempDir(), "jobs.db")+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err = db.RunMigrations(ctx, conn); err != nil {
		t.Fatal(err)
	}
	// Exercise the additive migrations twice, including an existing schema.
	if err = db.RunMigrations(ctx, conn); err != nil {
		t.Fatal(err)
	}
	b := authoring.Bundle{Slug: "test-number", Title: "Number", Category: "embedded", Difficulty: "medium", Format: "numeric", Comparator: judge.Comparator{Kind: "numeric"}, Answer: "12", Statement: "How many?", Explanation: "Private answer explanation"}
	report := authoring.Report{BundleHash: authoring.Hash(b), Image: "sha256:test", Checks: []string{"test"}, Cases: []authoring.Case{{Name: "answer", Expected: "12", Hidden: true}}}
	if err = db.ImportGenerated(ctx, conn, b, report); err != nil {
		t.Fatal(err)
	}
	var pid string
	if err = conn.QueryRow(`SELECT id FROM problems WHERE slug='test-number'`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	userID, err := repository.NewUsersRepository(conn).EnsureIdentity(ctx, "test-user", "tester")
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewSubmissionsRepository(conn)
	id, err := repo.Create(ctx, model.NewSubmission{UserID: userID, ProblemID: pid, Language: "answer", Mode: "submit", SourceCode: "12"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(`UPDATE problem_test_cases SET payload='{"input":"","expected":"99"}' WHERE problem_id=?`, pid); err != nil {
		t.Fatal(err)
	}
	snapshot, err := repo.LoadSnapshot(ctx, id)
	if err != nil || snapshot.Cases[0].Payload["expected"] != "12" {
		t.Fatalf("snapshot changed: %+v %v", snapshot, err)
	}
	first, err := repo.Claim(ctx)
	if err != nil || first == nil {
		t.Fatal(first, err)
	}
	if next, err := repo.Claim(ctx); err != nil || next != nil {
		t.Fatalf("duplicate claim %+v %v", next, err)
	}
	if err = repo.MarkRunning(ctx, id, first.Token); err != nil {
		t.Fatal(err)
	}
	// Simulate a worker crash and claim the expired lease with a fresh repository.
	if _, err = conn.Exec(`UPDATE judge_jobs SET lease_until=0 WHERE submission_id=?`, id); err != nil {
		t.Fatal(err)
	}
	second, err := repository.NewSubmissionsRepository(conn).Claim(ctx)
	if err != nil || second == nil || second.Token == first.Token || second.Attempts != 2 {
		t.Fatal(second, err)
	}
	result := model.JudgeResult{Status: "completed", Verdict: "accepted", TotalTests: 1, PassedTests: 1, Score: 100, Results: []model.SubmissionTestResult{{CaseName: "answer", IsHidden: true, Status: "passed", Message: "ok"}}}
	if err = repo.Complete(ctx, id, first.Token, result); err == nil || !strings.Contains(err.Error(), "lease lost") {
		t.Fatalf("stale worker completed: %v", err)
	}
	if err = repo.Complete(ctx, id, second.Token, result); err != nil {
		t.Fatal(err)
	}
	attempt, err := repo.GetByID(ctx, id)
	if err != nil || attempt.Verdict != "accepted" || attempt.ProblemVersionID == "" {
		t.Fatal(attempt, err)
	}
	if next, err := repo.Claim(ctx); err != nil || next != nil {
		t.Fatal(next, err)
	}
	// Only public option text may leave the backend.
	detail, err := repository.NewProblemsRepository(conn).GetPublishedBySlug(ctx, "test-number")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := detail.Metadata["explanation"]; ok || len(detail.Assets) != 0 {
		t.Fatal("private authoring assets exposed")
	}
}

func TestConcurrentMigrationAndJobClaims(t *testing.T) {
	ctx := context.Background()
	url := "file:" + filepath.Join(t.TempDir(), "concurrent.db") + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	one, err := db.OpenDatabase(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer one.Close()
	two, err := db.OpenDatabase(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer two.Close()
	errors := make(chan error, 2)
	var wg sync.WaitGroup
	for _, conn := range []*sql.DB{one, two} {
		wg.Add(1)
		go func(conn *sql.DB) { defer wg.Done(); errors <- db.RunMigrations(ctx, conn) }(conn)
	}
	wg.Wait()
	for range 2 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
	if _, err = one.Exec(`INSERT INTO problems(id,slug,title,difficulty,category,problem_type,short_description,statement_md) VALUES('p','counter','Counter','easy','embedded','numeric','','Question'); INSERT INTO problem_judge_configs(problem_id,runner,config) VALUES('p','answer_v1','{"comparator":{"kind":"numeric"}}'); INSERT INTO problem_test_cases(problem_id,name,payload,is_hidden) VALUES('p','answer','{"expected":"1"}',1)`); err != nil {
		t.Fatal(err)
	}
	user, err := repository.NewUsersRepository(one).EnsureIdentity(ctx, "concurrent-key", "concurrent-student")
	if err != nil {
		t.Fatal(err)
	}
	for range 8 {
		if _, err = repository.NewSubmissionsRepository(one).Create(ctx, model.NewSubmission{UserID: user, ProblemID: "p", Language: "answer", Mode: "submit", SourceCode: "1"}); err != nil {
			t.Fatal(err)
		}
	}
	ids := make(chan string, 8)
	for _, conn := range []*sql.DB{one, two} {
		wg.Add(1)
		go func(conn *sql.DB) {
			defer wg.Done()
			repo := repository.NewSubmissionsRepository(conn)
			for {
				job, err := repo.Claim(ctx)
				if err != nil {
					errors <- err
					return
				}
				if job == nil {
					errors <- nil
					return
				}
				ids <- job.SubmissionID
			}
		}(conn)
	}
	wg.Wait()
	close(ids)
	for range 2 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate claim %s", id)
		}
		seen[id] = true
	}
	if len(seen) != 8 {
		t.Fatalf("claimed %d jobs", len(seen))
	}
}
