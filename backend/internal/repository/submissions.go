package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
)

type SubmissionsRepository struct {
	db *sql.DB
}

func NewSubmissionsRepository(db *sql.DB) *SubmissionsRepository {
	return &SubmissionsRepository{db: db}
}

func (r *SubmissionsRepository) Create(ctx context.Context, input model.NewSubmission) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var public, config string
	var runner string
	if err = tx.QueryRowContext(ctx, `SELECT json_object(
 'slug',slug,'title',title,'difficulty',difficulty,'category',category,
 'problemType',problem_type,'statement',statement_md,'constraints',constraints_md,'metadata',json(metadata),
 'templates',json((SELECT json_group_array(json_object('language',language,'starterCode',starter_code,'notes',notes)) FROM (SELECT * FROM problem_language_templates WHERE problem_id=$1 ORDER BY language))),
 'tags',json((SELECT json_group_array(tag) FROM (SELECT tag FROM problem_tags WHERE problem_id=$1 ORDER BY tag))),
 'assets',json((SELECT json_group_array(json_object('name',name,'type',asset_type,'mimeType',mime_type,'content',content_text)) FROM (SELECT * FROM problem_assets WHERE problem_id=$1 AND is_hidden=0 ORDER BY name)))
 ) FROM problems WHERE id=$1 AND is_published=1`, input.ProblemID).Scan(&public); err != nil {
		return "", err
	}
	if err = tx.QueryRowContext(ctx, `SELECT runner,config FROM problem_judge_configs WHERE problem_id=$1`, input.ProblemID).Scan(&runner, &config); err != nil {
		return "", err
	}
	snapshot := Snapshot{Problem: json.RawMessage(public), Config: model.JudgeConfig{Runner: runner}}
	if err = json.Unmarshal([]byte(config), &snapshot.Config.Config); err != nil {
		return "", err
	}
	rows, err := tx.QueryContext(ctx, `SELECT name,is_hidden,payload,weight,sort_order FROM problem_test_cases WHERE problem_id=$1 ORDER BY sort_order,name`, input.ProblemID)
	if err != nil {
		return "", err
	}
	for rows.Next() {
		var tc model.JudgeTestCase
		var payload string
		if err = rows.Scan(&tc.Name, &tc.IsHidden, &payload, &tc.Weight, &tc.SortOrder); err != nil {
			rows.Close()
			return "", err
		}
		if err = json.Unmarshal([]byte(payload), &tc.Payload); err != nil {
			rows.Close()
			return "", err
		}
		snapshot.Cases = append(snapshot.Cases, tc)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(append([]byte(input.ProblemID), data...))
	version := hex.EncodeToString(hash[:])
	if _, err = tx.ExecContext(ctx, `INSERT INTO problem_versions(id,problem_id,snapshot) VALUES($1,$2,$3) ON CONFLICT(id) DO NOTHING`, version, input.ProblemID, string(data)); err != nil {
		return "", err
	}
	var id string
	if err = tx.QueryRowContext(ctx, `INSERT INTO submissions(user_id,problem_id,language,mode,source_code,status,verdict,problem_version_id) VALUES($1,$2,$3,$4,$5,'queued','pending',$6) RETURNING id`, input.UserID, input.ProblemID, input.Language, input.Mode, input.SourceCode, version).Scan(&id); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO judge_jobs(submission_id) VALUES($1)`, id); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

func (r *SubmissionsRepository) MarkRunning(ctx context.Context, submissionID, token string) error {
	const query = `
		UPDATE submissions
		SET status = 'running', started_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND EXISTS(SELECT 1 FROM judge_jobs WHERE submission_id=$1 AND lease_token=$2)
	`
	res, err := r.db.ExecContext(ctx, query, submissionID, token)
	if err != nil {
		return fmt.Errorf("mark submission running: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected running submission: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SubmissionsRepository) Complete(ctx context.Context, submissionID, token string, result model.JudgeResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin complete submission tx: %w", err)
	}
	defer tx.Rollback()
	var owned bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM judge_jobs WHERE submission_id=$1 AND lease_token=$2)`, submissionID, token).Scan(&owned); err != nil {
		return err
	}
	if !owned {
		return fmt.Errorf("job lease lost")
	}

	const clearResults = `DELETE FROM submission_test_results WHERE submission_id = $1`
	if _, err := tx.ExecContext(ctx, clearResults, submissionID); err != nil {
		return fmt.Errorf("clear previous test results: %w", err)
	}

	const insertResult = `
		INSERT INTO submission_test_results (submission_id, test_case_id, case_name, is_hidden, status, message, execution_ms, sort_order)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8)
	`
	for _, item := range result.Results {
		if _, err := tx.ExecContext(
			ctx,
			insertResult,
			submissionID,
			item.TestCaseID,
			item.CaseName,
			item.IsHidden,
			item.Status,
			item.Message,
			item.ExecutionMS,
			item.SortOrder,
		); err != nil {
			return fmt.Errorf("insert submission test result: %w", err)
		}
	}

	status := result.Status
	if status == "" {
		status = "completed"
	}
	const updateSubmission = `
		UPDATE submissions
		SET status = $2,
			verdict = $3,
			score = $4,
			total_tests = $5,
			passed_tests = $6,
			compile_output = $7,
			runtime_output = $8,
			error_message = $9,
			completed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	if _, err := tx.ExecContext(
		ctx,
		updateSubmission,
		submissionID,
		status,
		result.Verdict,
		result.Score,
		result.TotalTests,
		result.PassedTests,
		result.CompileOutput,
		result.RuntimeOutput,
		result.ErrorMessage,
	); err != nil {
		return fmt.Errorf("update completed submission: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM judge_jobs WHERE submission_id=$1 AND lease_token=$2`, submissionID, token); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit complete submission tx: %w", err)
	}
	return nil
}

