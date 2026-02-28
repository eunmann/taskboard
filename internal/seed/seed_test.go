package seed

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()

	n, err := Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if n != 6 {
		t.Errorf("Run() wrote %d files, want 6", n)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	if len(entries) != 6 {
		t.Errorf("dir has %d files, want 6", len(entries))
	}
}

func TestRunSkipsExisting(t *testing.T) {
	dir := t.TempDir()

	existing := filepath.Join(dir, "Sd2k7x-setup-ci-pipeline.yaml")
	original := []byte("original content")

	if err := os.WriteFile(existing, original, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	n, err := Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if n != 5 {
		t.Errorf("Run() wrote %d files, want 5", n)
	}

	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if string(got) != string(original) {
		t.Error("Run() overwrote existing file")
	}
}

func TestRunIdempotent(t *testing.T) {
	dir := t.TempDir()

	if _, err := Run(dir); err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	n, err := Run(dir)
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	if n != 0 {
		t.Errorf("second Run() wrote %d files, want 0", n)
	}
}

func TestRunFilesAreValidYAML(t *testing.T) {
	dir := t.TempDir()

	if _, err := Run(dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s) error: %v", e.Name(), err)
		}

		var doc map[string]any
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Errorf("%s: invalid YAML: %v", e.Name(), err)
		}
	}
}

func TestRunBadDirectory(t *testing.T) {
	_, err := Run("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Run() with bad directory returned nil error")
	}
}

func TestSampleCount(t *testing.T) {
	entries, err := fs.ReadDir(samplesFS, "samples")
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	if len(entries) != 6 {
		t.Errorf("embedded FS has %d samples, want 6", len(entries))
	}
}
