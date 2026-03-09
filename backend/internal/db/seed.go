package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type seedTemplate struct {
	Language    string
	StarterCode string
	Notes       string
}

type seedAsset struct {
	AssetType string
	Name      string
	MIMEType  string
	Content   string
	Hidden    bool
	Metadata  map[string]any
}

type seedCase struct {
	Name          string
	DisplayInput  string
	DisplayExpect string
	Explanation   string
	Payload       map[string]any
	Hidden        bool
	Weight        int
	SortOrder     int
}

type seedProblem struct {
	Slug         string
	Title        string
	Difficulty   string
	Category     string
	ProblemType  string
	Short        string
	Statement    string
	Constraints  string
	Metadata     map[string]any
	Tags         []string
	Templates    []seedTemplate
	Assets       []seedAsset
	JudgeRunner  string
	JudgeConfig  map[string]any
	VisibleCases []seedCase
	HiddenCases  []seedCase
}

type diskProblemSpec struct {
	Slug            string         `json:"slug"`
	Title           string         `json:"title"`
	Difficulty      string         `json:"difficulty"`
	Category        string         `json:"category"`
	ProblemType     string         `json:"problemType"`
	Short           string         `json:"shortDescription"`
	StatementFile   string         `json:"statementFile"`
	ConstraintsFile string         `json:"constraintsFile"`
	Metadata        map[string]any `json:"metadata"`
	Tags            []string       `json:"tags"`
	Templates       []diskTemplate `json:"templates"`
	Assets          []diskAsset    `json:"assets"`
	Judge           diskJudge      `json:"judge"`
	VisibleCases    []diskCase     `json:"visibleCases"`
	HiddenCases     []diskCase     `json:"hiddenCases"`
}

type diskTemplate struct {
	Language        string `json:"language"`
	StarterCodeFile string `json:"starterCodeFile"`
	Notes           string `json:"notes"`
}

type diskAsset struct {
	AssetType   string         `json:"assetType"`
	Name        string         `json:"name"`
	MIMEType    string         `json:"mimeType"`
	ContentFile string         `json:"contentFile"`
	Hidden      bool           `json:"hidden"`
	Metadata    map[string]any `json:"metadata"`
}

type diskJudge struct {
	Runner string         `json:"runner"`
	Config map[string]any `json:"config"`
}

type diskCase struct {
	Name          string `json:"name"`
	DisplayInput  string `json:"displayInput"`
	DisplayExpect string `json:"displayExpect"`
	Explanation   string `json:"explanation"`
	CodeFile      string `json:"codeFile"`
	Weight        int    `json:"weight"`
	SortOrder     int    `json:"sortOrder"`
}

