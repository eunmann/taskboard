package index

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	writeFile(t, dir, "config.yaml", `
project: Test
columns:
  status:
    order: 1
    values:
      - name: open
        color: green
      - name: closed
        color: red
`)

	writeFile(t, dir, "abc12345.yaml", `
title: First task
status: open
`)

	writeFile(t, dir, "def67890.yaml", `
title: Second task
status: closed
`)

	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestNew(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tasks := idx.List()
	if len(tasks) != 2 {
		t.Fatalf("List() length = %d, want 2", len(tasks))
	}

	if idx.Version() == 0 {
		t.Error("Version() = 0, want > 0")
	}
}

func TestNewMissingConfig(t *testing.T) {
	_, err := New(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestGet(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	task := idx.Get("abc12345")
	if task == nil {
		t.Fatal("Get(abc12345) returned nil")
	}

	if task.Title != "First task" {
		t.Errorf("Title = %q, want %q", task.Title, "First task")
	}

	if idx.Get("nonexistent") != nil {
		t.Error("Get(nonexistent) should return nil")
	}
}

func TestDir(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if idx.Dir() != dir {
		t.Errorf("Dir() = %q, want %q", idx.Dir(), dir)
	}
}

func TestConfig(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cfg := idx.Config()
	if cfg.Project != "Test" {
		t.Errorf("Config().Project = %q, want %q", cfg.Project, "Test")
	}
}

func TestReload(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	v1 := idx.Version()

	writeFile(t, dir, "ghi24680.yaml", "title: Third\nstatus: open\n")

	if err := idx.Reload(); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	if len(idx.List()) != 3 {
		t.Errorf("List() length = %d, want 3", len(idx.List()))
	}

	if idx.Version() <= v1 {
		t.Error("Version not incremented after Reload")
	}
}

func TestReloadFile(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	v1 := idx.Version()

	path := filepath.Join(dir, "abc12345.yaml")
	writeFile(t, dir, "abc12345.yaml", "title: Updated\nstatus: closed\n")

	idx.ReloadFile(path)

	task := idx.Get("abc12345")
	if task == nil {
		t.Fatal("task disappeared after ReloadFile")
	}

	if task.Title != "Updated" {
		t.Errorf("Title = %q, want %q", task.Title, "Updated")
	}

	if idx.Version() <= v1 {
		t.Error("Version not incremented after ReloadFile")
	}
}

func TestReloadFileRemoved(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	path := filepath.Join(dir, "abc12345.yaml")

	if err := os.Remove(path); err != nil {
		t.Fatalf("remove file: %v", err)
	}

	idx.ReloadFile(path)

	if idx.Get("abc12345") != nil {
		t.Error("removed task still present")
	}

	if len(idx.List()) != 1 {
		t.Errorf("List() length = %d, want 1", len(idx.List()))
	}
}

func TestListSorted(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tasks := idx.List()
	if len(tasks) < 2 {
		t.Fatal("need at least 2 tasks")
	}

	if tasks[0].FileName >= tasks[1].FileName {
		t.Errorf("tasks not sorted: %q >= %q", tasks[0].FileName, tasks[1].FileName)
	}
}

func TestSkipsNonYAMLFiles(t *testing.T) {
	dir := setupTestDir(t)
	writeFile(t, dir, "notes.txt", "not a task")
	writeFile(t, dir, "README.md", "not a task")

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if len(idx.List()) != 2 {
		t.Errorf("List() length = %d, want 2 (should skip non-YAML)", len(idx.List()))
	}
}

func TestSkipsConfigFile(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, task := range idx.List() {
		if task.FileName == "config.yaml" {
			t.Error("config.yaml should not be loaded as a task")
		}
	}
}

func TestInvalidYAMLLoadsWithWarning(t *testing.T) {
	dir := setupTestDir(t)
	writeFile(t, dir, "broken.yaml", "{{{invalid yaml")

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	task := idx.Get("broken")
	if task == nil {
		t.Fatal("invalid YAML task should still be loaded with warnings")
	}

	if len(task.Warnings) == 0 {
		t.Error("invalid YAML task should have warnings")
	}
}

func TestIsTaskFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"task.yaml", true},
		{"task.yml", true},
		{"TASK.YAML", true},
		{"TASK.YML", true},
		{"readme.md", false},
		{"notes.txt", false},
		{"config.json", false},
		{"noext", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTaskFile(tt.name); got != tt.want {
				t.Errorf("isTaskFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestDanglingRefWarnings(t *testing.T) {
	dir := setupTestDir(t)
	writeFile(t, dir, "reftest.yaml", `
title: Ref test
status: open
refs:
  - type: blocked-by
    id: nonexistent
`)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	task := idx.Get("reftest")
	if task == nil {
		t.Fatal("reftest task not found")
	}

	hasDanglingWarning := false

	for _, w := range task.Warnings {
		if w.Field == "refs" && w.Message == "reference to unknown task 'nonexistent'" {
			hasDanglingWarning = true

			break
		}
	}

	if !hasDanglingWarning {
		t.Errorf("expected dangling ref warning, got: %v", task.Warnings)
	}
}