func (r *SubmissionsRepository) Fail(ctx context.Context, submissionID, message string) error {
	const query = `
		UPDATE submissions
		SET status = 'failed',
			verdict = 'system_error',
			error_message = $2,
			completed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	if _, err := r.db.ExecContext(ctx, query, submissionID, message); err != nil {
		return fmt.Errorf("fail submission: %w", err)
	}
	return nil
}

func (r *SubmissionsRepository) GetForRunner(ctx context.Context, submissionID string) (*model.Submission, error) {
	const query = `
		SELECT s.id, s.problem_id, p.slug, s.user_id, s.language, s.mode, s.source_code,
		       s.status, s.verdict, s.score, s.total_tests, s.passed_tests,
		       s.compile_output, s.runtime_output, s.error_message,
		       s.queued_at, s.started_at, s.completed_at, COALESCE(s.problem_version_id,'')
		FROM submissions s
		JOIN problems p ON p.id = s.problem_id
		WHERE s.id = $1
	`

	var submission model.Submission
	if err := r.db.QueryRowContext(ctx, query, submissionID).Scan(
		&submission.ID,
		&submission.ProblemID,
		&submission.ProblemSlug,
		&submission.UserID,
		&submission.Language,
		&submission.Mode,
		&submission.SourceCode,
		&submission.Status,
		&submission.Verdict,
		&submission.Score,
		&submission.TotalTests,
		&submission.PassedTests,
		&submission.CompileOutput,
		&submission.RuntimeOutput,
		&submission.ErrorMessage,
		&submission.QueuedAt,
		&submission.StartedAt,
		&submission.CompletedAt,
		&submission.ProblemVersionID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get submission for runner: %w", err)
	}
	return &submission, nil
}

func (r *SubmissionsRepository) GetByID(ctx context.Context, submissionID string) (*model.Submission, error) {
	const query = `
		SELECT s.id, s.problem_id, p.slug, s.user_id, s.language, s.mode,
		       s.status, s.verdict, s.score, s.total_tests, s.passed_tests,
		       s.compile_output, s.runtime_output, s.error_message,
		       s.queued_at, s.started_at, s.completed_at, COALESCE(s.problem_version_id,'')
		FROM submissions s
		JOIN problems p ON p.id = s.problem_id
		WHERE s.id = $1
	`

	var submission model.Submission
	if err := r.db.QueryRowContext(ctx, query, submissionID).Scan(
		&submission.ID,
		&submission.ProblemID,
		&submission.ProblemSlug,
		&submission.UserID,
		&submission.Language,
		&submission.Mode,
		&submission.Status,
		&submission.Verdict,
		&submission.Score,
		&submission.TotalTests,
		&submission.PassedTests,
		&submission.CompileOutput,
		&submission.RuntimeOutput,
		&submission.ErrorMessage,
		&submission.QueuedAt,
		&submission.StartedAt,
		&submission.CompletedAt,
		&submission.ProblemVersionID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get submission by id: %w", err)
	}

	results, err := r.getResults(ctx, submission.ID)
	if err != nil {
		return nil, err
	}
	submission.Results = results

	return &submission, nil
}

func (r *SubmissionsRepository) AcceptedProblemIDs(ctx context.Context, userID string) (map[string]bool, error) {
	const query = `
		SELECT DISTINCT problem_id
		FROM submissions
		WHERE user_id = $1
		  AND mode = 'submit'
		  AND status = 'completed'
		  AND verdict = 'accepted'
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query accepted problem ids: %w", err)
	}
	defer rows.Close()

	ids := make(map[string]bool)
	for rows.Next() {
		var problemID string
		if err := rows.Scan(&problemID); err != nil {
			return nil, fmt.Errorf("scan accepted problem id: %w", err)
		}
		ids[problemID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accepted problem ids: %w", err)
	}
	return ids, nil
}

func (r *SubmissionsRepository) getResults(ctx context.Context, submissionID string) ([]model.SubmissionTestResult, error) {
	const query = `
		SELECT COALESCE(test_case_id, ''), case_name, is_hidden, status, message, execution_ms, sort_order
		FROM submission_test_results
		WHERE submission_id = $1
		ORDER BY sort_order ASC, case_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, submissionID)
	if err != nil {
		return nil, fmt.Errorf("query submission results: %w", err)
	}
	defer rows.Close()

	items := make([]model.SubmissionTestResult, 0)
	for rows.Next() {
		var item model.SubmissionTestResult
		if err := rows.Scan(&item.TestCaseID, &item.CaseName, &item.IsHidden, &item.Status, &item.Message, &item.ExecutionMS, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("scan submission result: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate submission results: %w", err)
	}
	return items, nil
}

func (r *SubmissionsRepository) IsReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := r.db.QueryRowContext(ctx, `SELECT 1`).Scan(new(int)); err != nil {
		return fmt.Errorf("database healthcheck: %w", err)
	}
	return nil
}
