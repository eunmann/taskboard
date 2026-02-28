package task

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eunmann/taskboard/internal/config"
)

const testTaskID = "k7x2m9"

func getTestModTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

func testColumns() map[string]config.Column {
	return map[string]config.Column{
		"status": {
			Order: 1,
			Values: []config.Value{
				{Name: "open", Color: "#22c55e"},
				{Name: "closed", Color: "#ef4444"},
			},
		},
		"priority": {
			Order: 2,
			Values: []config.Value{
				{Name: "low", Color: "#6b7280"},
				{Name: "high", Color: "#f97316"},
			},
		},
	}
}

func TestParse(t *testing.T) {
	data := []byte(`
title: Fix the bug
description: |
  This is a detailed description.
status: open
priority: high
tags: [backend, urgent]
created: 2025-01-01T00:00:00Z
updated: 2025-01-02T00:00:00Z
`)

	task, err := Parse(testTaskID+".yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if task.ID != testTaskID {
		t.Errorf("ID = %q, want %q", task.ID, "k7x2m9")
	}

	if task.Title != "Fix the bug" {
		t.Errorf("Title = %q, want %q", task.Title, "Fix the bug")
	}

	if task.Fields["status"] != "open" {
		t.Errorf("Fields[status] = %q, want %q", task.Fields["status"], "open")
	}

	if task.Fields["priority"] != "high" {
		t.Errorf("Fields[priority] = %q, want %q", task.Fields["priority"], "high")
	}

	if len(task.Tags) != 2 {
		t.Fatalf("Tags length = %d, want 2", len(task.Tags))
	}

	if task.Tags[0] != "backend" || task.Tags[1] != "urgent" {
		t.Errorf("Tags = %v, want [backend urgent]", task.Tags)
	}

	if len(task.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none", task.Warnings)
	}
}

func TestParseWithExplicitID(t *testing.T) {
	data := []byte(`
id: k7x2m9
title: Task with ID
status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
`)

	task, err := Parse(testTaskID+".yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if task.ID != testTaskID {
		t.Errorf("ID = %q, want %q", task.ID, "k7x2m9")
	}
}

func TestParseWithExplicitIDSlugFilename(t *testing.T) {
	data := []byte(`
id: k7x2m9
title: Task with slug
status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
`)

	task, err := Parse(testTaskID+"-task-with-slug.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if task.ID != testTaskID {
		t.Errorf("ID = %q, want %q", task.ID, testTaskID)
	}

	for _, w := range task.Warnings {
		if w.Field == "id" {
			t.Errorf("unexpected id warning for matching slugged filename: %v", w)
		}
	}
}

func TestParseIDMismatchWarning(t *testing.T) {
	data := []byte(`
id: abc123
title: Mismatched
status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
`)

	task, err := Parse("xyz789.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if task.ID != "abc123" {
		t.Errorf("ID = %q, want %q", task.ID, "abc123")
	}

	hasWarning := false

	for _, w := range task.Warnings {
		if w.Field == "id" {
			hasWarning = true

			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for id/filename mismatch")
	}
}

func TestParseIDFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		wantID   string
	}{
		{testTaskID + ".yaml", "k7x2m9"},
		{"abc123-fix-bug.yaml", "abc123"},
		{"simple.yaml", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			task, err := Parse(tt.filename, []byte("title: test\nstatus: open\ncreated: 2025-01-01T00:00:00Z\nupdated: 2025-01-01T00:00:00Z"), testColumns(), getTestModTime())
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if task.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", task.ID, tt.wantID)
			}
		})
	}
}

func TestParseInvalidYAML(t *testing.T) {
	task, err := Parse("bad.yaml", []byte("{{{invalid"), testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() should not return error for invalid YAML, got: %v", err)
	}

	if task.ID != "bad" {
		t.Errorf("ID = %q, want %q (from filename)", task.ID, "bad")
	}

	if len(task.Warnings) == 0 {
		t.Error("expected warnings for invalid YAML")
	}

	hasYAMLWarning := false

	for _, w := range task.Warnings {
		if w.Field == "yaml" {
			hasYAMLWarning = true

			break
		}
	}

	if !hasYAMLWarning {
		t.Error("expected YAML parse warning")
	}
}

