package repository_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

func TestInterviewImportOrderPrivacyAndRollback(t *testing.T) {
	ctx := context.Background()
	conn, err := db.OpenDatabase(ctx, "file:"+filepath.Join(t.TempDir(), "interview.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err = db.RunMigrations(ctx, conn); err != nil {
		t.Fatal(err)
	}
	manifest, bundles, err := authoring.LoadInterview("../../../generation/interviews/binary-classifiers/interview.json")
	if err != nil {
		t.Fatal(err)
	}
	// Persistence tests use fixture reports. Actual executables are validated by
	// import-interview before invoking this database boundary.
	reports := make([]authoring.Report, len(bundles))
	for i, b := range bundles {
		reports[i] = authoring.Report{BundleHash: authoring.Hash(b), Image: "test-image", Checks: []string{"fixture"}, Cases: []authoring.Case{{Name: "answer", Expected: b.Answer, Hidden: true}}}
	}
	bad := append([]authoring.Report(nil), reports...)
	bad[len(bad)-1].BundleHash = "wrong"
	if err = db.ImportInterview(ctx, conn, manifest, bundles, bad); err == nil {
		t.Fatal("invalid report accepted")
	}
	for _, table := range []string{"problems", "interviews", "interview_problems"} {
		var n int
		if err = conn.QueryRow(`SELECT count(*) FROM ` + table).Scan(&n); err != nil || n != 0 {
			t.Fatalf("partial publication in %s: %d %v", table, n, err)
		}
	}
	if err = db.ImportInterview(ctx, conn, manifest, bundles, reports); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewProblemsRepository(conn)
	detail, err := repo.GetInterview(ctx, manifest.Slug)
	if err != nil || detail == nil {
		t.Fatal(detail, err)
	}
	if len(detail.Problems) != 4 || detail.QuestionCount != 4 || detail.TargetMinutes != 45 {
		t.Fatalf("unexpected summary: %+v", detail)
	}
	for i, p := range detail.Problems {
		if p.Slug != bundles[i].Slug {
			t.Fatal("question order lost")
		}
	}
	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"rationale", "independent", "expected", "references", "reviewNotes"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("private field %s exposed", secret)
		}
	}
	question, err := repo.GetPublishedBySlug(ctx, bundles[0].Slug)
	if err != nil || len(question.Interviews) != 1 || question.Interviews[0].Slug != manifest.Slug {
		t.Fatal(question, err)
	}
	// A failing update must preserve both old member content and collection order.
	revised := append([]authoring.Bundle(nil), bundles...)
	revised[0].Title = "Changed draft"
	revisedReports := append([]authoring.Report(nil), reports...)
	revisedReports[0].BundleHash = authoring.Hash(revised[0])
	revisedReports[3].BundleHash = "bad"
	if err = db.ImportInterview(ctx, conn, manifest, revised, revisedReports); err == nil {
		t.Fatal("bad update accepted")
	}
	unchanged, err := repo.GetPublishedBySlug(ctx, bundles[0].Slug)
	if err != nil || unchanged.Title != bundles[0].Title {
		t.Fatal("failed import changed a question", err)
	}
	// Publishing the same group again replaces memberships without duplicates.
	if err = db.ImportInterview(ctx, conn, manifest, bundles, reports); err != nil {
		t.Fatal(err)
	}
	list, err := repo.ListInterviews(ctx, "")
	if err != nil || len(list) != 1 || list[0].QuestionCount != 4 {
		t.Fatal(list, err)
	}
	// An unpublished child hides the collection, rather than renumbering it.
	if _, err = conn.Exec(`UPDATE problems SET is_published=0 WHERE slug=?`, bundles[1].Slug); err != nil {
		t.Fatal(err)
	}
	hidden, err := repo.GetInterview(ctx, manifest.Slug)
	if err != nil || hidden != nil {
		t.Fatal(hidden, err)
	}
	list, err = repo.ListInterviews(ctx, "")
	if err != nil || len(list) != 0 {
		t.Fatal(list, err)
	}
	if missing, err := repo.GetInterview(ctx, "not-an-interview"); err != nil || missing != nil {
		t.Fatal(missing, err)
	}
}
