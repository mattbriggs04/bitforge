package authoring

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Interview is an ordered collection of independently validated questions.
// Paths are relative to the manifest, allowing the entire collection to move.
type Interview struct {
	SchemaVersion int      `json:"schemaVersion"`
	Slug          string   `json:"slug"`
	Title         string   `json:"title"`
	Category      string   `json:"category"`
	Description   string   `json:"description"`
	Questions     []string `json:"questions"`
}

func (i Interview) Check() error {
	if i.SchemaVersion != 1 || !slugPattern.MatchString(i.Slug) || len(i.Slug) > 100 {
		return fmt.Errorf("invalid interview schemaVersion or slug")
	}
	if strings.TrimSpace(i.Title) == "" || strings.TrimSpace(i.Category) == "" || strings.TrimSpace(i.Description) == "" {
		return fmt.Errorf("interview title, category and description required")
	}
	if len(i.Questions) < 1 || len(i.Questions) > 20 {
		return fmt.Errorf("interview requires 1..20 questions")
	}
	seen := map[string]bool{}
	for _, path := range i.Questions {
		if !fs.ValidPath(path) || path == "." || seen[path] {
			return fmt.Errorf("invalid or duplicate question path %q", path)
		}
		seen[path] = true
	}
	return nil
}
func LoadInterview(path string) (Interview, []Bundle, error) {
	var interview Interview
	data, err := os.ReadFile(path)
	if err != nil {
		return interview, nil, err
	}
	if err = Decode(data, &interview); err != nil {
		return interview, nil, err
	}
	if err = interview.Check(); err != nil {
		return interview, nil, err
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return interview, nil, err
	}
	defer root.Close()
	var bundles []Bundle
	slugs := map[string]bool{}
	for _, name := range interview.Questions {
		file, err := root.Open(name)
		if err != nil {
			return interview, nil, err
		}
		data, err := io.ReadAll(io.LimitReader(file, 2*1024*1024+1))
		file.Close()
		if err != nil {
			return interview, nil, err
		}
		if len(data) > 2*1024*1024 {
			return interview, nil, fmt.Errorf("question %s exceeds 2 MiB", name)
		}
		var b Bundle
		if err = Decode(data, &b); err != nil {
			return interview, nil, fmt.Errorf("%s: %w", name, err)
		}
		if err = b.Check(); err != nil {
			return interview, nil, fmt.Errorf("%s: %w", name, err)
		}
		if slugs[b.Slug] {
			return interview, nil, fmt.Errorf("duplicate question slug %q", b.Slug)
		}
		slugs[b.Slug] = true
		bundles = append(bundles, b)
	}
	return interview, bundles, nil
}
