package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
)

// ImportGenerated is called only after the CLI executes validation. Neither
// reports nor publication flags supplied by an authoring agent are trusted.
func ImportGenerated(ctx context.Context, conn *sql.DB, b authoring.Bundle, report authoring.Report) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = importGeneratedTx(ctx, tx, b, report); err != nil {
		return err
	}
	return tx.Commit()
}

func importGeneratedTx(ctx context.Context, tx *sql.Tx, b authoring.Bundle, report authoring.Report) (string, error) {

	if report.BundleHash != authoring.Hash(b) || len(report.Checks) == 0 || len(report.Cases) == 0 || report.Image == "" {
		return "", fmt.Errorf("invalid validation report")
	}
	p := seedProblem{Slug: b.Slug, Title: b.Title, Difficulty: b.Difficulty, Category: b.Category, ProblemType: b.Format, Short: b.ShortDescription, Statement: b.Statement, Constraints: b.Constraints, Tags: b.Tags, Metadata: map[string]any{"format": b.Format, "targetMinutes": b.TargetMinutes, "objectives": b.Objectives, "seed": b.Seed, "bundleHash": report.BundleHash, "provenance": b.Provenance, "practice": b.Practice, "hints": b.Hints}, JudgeRunner: "stdio_v1", JudgeConfig: map[string]any{"comparator": b.Comparator, "sandbox_image": report.Image, "compile_timeout_ms": 20000, "run_timeout_ms": 3000, "explanation": b.Explanation}}
	for lang, starter := range b.Templates {
		p.Templates = append(p.Templates, seedTemplate{Language: lang, StarterCode: starter, Notes: "Standalone program: read stdin, write only the specified answer to stdout. Python NumPy is available."})
	}
	if b.Format != "coding" {
		p.JudgeRunner = "answer_v1"
		p.JudgeConfig["format"] = b.Format
		options := []map[string]string{}
		for _, o := range b.Options {
			p.JudgeConfig["explanation"] = fmt.Sprint(p.JudgeConfig["explanation"]) + "\n\n**" + o.ID + ".** " + o.Rationale
		}
		for _, o := range b.Options {
			options = append(options, map[string]string{"id": o.ID, "text": o.Text})
		}
		p.Metadata["options"] = options
	}
	for i, c := range report.Cases {
		sc := seedCase{Name: c.Name, Payload: map[string]any{"input": c.Input, "expected": c.Expected}, Hidden: c.Hidden, Weight: 1, SortOrder: i}
		if c.Hidden {
			p.HiddenCases = append(p.HiddenCases, sc)
		} else {
			sc.DisplayInput = c.Input
			sc.DisplayExpect = c.Expected
			p.VisibleCases = append(p.VisibleCases, sc)
		}
	}
	data, _ := json.Marshal(map[string]any{"bundle": b, "validation": report})
	p.Assets = []seedAsset{{AssetType: "authoring", Name: "private-authoring.json", MIMEType: "application/json", Content: string(data), Hidden: true}}
	// Prevent accidental replacement of an existing hand-authored problem.
	var existing string
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(json_extract(metadata,'$.bundleHash'),'') FROM problems WHERE slug=$1`, b.Slug).Scan(&existing)
	if err == nil && existing == "" {
		return "", fmt.Errorf("slug belongs to an existing hand-authored problem")
	}
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	id, err := upsertProblem(ctx, tx, p)
	if err != nil {
		return "", err
	}
	if err = replaceProblemRelations(ctx, tx, id, p); err != nil {
		return "", err
	}
	return id, nil
}

// ImportInterview publishes all questions and their order in one transaction.
// Validation happens before this function; every report is bound to its bundle.
func ImportInterview(ctx context.Context, conn *sql.DB, interview authoring.Interview, bundles []authoring.Bundle, reports []authoring.Report) error {
	if err := authoring.CheckInterviewMembers(interview, bundles); err != nil {
		return err
	}
	if len(bundles) != len(interview.Questions) || len(reports) != len(bundles) {
		return fmt.Errorf("interview validation is incomplete")
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	contract, _ := json.Marshal(interview.Contract)
	_, err = tx.ExecContext(ctx, `INSERT INTO interviews(slug,title,category,description,contract) VALUES($1,$2,$3,$4,$5) ON CONFLICT(slug) DO UPDATE SET title=excluded.title,category=excluded.category,description=excluded.description,contract=excluded.contract,updated_at=CURRENT_TIMESTAMP`, interview.Slug, interview.Title, interview.Category, interview.Description, string(contract))
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM interview_problems WHERE interview_slug=$1`, interview.Slug); err != nil {
		return err
	}
	for index, b := range bundles {
		id, err := importGeneratedTx(ctx, tx, b, reports[index])
		if err != nil {
			return fmt.Errorf("question %d: %w", index+1, err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO interview_problems(interview_slug,problem_id,position) VALUES($1,$2,$3)`, interview.Slug, id, index+1); err != nil {
			return err
		}
	}
	return tx.Commit()
}
