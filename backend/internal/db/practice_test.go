package db_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/migrations"
	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

// Exercise the real pre-006 schema with history and generated/legacy content.
func TestPracticeMigrationPreservesGradingAndHistory(t *testing.T) {
	ctx := context.Background()
	conn, err := db.OpenDatabase(ctx, "file:"+filepath.Join(t.TempDir(), "migration.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err = conn.Exec(`CREATE TABLE schema_migrations(name TEXT PRIMARY KEY,applied_at TEXT DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001_init.sql", "004_durable_judge.sql", "005_interviews.sql"} {
		data, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = conn.Exec(string(data)); err != nil {
			t.Fatal(err)
		}
		if _, err = conn.Exec(`INSERT INTO schema_migrations(name) VALUES(?)`, name); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("SEED_PROBLEMS_DIR", "../../seed/problems")
	if err = db.SeedMVP(ctx, conn); err != nil {
		t.Fatal(err)
	}
	manifest, bundles, err := authoring.LoadInterview("../../../generation/interviews/binary-classifiers/interview.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bundles {
		report := authoring.Report{BundleHash: authoring.Hash(b), Image: "fixture", Checks: []string{"fixture"}, Cases: []authoring.Case{{Name: "private", Hidden: true, Expected: b.Answer}}}
		if err = db.ImportGenerated(ctx, conn, b, report); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = conn.Exec(`UPDATE problems SET metadata=json_remove(metadata,'$.practice','$.hints','$.targetMinutes') WHERE json_extract(metadata,'$.source')='seed'`); err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(`UPDATE problems SET metadata=json_remove(metadata,'$.practice','$.hints')`); err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(`INSERT INTO interviews(slug,title,category,description) VALUES(?,?,?,?)`, manifest.Slug, manifest.Title, manifest.Category, manifest.Description); err != nil {
		t.Fatal(err)
	}
	for i, b := range bundles {
		if _, err = conn.Exec(`INSERT INTO interview_problems SELECT ?,id,? FROM problems WHERE slug=?`, manifest.Slug, i+1, b.Slug); err != nil {
			t.Fatal(err)
		}
	}
	uid, err := repository.NewUsersRepository(conn).EnsureIdentity(ctx, "migration-user", "migration-user")
	if err != nil {
		t.Fatal(err)
	}
	var pid string
	if err = conn.QueryRow(`SELECT id FROM problems WHERE slug='can-control-frame'`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	submissions := repository.NewSubmissionsRepository(conn)
	sid, err := submissions.Create(ctx, model.NewSubmission{UserID: uid, ProblemID: pid, Language: "c", Mode: "submit", SourceCode: "draft"})
	if err != nil {
		t.Fatal(err)
	}
	before, err := submissions.LoadSnapshot(ctx, sid)
	if err != nil {
		t.Fatal(err)
	}
	var gradingBefore string
	if err = conn.QueryRow(`SELECT group_concat(payload,'|') FROM (SELECT payload FROM problem_test_cases ORDER BY id)`).Scan(&gradingBefore); err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, conn); err != nil {
		t.Fatal(err)
	}
	after, err := submissions.LoadSnapshot(ctx, sid)
	if err != nil || authoring.Hash(before) != authoring.Hash(after) {
		t.Fatal("historical snapshot changed", err)
	}
	var gradingAfter string
	if err = conn.QueryRow(`SELECT group_concat(payload,'|') FROM (SELECT payload FROM problem_test_cases ORDER BY id)`).Scan(&gradingAfter); err != nil || gradingAfter != gradingBefore {
		t.Fatal("grading payload changed", err)
	}
	repo := repository.NewProblemsRepository(conn)
	items, err := repo.ListPublished(ctx, model.ProblemFilter{})
	if err != nil || len(items) != 30 {
		t.Fatal(len(items), err)
	}
	for _, item := range items {
		p, err := repo.GetPublishedBySlug(ctx, item.Slug)
		if err != nil {
			t.Fatal(err)
		}
		if err = p.Practice.Check(); err != nil {
			t.Fatal(item.Slug, err)
		}
		if err = p.Hints.Check(); err != nil {
			t.Fatal(item.Slug, err)
		}
		if p.TargetMinutes < 1 || len(p.Objectives) == 0 {
			t.Fatal("missing learning metadata", item.Slug)
		}
		if strings.Contains(p.Hints.Nudge, "Identify the key input") {
			t.Fatal("checked-in question received fallback hints", item.Slug)
		}
	}
	interview, err := repo.GetInterview(ctx, manifest.Slug)
	if err != nil || interview.Contract.Preset != "custom" || interview.TargetMinutes != 45 {
		t.Fatal(interview, err)
	}
	// New seeding keeps the enriched content rather than reverting the migration.
	if err = db.SeedMVP(ctx, conn); err != nil {
		t.Fatal(err)
	}
	p, err := repo.GetPublishedBySlug(ctx, "can-control-frame")
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Hints.Check(); err != nil {
		t.Fatal("reseed lost hints", err)
	}
}