func TestParseUnknownColumn(t *testing.T) {
	data := []byte(`
title: Test
unknown_field: some-value
status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
`)

	task, err := Parse("test.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if task.Fields["unknown_field"] != "some-value" {
		t.Errorf("unknown field not stored: %v", task.Fields)
	}

	hasWarning := false

	for _, w := range task.Warnings {
		if w.Field == "unknown_field" {
			hasWarning = true

			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for unknown column")
	}
}

func TestParseInvalidEnumValue(t *testing.T) {
	data := []byte(`
title: Test
status: invalid-value
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
`)

	task, err := Parse("test.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if task.Fields["status"] != "invalid-value" {
		t.Errorf("invalid value not stored raw: %q", task.Fields["status"])
	}

	hasWarning := false

	for _, w := range task.Warnings {
		if w.Field == "status" {
			hasWarning = true

			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for invalid enum value")
	}
}

func TestParseRefs(t *testing.T) {
	data := []byte(`
title: Test
status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
refs:
  - type: blocked-by
    id: abc123
  - type: relates-to
    id: def789
`)

	task, err := Parse("test.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(task.Refs) != 2 {
		t.Fatalf("Refs length = %d, want 2", len(task.Refs))
	}

	if task.Refs[0].Type != "blocked-by" || task.Refs[0].ID != "abc123" {
		t.Errorf("Refs[0] = %+v, want blocked-by/abc123", task.Refs[0])
	}
}

func TestParseRefsSkipsMissingFields(t *testing.T) {
	data := []byte(`
title: Test
status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
refs:
  - type: parent
  - id: abc123
  - type: blocked-by
    id: def789
`)

	task, err := Parse("test.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(task.Refs) != 1 {
		t.Errorf("Refs length = %d, want 1 (should skip entries with missing type/id)", len(task.Refs))
	}
}

func TestParseMissingTitle(t *testing.T) {
	data := []byte(`status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z`)

	task, err := Parse("test.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	hasWarning := false

	for _, w := range task.Warnings {
		if w.Field == fieldTitle {
			hasWarning = true

			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for missing title")
	}
}

func TestParseMissingTimestampsFallback(t *testing.T) {
	data := []byte(`
title: Test
status: open
`)

	task, err := Parse("test.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if task.Created != getTestModTime() {
		t.Errorf("Created = %v, want fallback to modTime %v", task.Created, getTestModTime())
	}

	if task.Updated != getTestModTime() {
		t.Errorf("Updated = %v, want fallback to modTime %v", task.Updated, getTestModTime())
	}
}

func TestParseMultipleParentsWarning(t *testing.T) {
	data := []byte(`
title: Test
status: open
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
refs:
  - type: parent
    id: abc123
  - type: parent
    id: def789
`)

	task, err := Parse("test.yaml", data, testColumns(), getTestModTime())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	hasWarning := false

	for _, w := range task.Warnings {
		if w.Field == fieldRefs {
			hasWarning = true

			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for multiple parent refs")
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, testTaskID+".yaml")

	err := os.WriteFile(path, []byte("title: From file\nstatus: open\ncreated: 2025-01-01T00:00:00Z\nupdated: 2025-01-01T00:00:00Z\n"), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}

	task, err := ParseFile(path, testColumns())
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if task.Title != "From file" {
		t.Errorf("Title = %q, want %q", task.Title, "From file")
	}

	if task.ID != testTaskID {
		t.Errorf("ID = %q, want %q", task.ID, "k7x2m9")
	}
}

func TestParseInvalidTimestampFormat(t *testing.T) {
	data := []byte(`
title: Test
status: open
created: "not-a-date"
updated: "also-not-a-date"
`)

	task, err := Parse("test.yaml", data, testColumns(), time.Time{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !task.Created.IsZero() {
		t.Error("Created should be zero for invalid format")
	}

	if !task.Updated.IsZero() {
		t.Error("Updated should be zero for invalid format")
	}
}

func TestParseNonStringTimestamp(t *testing.T) {
	data := []byte(`
title: Test
status: open
created: 12345
updated: true
`)

	task, err := Parse("test.yaml", data, testColumns(), time.Time{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !task.Created.IsZero() {
		t.Error("Created should be zero for non-string/non-time type")
	}

	if !task.Updated.IsZero() {
		t.Error("Updated should be zero for non-string/non-time type")
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path.yaml", testColumns())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
