package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
)

type Snapshot struct {
	Problem json.RawMessage       `json:"problem"`
	Config  model.JudgeConfig     `json:"judge"`
	Cases   []model.JudgeTestCase `json:"cases"`
}
type Job struct {
	SubmissionID, Token string
	Attempts            int
}

// UPDATE RETURNING atomically claims one job across worker processes. Leases are
// longer than the enforced job deadline; the token fences out stale workers.
func (r *SubmissionsRepository) Claim(ctx context.Context) (*Job, error) {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx, `UPDATE submissions SET status='failed',verdict='system_error',error_message='judge retry limit exceeded',completed_at=CURRENT_TIMESTAMP WHERE id IN (SELECT submission_id FROM judge_jobs WHERE attempts>=3 AND lease_until<$1) AND status IN ('queued','running')`, now)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowContext(ctx, `UPDATE judge_jobs SET attempts=attempts+1,lease_token=lower(hex(randomblob(16))),lease_until=$1 WHERE submission_id=(SELECT j.submission_id FROM judge_jobs j JOIN submissions s ON s.id=j.submission_id WHERE s.status IN ('queued','running') AND j.attempts<3 AND j.available_at<=$2 AND j.lease_until<$2 ORDER BY s.queued_at,j.submission_id LIMIT 1) RETURNING submission_id,lease_token,attempts`, now+600, now)
	var j Job
	if err := row.Scan(&j.SubmissionID, &j.Token, &j.Attempts); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &j, nil
}
func (r *SubmissionsRepository) Retry(ctx context.Context, j Job, cause error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE judge_jobs SET lease_until=0,lease_token=NULL,available_at=$3,last_error=$4 WHERE submission_id=$1 AND lease_token=$2`, j.SubmissionID, j.Token, time.Now().Unix()+int64(j.Attempts*2), cause.Error())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil
	}
	status, verdict := "queued", "pending"
	if j.Attempts >= 3 {
		status, verdict = "failed", "system_error"
	}
	_, err = tx.ExecContext(ctx, `UPDATE submissions SET status=$2,verdict=$3,error_message=$4 WHERE id=$1`, j.SubmissionID, status, verdict, cause.Error())
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (r *SubmissionsRepository) LoadSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	var data sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT v.snapshot FROM submissions s LEFT JOIN problem_versions v ON v.id=s.problem_version_id WHERE s.id=$1`, id).Scan(&data)
	if err != nil {
		return nil, err
	}
	if !data.Valid {
		return nil, fmt.Errorf("historical submission has no frozen problem version; resubmit to judge current content")
	}
	var snapshot Snapshot
	if err = json.Unmarshal([]byte(data.String), &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}
