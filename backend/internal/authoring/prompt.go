package authoring

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"slices"
	"strings"

	"github.com/mattbriggs04/bitforge/backend/internal/curriculum"
)

type Config struct {
	Kind          string           `json:"kind,omitempty"`
	Role          string           `json:"role,omitempty"`
	Preset        string           `json:"preset,omitempty"`
	Mix           []curriculum.Mix `json:"mix,omitempty"`
	Objectives    []string         `json:"objectives,omitempty"`
	SchemaVersion int              `json:"schemaVersion"`
	Seed          int64            `json:"seed"`
	Category      string           `json:"category"`
	Format        string           `json:"format,omitempty"`
	Language      string           `json:"language"`
	Difficulty    string           `json:"difficulty"`
	Level         string           `json:"level"`
	Topic         string           `json:"topic"`
	TargetMinutes int              `json:"targetMinutes"`
	Avoid         []string         `json:"avoid"`
}

func (c Config) Resolve() (Config, error) {
	if c.Seed < 0 {
		return c, fmt.Errorf("seed must be nonnegative")
	}
	c.Objectives = slices.DeleteFunc(slices.Clone(c.Objectives), func(s string) bool { return strings.TrimSpace(s) == "" })
	c.Avoid = slices.DeleteFunc(slices.Clone(c.Avoid), func(s string) bool { return strings.TrimSpace(s) == "" })
	if c.SchemaVersion != 1 {
		return c, fmt.Errorf("schemaVersion must be 1")
	}
	if c.Seed == 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(1<<52))
		if err != nil {
			return c, err
		}
		c.Seed = n.Int64() + 1
	}
	if c.Kind == "" {
		c.Kind = "question"
	}
	if c.Kind != "question" && c.Kind != "interview" {
		return c, fmt.Errorf("kind must be question or interview")
	}
	if c.Role != "" && c.Role != "random" {
		role, ok := curriculum.FindRole(c.Role)
		if !ok {
			return c, fmt.Errorf("unknown role")
		}
		if c.Category == "" || c.Category == "random" {
			c.Category = role.Category
		}
		if c.Language == "" {
			c.Language = role.Language
		}
	}
	r := mathrand.New(mathrand.NewSource(c.Seed))
	choose := func(value *string, options []string) error {
		if *value == "random" {
			*value = options[r.Intn(len(options))]
		}
		if !slices.Contains(options, *value) {
			return fmt.Errorf("unsupported selection %q", *value)
		}
		return nil
	}
	if c.Category == "" {
		c.Category = "random"
	}
	if err := choose(&c.Category, []string{"embedded", "ai_ml", "computer_vision", "computer_architecture", "operating_systems"}); err != nil {
		return c, err
	}
	if c.Role == "" || c.Role == "random" {
		c.Role = c.Category
	}
	if c.Language == "" {
		role, _ := curriculum.FindRole(c.Role)
		c.Language = role.Language
	}
	if c.Difficulty == "" {
		c.Difficulty = "medium"
	}
	if c.Kind == "question" {
		if c.Format == "" {
			c.Format = "random"
		}
		if err := choose(&c.Format, curriculum.Formats); err != nil {
			return c, err
		}
	} else {
		c.Format = ""
	}

	if err := choose(&c.Language, []string{"c", "cpp", "python", "rust"}); err != nil {
		return c, err
	}
	if err := choose(&c.Difficulty, []string{"easy", "medium", "hard"}); err != nil {
		return c, err
	}
	if strings.TrimSpace(c.Level) == "" {
		c.Level = "Full-time interview"
	}
	if strings.TrimSpace(c.Topic) == "" {
		c.Topic = "random"
	}
	if len(c.Topic) > 2000 || len(c.Level) > 200 || len(c.Avoid) > 100 {
		return c, fmt.Errorf("request text too long")
	}
	if c.Kind == "interview" {
		if c.Preset == "" {
			c.Preset = "interview"
		}
		if c.Preset != "custom" {
			c.Mix = curriculum.Defaults(c.Role, c.Preset)
		}
		contract := curriculum.Contract{Role: c.Role, Preset: c.Preset, Level: c.Level, TargetMinutes: c.TargetMinutes, Objectives: []string{"placeholder"}}
		if err := contract.Check(); err != nil {
			return c, err
		}
		if err := curriculum.CheckMix(c.Mix); err != nil {
			return c, err
		}
		count := 0
		for _, m := range c.Mix {
			count += m.Count
		}
		if c.TargetMinutes < count {
			return c, fmt.Errorf("allow at least one minute per question")
		}
	} else {
		c.Preset = "custom"
		c.Mix = nil
	}
	if c.TargetMinutes < 1 || c.TargetMinutes > 180 {
		return c, fmt.Errorf("targetMinutes must be 1..180")
	}
	return c, nil
}
func Prompt(c Config) (string, error) {
	resolved, err := c.Resolve()
	if err != nil {
		return "", err
	}
	return promptResolved(resolved), nil
}

