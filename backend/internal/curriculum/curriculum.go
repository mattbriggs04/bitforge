// Package curriculum owns the shared practice vocabulary and role defaults.
package curriculum

import (
	"fmt"
	"slices"
	"strings"
)

type Role struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Category string `json:"category"`
	Language string `json:"language"`
	Emphasis string `json:"emphasis"`
}

var Roles = []Role{
	{"embedded", "Embedded engineer", "embedded", "c", "Byte-level reasoning, fixed-width arithmetic, memory, peripherals, and concurrency."},
	{"ai_ml", "ML engineer", "ai_ml", "python", "Model behavior, gradients, numerical stability, evaluation, and implementation."},
	{"computer_vision", "Computer vision engineer", "computer_vision", "python", "Geometry, image operations, numerical reasoning, and algorithm implementation."},
	{"computer_architecture", "Computer architect", "computer_architecture", "cpp", "Data representation, caches, pipelines, memory ordering, and performance models."},
	{"operating_systems", "Systems engineer", "operating_systems", "c", "Processes, threads, synchronization, virtual memory, scheduling, and Linux behavior."},
}

type Preset struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	TargetMinutes int    `json:"targetMinutes"`
}

var Presets = []Preset{{"daily", "Daily practice", 25}, {"interview", "Technical interview", 60}, {"exam", "Exam practice", 60}, {"custom", "Custom", 45}}

type Profile struct {
	Role  string `json:"role"`
	Level string `json:"level"`
	Focus string `json:"focus"`
}
type Hints struct {
	Nudge    string `json:"nudge"`
	Approach string `json:"approach"`
}
type Contract struct {
	Role          string   `json:"role"`
	Preset        string   `json:"preset"`
	Level         string   `json:"level"`
	TargetMinutes int      `json:"targetMinutes"`
	Objectives    []string `json:"objectives"`
}
type Mix struct {
	Format string `json:"format"`
	Focus  string `json:"focus"`
	Count  int    `json:"count"`
}

var Formats = []string{"coding", "single_choice", "multiple_select", "numeric"}
var Focuses = []string{"fundamentals", "applied", "reasoning"}

func FindRole(id string) (Role, bool) {
	for _, r := range Roles {
		if r.ID == id {
			return r, true
		}
	}
	return Role{}, false
}
func (p Profile) Check() error {
	if _, ok := FindRole(p.Role); !ok {
		return fmt.Errorf("unknown practice role %q", p.Role)
	}
	if strings.TrimSpace(p.Level) == "" || !slices.Contains(Focuses, p.Focus) {
		return fmt.Errorf("practice requires level and fundamentals/applied/reasoning focus")
	}
	return nil
}
func (h Hints) Check() error {
	if strings.TrimSpace(h.Nudge) == "" || strings.TrimSpace(h.Approach) == "" || len(h.Nudge) > 4000 || len(h.Approach) > 8000 {
		return fmt.Errorf("hints require a short nudge and approach")
	}
	return nil
}
func (c Contract) Check() error {
	if _, ok := FindRole(c.Role); !ok {
		return fmt.Errorf("unknown contract role")
	}
	if !slices.Contains([]string{"daily", "interview", "exam", "custom"}, c.Preset) || strings.TrimSpace(c.Level) == "" || c.TargetMinutes < 1 || c.TargetMinutes > 180 || len(c.Objectives) < 1 {
		return fmt.Errorf("contract requires preset, level, 1..180 minutes and objectives")
	}
	for _, o := range c.Objectives {
		if strings.TrimSpace(o) == "" {
			return fmt.Errorf("empty contract objective")
		}
	}
	return nil
}
func CheckMix(mix []Mix) error {
	count := 0
	for _, m := range mix {
		if !slices.Contains(Formats, m.Format) || !slices.Contains(Focuses, m.Focus) || m.Count < 1 || m.Count > 20 {
			return fmt.Errorf("invalid question mix")
		}
		count += m.Count
	}
	if count < 1 || count > 20 {
		return fmt.Errorf("select 1..20 questions")
	}
	return nil
}

// Defaults are suggestions resolved into an explicit mix, never a fixed 3+1 rule.
func Defaults(role, preset string) []Mix {
	applied := "coding"
	check := "single_choice"
	if role == "computer_architecture" {
		applied = "numeric"
	}
	if role == "ai_ml" || role == "computer_vision" {
		check = "numeric"
	}
	switch preset {
	case "daily":
		return []Mix{{check, "fundamentals", 1}, {applied, "applied", 1}}
	case "exam":
		return []Mix{{check, "fundamentals", 2}, {"numeric", "reasoning", 2}, {applied, "applied", 1}}
	default:
		return []Mix{{"single_choice", "fundamentals", 1}, {applied, "applied", 1}, {"multiple_select", "reasoning", 1}}
	}
}
