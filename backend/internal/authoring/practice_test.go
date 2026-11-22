package authoring

import (
	"strings"
	"testing"

	"github.com/mattbriggs04/bitforge/backend/internal/curriculum"
)

func TestRolePresetsAndCustomPrompts(t *testing.T) {
	for _, role := range curriculum.Roles {
		for _, preset := range curriculum.Presets[:3] {
			config := Config{SchemaVersion: 1, Kind: "interview", Role: role.ID, Preset: preset.ID, Category: "random", Language: "", Format: "random", Difficulty: "medium", TargetMinutes: preset.TargetMinutes, Seed: 123}
			compiled, err := Compile(config)
			if err != nil {
				t.Fatal(role, preset, err)
			}
			if compiled.Config.Category != role.Category || compiled.Config.Language != role.Language || len(compiled.Config.Mix) == 0 {
				t.Fatalf("role defaults lost: %+v", compiled.Config)
			}
			again, err := Compile(compiled.Config)
			if err != nil || Hash(again) != Hash(compiled) {
				t.Fatal("resolved config does not reproduce prompt", err)
			}
			if !strings.Contains(compiled.Prompt, "import-interview") || !strings.Contains(compiled.Prompt, "hints=") {
				t.Fatal("missing publishing/hint contract")
			}
		}
	}
	config := Config{SchemaVersion: 1, Kind: "interview", Role: "embedded", Preset: "custom", Category: "embedded", Language: "rust", Difficulty: "hard", TargetMinutes: 20, Seed: 23, Mix: []curriculum.Mix{{Format: "numeric", Focus: "reasoning", Count: 2}}}
	resolved, err := config.Resolve()
	if err != nil || Hash(resolved.Mix) != Hash(config.Mix) {
		t.Fatal("custom mix overwritten", resolved, err)
	}
	for _, mix := range [][]curriculum.Mix{nil, {{Format: "essay", Focus: "reasoning", Count: 1}}, {{Format: "coding", Focus: "unknown", Count: 1}}, {{Format: "coding", Focus: "applied", Count: 21}}, {{Format: "coding", Focus: "applied", Count: 0}}} {
		config.Mix = mix
		if _, err = config.Resolve(); err == nil {
			t.Fatalf("invalid mix accepted: %v", mix)
		}
	}
}
func TestLearningContractRejectsMismatchedMembers(t *testing.T) {
	manifest, bundles, err := LoadInterview("../../../generation/interviews/binary-classifiers/interview.json")
	if err != nil {
		t.Fatal(err)
	}
	bad := manifest
	bad.Contract.TargetMinutes++
	if CheckInterviewMembers(bad, bundles) == nil {
		t.Fatal("mismatched time accepted")
	}
	bad = manifest
	bad.Contract.Role = "embedded"
	if CheckInterviewMembers(bad, bundles) == nil {
		t.Fatal("mismatched role accepted")
	}
	bad = manifest
	bad.Contract.Level = "Graduate course"
	if CheckInterviewMembers(bad, bundles) == nil {
		t.Fatal("mismatched level accepted")
	}
	bad = manifest
	bad.Contract.Preset = "interview"
	if err = CheckInterviewMembers(bad, bundles); err != nil {
		t.Fatal("valid three-focus contract rejected", err)
	}
	changed := append([]Bundle(nil), bundles...)
	for i := range changed {
		changed[i].Practice.Focus = "fundamentals"
	}
	if CheckInterviewMembers(bad, changed) == nil {
		t.Fatal("preset without applied/reasoning accepted")
	}
	bad.Contract.Preset = "custom"
	if err = CheckInterviewMembers(bad, changed); err != nil {
		t.Fatal("custom mix should allow a single focus", err)
	}
	b := bundles[0]
	b.Hints.Approach = ""
	if b.Check() == nil {
		t.Fatal("missing approach accepted")
	}
	b = bundles[0]
	b.Practice.Focus = "trivia"
	if b.Check() == nil {
		t.Fatal("unknown focus accepted")
	}
}
