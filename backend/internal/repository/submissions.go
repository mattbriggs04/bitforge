package repository

import (
	"context"
	"database/sql"
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
	const query = `
		INSERT INTO submissions (user_id, problem_id, language, mode, source_code, status, verdict)
		VALUES ($1, $2, $3, $4, $5, 'queued', 'pending')
		RETURNING id
	`
	var id string
	if err := r.db.QueryRowContext(ctx, query, input.UserID, input.ProblemID, input.Language, input.Mode, input.SourceCode).Scan(&id); err != nil {
		return "", fmt.Errorf("create submission: %w", err)
	}
	return id, nil
}

func (r *SubmissionsRepository) MarkRunning(ctx context.Context, submissionID string) error {
	const query = `
		UPDATE submissions
		SET status = 'running', started_at = now()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, submissionID)
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

func (r *SubmissionsRepository) Complete(ctx context.Context, submissionID string, result model.JudgeResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin complete submission tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const clearResults = `DELETE FROM submission_test_results WHERE submission_id = $1`
	if _, err := tx.ExecContext(ctx, clearResults, submissionID); err != nil {
		return fmt.Errorf("clear previous test results: %w", err)
	}

	const insertResult = `
		INSERT INTO submission_test_results (submission_id, test_case_id, case_name, is_hidden, status, message, execution_ms, sort_order)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7, $8)
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
			completed_at = now()
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
			completed_at = now()
		WHERE id = $1
	`
	if _, err := r.db.ExecContext(ctx, query, submissionID, message); err != nil {
		return fmt.Errorf("fail submission: %w", err)
	}
	return nil
}

func (r *SubmissionsRepository) GetForWorker(ctx context.Context, submissionID string) (*model.Submission, error) {
	const query = `
		SELECT s.id, s.problem_id, p.slug, s.user_id, s.language, s.mode, s.source_code,
		       s.status, s.verdict, s.score, s.total_tests, s.passed_tests,
		       s.compile_output, s.runtime_output, s.error_message,
		       s.queued_at, s.started_at, s.completed_at
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
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get submission for worker: %w", err)
	}
	return &submission, nil
}

func (r *SubmissionsRepository) GetByID(ctx context.Context, submissionID string) (*model.Submission, error) {
	const query = `
		SELECT s.id, s.problem_id, p.slug, s.user_id, s.language, s.mode,
		       s.status, s.verdict, s.score, s.total_tests, s.passed_tests,
		       s.compile_output, s.runtime_output, s.error_message,
		       s.queued_at, s.started_at, s.completed_at
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

func (r *SubmissionsRepository) getResults(ctx context.Context, submissionID string) ([]model.SubmissionTestResult, error) {
	const query = `
		SELECT COALESCE(test_case_id::text, ''), case_name, is_hidden, status, message, execution_ms, sort_order
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
		return fmt.Errorf("postgres healthcheck: %w", err)
	}
	return nil
}