type CompiledPrompt struct {
	Config Config `json:"config"`
	Prompt string `json:"prompt"`
}

func Compile(c Config) (CompiledPrompt, error) {
	resolved, err := c.Resolve()
	if err != nil {
		return CompiledPrompt{}, err
	}
	return CompiledPrompt{resolved, promptResolved(resolved)}, nil
}
func promptResolved(resolved Config) string {
	data, _ := json.MarshalIndent(resolved, "", "  ")
	intro := "Create exactly one BitForge question using this resolved request:"
	ending := ""
	location := "generation/drafts/<slug>.json"
	validateCommand := "cd backend && go run ./cmd/problems validate -bundle ../generation/drafts/<slug>.json"
	importCommand := "cd backend && go run ./cmd/problems import -bundle ../generation/drafts/<slug>.json"
	resultPath := "bundle path"
	if resolved.Kind == "interview" {
		intro = "Create a BitForge interview using this resolved contract and exact question mix:"
		location = "generation/interviews/<new-slug>/ (all question bundles plus interview.json)"
		validateCommand = "cd backend && go run ./cmd/problems validate-interview -bundle ../generation/interviews/<new-slug>/interview.json"
		importCommand = "cd backend && go run ./cmd/problems import-interview -bundle ../generation/interviews/<new-slug>/interview.json"
		resultPath = "manifest path, resolved contract"
		ending = `
INTERVIEW CONTRACT: Read backend/internal/authoring/interview.go. Write all NEW bundles and interview.json together in generation/interviews/<new-slug>/. The manifest has schemaVersion=2, slug, title, category, description, contract, and questions (ordered relative filenames). contract has role, preset, level, targetMinutes, objectives. Use the resolved mix exactly, assigning each member its requested format and practice.focus. Every member's practice.role and practice.level must match the contract. Allocate integer question time budgets that SUM EXACTLY to the contract targetMinutes. Author measurable collection objectives (use requested objectives if supplied). All members follow the full validation contract above. Write a coherent role-specific assessment, not unrelated trivia. For preset interview cover fundamentals, applied work, and reasoning; daily covers fundamentals and applied; exam covers fundamentals and reasoning. Custom permits any explicit mix. Applied work can be a numerical trace rather than coding where the role calls for it. Put testing/debugging/tradeoff reasoning into the member tasks and explanations, not unsupported free-text grading. Do not enforce a universal 3-choice/1-code formula.
Import the collection once; it validates and publishes every member atomically. Return the manifest, resolved contract, and actual validation results.
`
	}
	role, _ := curriculum.FindRole(resolved.Role)

	return fmt.Sprintf(`%s
%s

Read AGENTS.md first. Role emphasis: %s

Read backend/internal/authoring/bundle.go for the strict JSON bundle contract and generation/examples/*.json for complete examples. Write NEW content at %s; do not modify existing questions, the validator, worker, tests or publication code. Never place private artifacts in frontend/public.

1. Define the specification, assumptions, units, learning objectives, and grading contract before solving. Match the requested role, category, question format or interview mix, topic, level, difficulty, time budget, language, seed, and any supplied objectives. Avoid the supplied topics/problems. Use an original problem with a precise unambiguous answer.
2. Emit schemaVersion=2. Include practice={"role":"<resolved role>","level":"<resolved level>","focus":"fundamentals|applied|reasoning"}. Focus describes what the question tests, independently of its answer format. Keep objectives measurable and targetMinutes realistic. Include hints={"nudge":"...","approach":"..."}: the nudge points to a concept or useful question; the approach explains a strategy without a final answer or complete implementation. Both are public, progressively revealed Markdown supporting LaTeX. For choice questions never reveal the correct ID or eliminate every alternative. Scaffolding and walkthroughs belong to on-demand tutoring, not stored hints. Include provenance (agent/model if known; never invent it), explanation, and reviewNotes describing independent reasoning and limitations. Difficulty is easy/medium/hard; graduate level belongs in the requested content, not a new enum.
3. Coding: full standalone stdin/stdout programs, not function snippets. Use only installed standard libraries or Python NumPy. Include a compiling starter in templates, one reference per advertised language, a DIFFERENT simpler independent verifier, and at least two distinct plausible compiling mutants. Comparators are exact (trimmed text), numeric (one finite number), or json (recursive numeric tolerance); specify any tolerances in the statement. No custom checker code is accepted.
4. Include 2..20 hand-calculated anchor cases, with both public and private coverage. Provide a deterministic generator program: it reads the integer seed from stdin and prints ONLY a JSON array of 20..100 objects {"name":"unique-case-name","input":"stdin text"}. It generates inputs, never expected answers. Cover boundaries, invalid inputs only where the specification defines behavior, representative random inputs, and relevant invariants.
5. Answer formats: no templates/references/generator/cases/mutants. Numeric answer is a finite numeric string with units specified in the statement. Choice options have stable IDs, text, and a rationale for EVERY option. Answer is an encoded JSON string ID for single_choice or encoded JSON array of IDs for multiple_select. Choice comparator kind is json. Provide an independent executable verifier whose stdout matches answer, based on a real calculation/model rather than simply printing the answer. Verify theoretical reasoning separately in reviewNotes; execution alone cannot prove conceptual truth.
Example authoring: visible coding examples must be reproducible. For stdin/stdout programs, provide the complete literal input and output. For function exercises, show argument values, buffer bytes, lengths, initial state, the exact call sequence, and concrete return values plus mutated outputs. Do not replace data with summaries such as "8-byte payload" or "fields decoded". Explain each specific case briefly and keep visible examples consistent with the actual public tests. Example input/output stays literal; LaTeX belongs only in explanatory prose.

Math authoring: use LaTeX math in Markdown: $...$ inline and $$ on separate lines for display equations. This is supported in statement, constraints, explanation, option rationales, and visible-case explanation prose. Option text supports INLINE math only. Prefer \frac, \nabla, \sum, and proper subscripts to Unicode approximations. In JSON source, double each LaTeX backslash (for example, "$\\nabla L(w)$"). Use JSON serialization to avoid accidental \n, \t, \b, or \f escape sequences. Keep titles, shortDescription, tags, test input/output, code, and stored answer IDs plain. Literal dollar amounts must escape the dollar sign in Markdown. Use only KaTeX-supported notation; do not use HTML, remote resources, color commands, or document/preamble commands. Preview math in the browser before publication.

6. Keep solutions, correct option IDs, rationales, private tests, and verifier code out of public statements and templates. Reference and submitted programs run separately; do not put expected outputs or grader logic inside user programs.
7. Run: %s
Repair actual failures within two attempts. If Docker is unavailable or validation fails, leave a draft and report the failure. NEVER claim checks ran when they did not. Do not use host execution as a fallback.
8. To publish, run: %s
Import reruns validation and freezes actual cases. Do not mark a draft published by editing a report. Return the %s and actual validation result.
`, intro, data, role.Emphasis, location, validateCommand, importCommand, resultPath) + ending
}
