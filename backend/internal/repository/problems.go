package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
)

type ProblemsRepository struct {
	db *sql.DB
}

func NewProblemsRepository(db *sql.DB) *ProblemsRepository {
	return &ProblemsRepository{db: db}
}

func (r *ProblemsRepository) ListPublished(ctx context.Context, filter model.ProblemFilter) ([]model.ProblemSummary, error) {
	args := []any{}
	where := []string{"is_published = TRUE"}

	if filter.Query != "" {
		args = append(args, "%"+filter.Query+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR short_description ILIKE $%d OR statement_md ILIKE $%d)", idx, idx, idx))
	}
	if filter.Difficulty != "" {
		args = append(args, strings.ToLower(filter.Difficulty))
		where = append(where, fmt.Sprintf("difficulty = $%d", len(args)))
	}
	if filter.Category != "" {
		args = append(args, filter.Category)
		where = append(where, fmt.Sprintf("category = $%d", len(args)))
	}
	if filter.Tag != "" {
		args = append(args, filter.Tag)
		where = append(where, fmt.Sprintf("id IN (SELECT problem_id FROM problem_tags WHERE tag = $%d)", len(args)))
	}

	query := `
		SELECT id, slug, title, difficulty, category, problem_type, short_description
		FROM problems
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY
			CASE difficulty WHEN 'easy' THEN 1 WHEN 'medium' THEN 2 WHEN 'hard' THEN 3 ELSE 4 END,
			title ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list problems: %w", err)
	}
	defer rows.Close()

	items := make([]model.ProblemSummary, 0)
	problemIDs := make([]string, 0)
	for rows.Next() {
		var item model.ProblemSummary
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Difficulty,
			&item.Category,
			&item.ProblemType,
			&item.ShortDescription,
		); err != nil {
			return nil, fmt.Errorf("scan problem summary: %w", err)
		}
		items = append(items, item)
		problemIDs = append(problemIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate problem summaries: %w", err)
	}

	tagsByProblem, err := r.getTagsForProblems(ctx, problemIDs)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Tags = tagsByProblem[items[i].ID]
	}

	return items, nil
}

func (r *ProblemsRepository) GetPublishedBySlug(ctx context.Context, slug string) (*model.ProblemDetail, error) {
	const query = `
		SELECT id, slug, title, difficulty, category, problem_type, short_description, statement_md, constraints_md, metadata
		FROM problems
		WHERE slug = $1 AND is_published = TRUE
	`

	var detail model.ProblemDetail
	var metadataRaw []byte
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&detail.ID,
		&detail.Slug,
		&detail.Title,
		&detail.Difficulty,
		&detail.Category,
		&detail.ProblemType,
		&detail.ShortDescription,
		&detail.StatementMarkdown,
		&detail.ConstraintsMarkdown,
		&metadataRaw,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get problem: %w", err)
	}

	tagsByProblem, err := r.getTagsForProblems(ctx, []string{detail.ID})
	if err != nil {
		return nil, err
	}
	detail.Tags = tagsByProblem[detail.ID]

	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &detail.Metadata); err != nil {
			return nil, fmt.Errorf("decode problem metadata: %w", err)
		}
	} else {
		detail.Metadata = map[string]any{}
	}

	samples, err := r.getSamples(ctx, detail.ID)
	if err != nil {
		return nil, err
	}
	detail.Samples = samples

	templates, err := r.getTemplates(ctx, detail.ID)
	if err != nil {
		return nil, err
	}
	detail.LanguageTemplates = templates

	assets, err := r.getAssets(ctx, detail.ID)
	if err != nil {
		return nil, err
	}
	detail.Assets = assets

	return &detail, nil
}

func (r *ProblemsRepository) GetBySlug(ctx context.Context, slug string) (*model.ProblemSummary, error) {
	const query = `
		SELECT id, slug, title, difficulty, category, problem_type, short_description
		FROM problems
		WHERE slug = $1
	`
	var item model.ProblemSummary
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&item.ID,
		&item.Slug,
		&item.Title,
		&item.Difficulty,
		&item.Category,
		&item.ProblemType,
		&item.ShortDescription,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get problem by slug: %w", err)
	}
	tagsByProblem, err := r.getTagsForProblems(ctx, []string{item.ID})
	if err != nil {
		return nil, err
	}
	item.Tags = tagsByProblem[item.ID]
	return &item, nil
}

func (r *ProblemsRepository) GetStarterCode(ctx context.Context, problemID, language string) (string, error) {
	const query = `
		SELECT starter_code
		FROM problem_language_templates
		WHERE problem_id = $1 AND language = $2
	`
	var starter string
	if err := r.db.QueryRowContext(ctx, query, problemID, strings.ToLower(language)).Scan(&starter); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get starter code: %w", err)
	}
	return starter, nil
}

func (r *ProblemsRepository) GetJudgeConfig(ctx context.Context, problemID string) (*model.JudgeConfig, error) {
	const query = `
		SELECT runner, config
		FROM problem_judge_configs
		WHERE problem_id = $1
	`
	var cfg model.JudgeConfig
	var configRaw []byte
	if err := r.db.QueryRowContext(ctx, query, problemID).Scan(&cfg.Runner, &configRaw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get judge config: %w", err)
	}
	cfg.Config = map[string]any{}
	if len(configRaw) > 0 {
		if err := json.Unmarshal(configRaw, &cfg.Config); err != nil {
			return nil, fmt.Errorf("decode judge config: %w", err)
		}
	}
	return &cfg, nil
}

func (r *ProblemsRepository) GetTestCases(ctx context.Context, problemID string, includeHidden bool) ([]model.JudgeTestCase, error) {
	query := `
		SELECT id, name, is_hidden, payload, weight, sort_order, display_input, display_expected
		FROM problem_test_cases
		WHERE problem_id = $1
	`
	args := []any{problemID}
	if !includeHidden {
		query += ` AND is_hidden = FALSE`
	}
	query += ` ORDER BY sort_order ASC, name ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get test cases: %w", err)
	}
	defer rows.Close()

	cases := make([]model.JudgeTestCase, 0)
	for rows.Next() {
		var testCase model.JudgeTestCase
		var payloadRaw []byte
		if err := rows.Scan(
			&testCase.ID,
			&testCase.Name,
			&testCase.IsHidden,
			&payloadRaw,
			&testCase.Weight,
			&testCase.SortOrder,
			&testCase.DisplayIn,
			&testCase.DisplayOut,
		); err != nil {
			return nil, fmt.Errorf("scan test case: %w", err)
		}
		testCase.Payload = map[string]any{}
		if len(payloadRaw) > 0 {
			if err := json.Unmarshal(payloadRaw, &testCase.Payload); err != nil {
				return nil, fmt.Errorf("decode test case payload: %w", err)
			}
		}
		cases = append(cases, testCase)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate test cases: %w", err)
	}

	return cases, nil
}

