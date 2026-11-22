package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
)

// An interview with an unpublished member is itself unavailable. Filtering only
// the hidden child would silently change the collection and its question order.
const interviewSummaryQuery = `SELECT i.slug,i.title,i.category,i.description,count(p.id),COALESCE(sum(CAST(json_extract(p.metadata,'$.targetMinutes') AS INTEGER)),0),i.contract
 FROM interviews i JOIN interview_problems ip ON ip.interview_slug=i.slug JOIN problems p ON p.id=ip.problem_id
 WHERE NOT EXISTS(SELECT 1 FROM interview_problems hidden JOIN problems hp ON hp.id=hidden.problem_id WHERE hidden.interview_slug=i.slug AND hp.is_published=0)`

func (r *ProblemsRepository) ListInterviews(ctx context.Context, problemID string) ([]model.InterviewSummary, error) {
	query := interviewSummaryQuery
	args := []any{}
	if problemID != "" {
		query += ` AND EXISTS(SELECT 1 FROM interview_problems member WHERE member.interview_slug=i.slug AND member.problem_id=$1)`
		args = append(args, problemID)
	}
	query += ` GROUP BY i.slug ORDER BY i.created_at DESC,i.slug`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.InterviewSummary{}
	for rows.Next() {
		var item model.InterviewSummary
		var contract []byte
		if err = rows.Scan(&item.Slug, &item.Title, &item.Category, &item.Description, &item.QuestionCount, &item.TargetMinutes, &contract); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(contract, &item.Contract); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *ProblemsRepository) GetInterview(ctx context.Context, slug string) (*model.InterviewDetail, error) {
	var detail model.InterviewDetail
	var contract []byte
	if err := r.db.QueryRowContext(ctx, interviewSummaryQuery+` AND i.slug=$1 GROUP BY i.slug`, slug).Scan(&detail.Slug, &detail.Title, &detail.Category, &detail.Description, &detail.QuestionCount, &detail.TargetMinutes, &contract); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(contract, &detail.Contract); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT p.id,p.slug,p.title,p.difficulty,p.category,p.problem_type,p.short_description,p.metadata FROM interview_problems ip JOIN problems p ON p.id=ip.problem_id WHERE ip.interview_slug=$1 AND p.is_published=1 ORDER BY ip.position`, slug)
	if err != nil {
		return nil, err
	}
	detail.Problems = []model.ProblemSummary{}
	ids := []string{}
	for rows.Next() {
		var item model.ProblemSummary
		var metadata []byte
		if err = rows.Scan(&item.ID, &item.Slug, &item.Title, &item.Difficulty, &item.Category, &item.ProblemType, &item.ShortDescription, &metadata); err != nil {
			rows.Close()
			return nil, err
		}
		if err = hydratePractice(&item, metadata); err != nil {
			rows.Close()
			return nil, err
		}
		detail.Problems = append(detail.Problems, item)
		ids = append(ids, item.ID)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	tags, err := r.getTagsForProblems(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("interview tags: %w", err)
	}
	for i := range detail.Problems {
		detail.Problems[i].Tags = tags[detail.Problems[i].ID]
	}
	return &detail, nil
}
