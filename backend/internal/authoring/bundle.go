// Package authoring defines the versioned contract between an authoring agent and BitForge.
package authoring

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/mattbriggs04/bitforge/backend/internal/judge"
)

type Program struct {
	Language string `json:"language"`
	Source   string `json:"source"`
}
type Option struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Rationale string `json:"rationale"`
}
type Case struct {
	Name     string `json:"name"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Hidden   bool   `json:"hidden"`
}
type Bundle struct {
	SchemaVersion    int               `json:"schemaVersion"`
	Slug             string            `json:"slug"`
	Title            string            `json:"title"`
	Category         string            `json:"category"`
	Difficulty       string            `json:"difficulty"`
	Format           string            `json:"format"`
	ShortDescription string            `json:"shortDescription"`
	Statement        string            `json:"statement"`
	Constraints      string            `json:"constraints"`
	Objectives       []string          `json:"objectives"`
	Tags             []string          `json:"tags"`
	TargetMinutes    int               `json:"targetMinutes"`
	Seed             int64             `json:"seed"`
	Provenance       string            `json:"provenance"`
	Templates        map[string]string `json:"templates,omitempty"`
	References       []Program         `json:"references,omitempty"`
	Independent      Program           `json:"independent"`
	Generator        *Program          `json:"generator,omitempty"`
	Mutants          []Program         `json:"mutants,omitempty"`
	Cases            []Case            `json:"cases,omitempty"`
	Comparator       judge.Comparator  `json:"comparator"`
	Options          []Option          `json:"options,omitempty"`
	Answer           string            `json:"answer,omitempty"`
	Explanation      string            `json:"explanation"`
	ReviewNotes      string            `json:"reviewNotes"`
}

func Decode(data []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("expected one JSON value")
	}
	return nil
}
func Hash(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (p Program) Check() error {
	if !slices.Contains([]string{"c", "cpp", "python", "rust"}, p.Language) || strings.TrimSpace(p.Source) == "" || len(p.Source) > 300000 {
		return fmt.Errorf("invalid program language or source")
	}
	return nil
}
func (b Bundle) Check() error {
	if b.SchemaVersion != 1 || !slugPattern.MatchString(b.Slug) || len(b.Slug) > 100 {
		return fmt.Errorf("invalid schemaVersion or slug")
	}
	for _, v := range []string{b.Title, b.Category, b.Statement, b.Constraints, b.ShortDescription, b.Explanation, b.ReviewNotes, b.Provenance} {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("required text field is empty")
		}
	}
	if !slices.Contains([]string{"embedded", "ai_ml", "computer_vision", "computer_architecture", "operating_systems"}, b.Category) {
		return fmt.Errorf("invalid category")
	}
	if !slices.Contains([]string{"easy", "medium", "hard"}, b.Difficulty) {
		return fmt.Errorf("invalid difficulty")
	}
	if len(b.Objectives) == 0 || b.TargetMinutes < 1 || b.TargetMinutes > 180 {
		return fmt.Errorf("objectives and targetMinutes required")
	}
	if err := b.Comparator.Validate(); err != nil {
		return err
	}
	if err := b.Independent.Check(); err != nil {
		return fmt.Errorf("independent verifier: %w", err)
	}
	if b.Format == "coding" {
		if len(b.References) == 0 || len(b.Templates) != len(b.References) || len(b.Mutants) < 2 || b.Generator == nil {
			return fmt.Errorf("coding requires templates, references, generator, and at least two mutants")
		}
		if err := b.Generator.Check(); err != nil {
			return err
		}
		langs := map[string]bool{}
		for _, p := range b.References {
			if err := p.Check(); err != nil {
				return err
			}
			if langs[p.Language] || strings.TrimSpace(b.Templates[p.Language]) == "" {
				return fmt.Errorf("duplicate reference or missing starter")
			}
			langs[p.Language] = true
			if p.Source == b.Independent.Source {
				return fmt.Errorf("independent verifier must differ from reference")
			}
		}
		mutants := map[string]bool{}
		for _, p := range b.Mutants {
			if mutants[p.Source] {
				return fmt.Errorf("mutants must be distinct")
			}
			mutants[p.Source] = true
			if err := p.Check(); err != nil {
				return err
			}
			if !langs[p.Language] {
				return fmt.Errorf("mutant language has no reference")
			}
		}
		if len(b.Cases) < 2 || len(b.Cases) > 20 {
			return fmt.Errorf("2..20 hand-derived anchor cases required")
		}
		visible, hidden := false, false
		names := map[string]bool{}
		for _, c := range b.Cases {
			if c.Name == "" || names[c.Name] || len(c.Input) > 16000 || len(c.Expected) > 64000 || !b.Comparator.Equal(c.Expected, c.Expected) {
				return fmt.Errorf("invalid anchor case")
			}
			names[c.Name] = true
			visible = visible || !c.Hidden
			hidden = hidden || c.Hidden
		}
		if !visible || !hidden {
			return fmt.Errorf("visible and hidden anchor cases required")
		}
		return nil
	}
	if !slices.Contains([]string{"single_choice", "multiple_select", "numeric"}, b.Format) {
		return fmt.Errorf("invalid format")
	}
	if len(b.References) > 0 || len(b.Templates) > 0 || b.Generator != nil || len(b.Cases) > 0 || len(b.Mutants) > 0 {
		return fmt.Errorf("answer formats cannot contain coding assets")
	}
	if b.Format == "numeric" {
		if b.Comparator.Kind != "numeric" || !b.Comparator.Equal(b.Answer, b.Answer) {
			return fmt.Errorf("invalid numeric answer")
		}
		return nil
	}
	if b.Comparator.Kind != "json" {
		return fmt.Errorf("choice answers use JSON comparator")
	}
	if len(b.Options) < 3 {
		return fmt.Errorf("at least three options required")
	}
	ids := map[string]bool{}
	for _, o := range b.Options {
		if o.ID == "" || ids[o.ID] || o.Text == "" || o.Rationale == "" {
			return fmt.Errorf("invalid option")
		}
		ids[o.ID] = true
	}
	var answers []string
	if b.Format == "single_choice" {
		var a string
		if json.Unmarshal([]byte(b.Answer), &a) != nil {
			return fmt.Errorf("answer must be a JSON option ID string")
		}
		answers = []string{a}
	} else {
		if json.Unmarshal([]byte(b.Answer), &answers) != nil || len(answers) == 0 {
			return fmt.Errorf("answer must be a JSON array of option IDs")
		}
	}
	seen := map[string]bool{}
	for _, a := range answers {
		if !ids[a] || seen[a] {
			return fmt.Errorf("answer references invalid or duplicate option")
		}
		seen[a] = true
	}
	return nil
}