func (r *ProblemsRepository) getSamples(ctx context.Context, problemID string) ([]model.ProblemSample, error) {
	const query = `
		SELECT name, display_input, display_expected, explanation, sort_order
		FROM problem_test_cases
		WHERE problem_id = $1 AND is_hidden = FALSE
		ORDER BY sort_order ASC, name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, problemID)
	if err != nil {
		return nil, fmt.Errorf("query samples: %w", err)
	}
	defer rows.Close()

	samples := make([]model.ProblemSample, 0)
	for rows.Next() {
		var sample model.ProblemSample
		if err := rows.Scan(&sample.Name, &sample.Input, &sample.Expected, &sample.Explanation, &sample.SortOrder); err != nil {
			return nil, fmt.Errorf("scan sample: %w", err)
		}
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate samples: %w", err)
	}
	return samples, nil
}

func (r *ProblemsRepository) getTemplates(ctx context.Context, problemID string) ([]model.LanguageTemplate, error) {
	const query = `
		SELECT language, starter_code, notes
		FROM problem_language_templates
		WHERE problem_id = $1
		ORDER BY language ASC
	`
	rows, err := r.db.QueryContext(ctx, query, problemID)
	if err != nil {
		return nil, fmt.Errorf("query language templates: %w", err)
	}
	defer rows.Close()

	templates := make([]model.LanguageTemplate, 0)
	for rows.Next() {
		var tmpl model.LanguageTemplate
		if err := rows.Scan(&tmpl.Language, &tmpl.StarterCode, &tmpl.Notes); err != nil {
			return nil, fmt.Errorf("scan language template: %w", err)
		}
		templates = append(templates, tmpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate language templates: %w", err)
	}
	return templates, nil
}

func (r *ProblemsRepository) getAssets(ctx context.Context, problemID string) ([]model.ProblemAsset, error) {
	const query = `
		SELECT asset_type, name, mime_type, content_text, metadata
		FROM problem_assets
		WHERE problem_id = $1 AND is_hidden = FALSE
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, problemID)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	assets := make([]model.ProblemAsset, 0)
	for rows.Next() {
		var asset model.ProblemAsset
		var metadataRaw []byte
		if err := rows.Scan(&asset.AssetType, &asset.Name, &asset.MIMEType, &asset.Content, &metadataRaw); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		asset.Metadata = map[string]any{}
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &asset.Metadata); err != nil {
				return nil, fmt.Errorf("decode asset metadata: %w", err)
			}
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assets: %w", err)
	}
	return assets, nil
}

func (r *ProblemsRepository) getTagsForProblems(ctx context.Context, problemIDs []string) (map[string][]string, error) {
	result := make(map[string][]string, len(problemIDs))
	if len(problemIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(problemIDs))
	args := make([]any, 0, len(problemIDs))
	for i, id := range problemIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}

	query := `
		SELECT problem_id::text, tag
		FROM problem_tags
		WHERE problem_id IN (` + strings.Join(placeholders, ", ") + `)
		ORDER BY tag ASC
	`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query problem tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var problemID, tag string
		if err := rows.Scan(&problemID, &tag); err != nil {
			return nil, fmt.Errorf("scan problem tag: %w", err)
		}
		result[problemID] = append(result[problemID], tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate problem tags: %w", err)
	}

	return result, nil
}