func SeedMVP(ctx context.Context, conn *sql.DB) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (handle, email)
		VALUES ('demo', 'demo@bitforge.local')
		ON CONFLICT (handle) DO UPDATE SET email = EXCLUDED.email
	`); err != nil {
		return fmt.Errorf("upsert demo user: %w", err)
	}

	problemsDir := os.Getenv("SEED_PROBLEMS_DIR")
	if problemsDir == "" {
		problemsDir = filepath.Join("seed", "problems")
	}

	problems, err := loadProblemsFromDisk(problemsDir)
	if err != nil {
		return fmt.Errorf("load seed problems: %w", err)
	}

	for _, problem := range problems {
		problemID, err := upsertProblem(ctx, tx, problem)
		if err != nil {
			return err
		}

		if err := replaceProblemRelations(ctx, tx, problemID, problem); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}
	return nil
}

func loadProblemsFromDisk(baseDir string) ([]seedProblem, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read seed directory %q: %w", baseDir, err)
	}

	dirs := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)

	problems := make([]seedProblem, 0, len(dirs))
	for _, dirName := range dirs {
		problemDir := filepath.Join(baseDir, dirName)
		spec, err := readProblemSpec(filepath.Join(problemDir, "problem.json"))
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", dirName, err)
		}

		converted, err := convertProblemSpec(problemDir, spec)
		if err != nil {
			return nil, fmt.Errorf("convert %s: %w", dirName, err)
		}
		problems = append(problems, converted)
	}

	if len(problems) == 0 {
		return nil, fmt.Errorf("no problems found in %s", baseDir)
	}

	return problems, nil
}

func readProblemSpec(path string) (diskProblemSpec, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return diskProblemSpec{}, fmt.Errorf("read problem spec: %w", err)
	}
	var spec diskProblemSpec
	if err := json.Unmarshal(bytes, &spec); err != nil {
		return diskProblemSpec{}, fmt.Errorf("decode problem spec json: %w", err)
	}
	return spec, nil
}

func convertProblemSpec(problemDir string, spec diskProblemSpec) (seedProblem, error) {
	if spec.Slug == "" {
		return seedProblem{}, fmt.Errorf("slug is required")
	}
	if spec.Title == "" {
		return seedProblem{}, fmt.Errorf("title is required")
	}
	if spec.Difficulty == "" {
		return seedProblem{}, fmt.Errorf("difficulty is required")
	}
	if spec.Category == "" {
		return seedProblem{}, fmt.Errorf("category is required")
	}
	if spec.ProblemType == "" {
		return seedProblem{}, fmt.Errorf("problemType is required")
	}
	if len(spec.Templates) == 0 {
		return seedProblem{}, fmt.Errorf("at least one template is required")
	}
	if len(spec.VisibleCases) == 0 {
		return seedProblem{}, fmt.Errorf("at least one visible case is required")
	}
	if len(spec.VisibleCases)+len(spec.HiddenCases) == 0 {
		return seedProblem{}, fmt.Errorf("at least one test case is required")
	}

	statement, err := readSeedText(problemDir, spec.StatementFile)
	if err != nil {
		return seedProblem{}, fmt.Errorf("read statement: %w", err)
	}
	constraints, err := readSeedText(problemDir, spec.ConstraintsFile)
	if err != nil {
		return seedProblem{}, fmt.Errorf("read constraints: %w", err)
	}

	templates := make([]seedTemplate, 0, len(spec.Templates))
	for _, tmpl := range spec.Templates {
		starter, err := readSeedText(problemDir, tmpl.StarterCodeFile)
		if err != nil {
			return seedProblem{}, fmt.Errorf("read template %s: %w", tmpl.Language, err)
		}
		templates = append(templates, seedTemplate{
			Language:    tmpl.Language,
			StarterCode: starter,
			Notes:       tmpl.Notes,
		})
	}

	assets := make([]seedAsset, 0, len(spec.Assets))
	for _, asset := range spec.Assets {
		content, err := readSeedText(problemDir, asset.ContentFile)
		if err != nil {
			return seedProblem{}, fmt.Errorf("read asset %s: %w", asset.Name, err)
		}
		assets = append(assets, seedAsset{
			AssetType: asset.AssetType,
			Name:      asset.Name,
			MIMEType:  asset.MIMEType,
			Content:   content,
			Hidden:    asset.Hidden,
			Metadata:  asset.Metadata,
		})
	}

	visibleCases, err := convertCases(problemDir, spec.VisibleCases, false)
	if err != nil {
		return seedProblem{}, fmt.Errorf("convert visible cases: %w", err)
	}
	hiddenCases, err := convertCases(problemDir, spec.HiddenCases, true)
	if err != nil {
		return seedProblem{}, fmt.Errorf("convert hidden cases: %w", err)
	}

	metadata := spec.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	judgeConfig := spec.Judge.Config
	if judgeConfig == nil {
		judgeConfig = map[string]any{}
	}

	return seedProblem{
		Slug:         spec.Slug,
		Title:        spec.Title,
		Difficulty:   spec.Difficulty,
		Category:     spec.Category,
		ProblemType:  spec.ProblemType,
		Short:        spec.Short,
		Statement:    statement,
		Constraints:  constraints,
		Metadata:     metadata,
		Tags:         spec.Tags,
		Templates:    templates,
		Assets:       assets,
		JudgeRunner:  spec.Judge.Runner,
		JudgeConfig:  judgeConfig,
		VisibleCases: visibleCases,
		HiddenCases:  hiddenCases,
	}, nil
}

func convertCases(problemDir string, source []diskCase, hidden bool) ([]seedCase, error) {
	cases := make([]seedCase, 0, len(source))
	for _, c := range source {
		code, err := readSeedText(problemDir, c.CodeFile)
		if err != nil {
			return nil, fmt.Errorf("read case code %s: %w", c.Name, err)
		}
		cases = append(cases, seedCase{
			Name:          c.Name,
			DisplayInput:  c.DisplayInput,
			DisplayExpect: c.DisplayExpect,
			Explanation:   c.Explanation,
			Payload:       map[string]any{"code": code},
			Hidden:        hidden,
			Weight:        max(c.Weight, 1),
			SortOrder:     c.SortOrder,
		})
	}
	return cases, nil
}

func readSeedText(problemDir, relativePath string) (string, error) {
	if relativePath == "" {
		return "", fmt.Errorf("file path is required")
	}
	root := filepath.Clean(problemDir)
	candidate := filepath.Clean(filepath.Join(problemDir, relativePath))
	if candidate != root && !strings.HasPrefix(candidate, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("file path %q escapes problem directory", relativePath)
	}
	bytes, err := os.ReadFile(candidate)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", candidate, err)
	}
	return string(bytes), nil
}

func upsertProblem(ctx context.Context, tx *sql.Tx, p seedProblem) (string, error) {
	const query = `
		INSERT INTO problems (
			slug,
			title,
			difficulty,
			category,
			problem_type,
			short_description,
			statement_md,
			constraints_md,
			metadata,
			is_published,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, TRUE, now())
		ON CONFLICT (slug)
		DO UPDATE SET
			title = EXCLUDED.title,
			difficulty = EXCLUDED.difficulty,
			category = EXCLUDED.category,
			problem_type = EXCLUDED.problem_type,
			short_description = EXCLUDED.short_description,
			statement_md = EXCLUDED.statement_md,
			constraints_md = EXCLUDED.constraints_md,
			metadata = EXCLUDED.metadata,
			is_published = TRUE,
			updated_at = now()
		RETURNING id
	`
	var id string
	if err := tx.QueryRowContext(
		ctx,
		query,
		p.Slug,
		p.Title,
		p.Difficulty,
		p.Category,
		p.ProblemType,
		p.Short,
		p.Statement,
		p.Constraints,
		mustJSON(p.Metadata),
	).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert problem %s: %w", p.Slug, err)
	}
	return id, nil
}

func replaceProblemRelations(ctx context.Context, tx *sql.Tx, problemID string, p seedProblem) error {
	for _, query := range []string{
		`DELETE FROM problem_tags WHERE problem_id = $1`,
		`DELETE FROM problem_language_templates WHERE problem_id = $1`,
		`DELETE FROM problem_assets WHERE problem_id = $1`,
		`DELETE FROM problem_test_cases WHERE problem_id = $1`,
		`DELETE FROM problem_judge_configs WHERE problem_id = $1`,
	} {
		if _, err := tx.ExecContext(ctx, query, problemID); err != nil {
			return fmt.Errorf("clear problem relations for %s: %w", p.Slug, err)
		}
	}

	for _, tag := range p.Tags {
		if _, err := tx.ExecContext(ctx, `INSERT INTO problem_tags (problem_id, tag) VALUES ($1, $2)`, problemID, tag); err != nil {
			return fmt.Errorf("insert tag %s for %s: %w", tag, p.Slug, err)
		}
	}

	for _, template := range p.Templates {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO problem_language_templates (problem_id, language, starter_code, notes)
			VALUES ($1, $2, $3, $4)
		`, problemID, template.Language, template.StarterCode, template.Notes); err != nil {
			return fmt.Errorf("insert template %s for %s: %w", template.Language, p.Slug, err)
		}
	}

	for _, asset := range p.Assets {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO problem_assets (problem_id, asset_type, name, mime_type, content_text, is_hidden, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		`, problemID, asset.AssetType, asset.Name, asset.MIMEType, asset.Content, asset.Hidden, mustJSON(asset.Metadata)); err != nil {
			return fmt.Errorf("insert asset %s for %s: %w", asset.Name, p.Slug, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO problem_judge_configs (problem_id, runner, config, updated_at)
		VALUES ($1, $2, $3::jsonb, now())
	`, problemID, p.JudgeRunner, mustJSON(p.JudgeConfig)); err != nil {
		return fmt.Errorf("insert judge config for %s: %w", p.Slug, err)
	}

	insertCase := `
		INSERT INTO problem_test_cases (
			problem_id,
			name,
			display_input,
			display_expected,
			explanation,
			payload,
			is_hidden,
			weight,
			sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
	`
	for _, c := range append(p.VisibleCases, p.HiddenCases...) {
		if _, err := tx.ExecContext(
			ctx,
			insertCase,
			problemID,
			c.Name,
			c.DisplayInput,
			c.DisplayExpect,
			c.Explanation,
			mustJSON(c.Payload),
			c.Hidden,
			max(c.Weight, 1),
			c.SortOrder,
		); err != nil {
			return fmt.Errorf("insert test case %s for %s: %w", c.Name, p.Slug, err)
		}
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mustJSON(v any) string {
	if v == nil {
		return "{}"
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
