package authoring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInterviewManifestAndConfinement(t *testing.T) {
	manifest, bundles, err := LoadInterview("../../../generation/interviews/binary-classifiers/interview.json")
	if err != nil || len(bundles) != 4 {
		t.Fatal(manifest, err)
	}
	if bundles[0].Format != "single_choice" || bundles[1].Format != "single_choice" || bundles[2].Format != "single_choice" || bundles[3].Format != "coding" {
		t.Fatal("unexpected interview composition")
	}
	for _, paths := range [][]string{{"../outside.json"}, {"/absolute.json"}, {"same.json", "same.json"}, {}} {
		bad := manifest
		bad.Questions = paths
		if bad.Check() == nil {
			t.Fatalf("invalid paths accepted: %v", paths)
		}
	}
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err = os.WriteFile(outside, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(outside, filepath.Join(dir, "escape.json")); err != nil {
		t.Fatal(err)
	}
	manifest.Questions = []string{"escape.json"}
	data, _ := json.Marshal(manifest)
	path := filepath.Join(dir, "interview.json")
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err = LoadInterview(path); err == nil {
		t.Fatal("symlink escaped interview directory")
	}
}
