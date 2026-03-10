package model

import "time"

type ProblemSummary struct {
	ID               string   `json:"id"`
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	Difficulty       string   `json:"difficulty"`
	Category         string   `json:"category"`
	ProblemType      string   `json:"problemType"`
	ShortDescription string   `json:"shortDescription"`
	Tags             []string `json:"tags"`
}

type ProblemDetail struct {
	ProblemSummary
	StatementMarkdown   string             `json:"statementMarkdown"`
	ConstraintsMarkdown string             `json:"constraintsMarkdown"`
	Samples             []ProblemSample    `json:"samples"`
	LanguageTemplates   []LanguageTemplate `json:"languageTemplates"`
	Assets              []ProblemAsset     `json:"assets"`
	Metadata            map[string]any     `json:"metadata"`
}

type ProblemSample struct {
	Name        string `json:"name"`
	Input       string `json:"input"`
	Expected    string `json:"expected"`
	Explanation string `json:"explanation"`
	SortOrder   int    `json:"sortOrder"`
}

type LanguageTemplate struct {
	Language    string `json:"language"`
	StarterCode string `json:"starterCode"`
	Notes       string `json:"notes"`
}

type ProblemAsset struct {
	AssetType string         `json:"assetType"`
	Name      string         `json:"name"`
	MIMEType  string         `json:"mimeType"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata"`
}

type JudgeConfig struct {
	Runner string         `json:"runner"`
	Config map[string]any `json:"config"`
}

type JudgeTestCase struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	IsHidden   bool           `json:"isHidden"`
	Payload    map[string]any `json:"payload"`
	Weight     int            `json:"weight"`
	SortOrder  int            `json:"sortOrder"`
	DisplayIn  string         `json:"displayInput"`
	DisplayOut string         `json:"displayExpected"`
}

type ProblemFilter struct {
	Query      string
	Difficulty string
	Category   string
	Tag        string
}

type CompetitionRoomMember struct {
	UserID   string    `json:"userId"`
	Handle   string    `json:"handle"`
	IsHost   bool      `json:"isHost"`
	JoinedAt time.Time `json:"joinedAt"`
}

type CompetitionRoom struct {
	ID               string                  `json:"id"`
	Code             string                  `json:"code"`
	HostUserID       string                  `json:"hostUserId"`
	HostHandle       string                  `json:"hostHandle"`
	Name             string                  `json:"name"`
	Mode             string                  `json:"mode"`
	QuestionCount    int                     `json:"questionCount"`
	DifficultyPolicy string                  `json:"difficultyPolicy"`
	Status           string                  `json:"status"`
	Metadata         map[string]any          `json:"metadata"`
	CreatedAt        time.Time               `json:"createdAt"`
	UpdatedAt        time.Time               `json:"updatedAt"`
	Members          []CompetitionRoomMember `json:"members"`
}

type NewCompetitionRoom struct {
	Code             string
	HostUserID       string
	Name             string
	Mode             string
	QuestionCount    int
	DifficultyPolicy string
	Status           string
	Metadata         map[string]any
}

type Submission struct {
	ID            string                 `json:"id"`
	ProblemID     string                 `json:"problemId"`
	ProblemSlug   string                 `json:"problemSlug"`
	UserID        string                 `json:"userId"`
	Language      string                 `json:"language"`
	Mode          string                 `json:"mode"`
	SourceCode    string                 `json:"sourceCode,omitempty"`
	Status        string                 `json:"status"`
	Verdict       string                 `json:"verdict"`
	Score         int                    `json:"score"`
	TotalTests    int                    `json:"totalTests"`
	PassedTests   int                    `json:"passedTests"`
	CompileOutput string                 `json:"compileOutput,omitempty"`
	RuntimeOutput string                 `json:"runtimeOutput,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	QueuedAt      time.Time              `json:"queuedAt"`
	StartedAt     *time.Time             `json:"startedAt,omitempty"`
	CompletedAt   *time.Time             `json:"completedAt,omitempty"`
	Results       []SubmissionTestResult `json:"results"`
}

type SubmissionTestResult struct {
	CaseName    string `json:"caseName"`
	IsHidden    bool   `json:"isHidden"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	ExecutionMS int    `json:"executionMs"`
	SortOrder   int    `json:"sortOrder"`
	TestCaseID  string `json:"testCaseId,omitempty"`
}

type NewSubmission struct {
	UserID     string
	ProblemID  string
	Language   string
	Mode       string
	SourceCode string
}

type JudgeResult struct {
	Verdict       string
	Status        string
	Score         int
	TotalTests    int
	PassedTests   int
	CompileOutput string
	RuntimeOutput string
	ErrorMessage  string
	Results       []SubmissionTestResult
}
