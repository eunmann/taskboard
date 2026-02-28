package task

import (
	"testing"

	"github.com/eunmann/taskboard/internal/config"
)

const (
	fieldTitle = "title"
	fieldRefs  = "refs"
)

func TestValidateNoWarnings(t *testing.T) {
	task := &Task{
		Title:    "Valid task",
		ID:       "test",
		FileName: "test.yaml",
		Fields: map[string]string{
			"status": "open",
		},
		Created: getTestModTime(),
		Updated: getTestModTime(),
	}

	warnings := Validate(task, testColumns())
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestValidateEmptyTitle(t *testing.T) {
	task := &Task{
		Title:   "",
		ID:      "test",
		Fields:  map[string]string{"status": "open"},
		Created: getTestModTime(),
		Updated: getTestModTime(),
	}

	warnings := Validate(task, testColumns())

	found := false

	for _, w := range warnings {
		if w.Field == fieldTitle {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected warning for empty title")
	}
}

func TestValidateUnknownColumn(t *testing.T) {
	task := &Task{
		Title:   "Test",
		ID:      "test",
		Created: getTestModTime(),
		Updated: getTestModTime(),
		Fields: map[string]string{
			"nonexistent": "value",
		},
	}

	warnings := Validate(task, testColumns())

	found := false

	for _, w := range warnings {
		if w.Field == "nonexistent" {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected warning for unknown column")
	}
}

func TestValidateInvalidValue(t *testing.T) {
	task := &Task{
		Title:   "Test",
		ID:      "test",
		Created: getTestModTime(),
		Updated: getTestModTime(),
		Fields: map[string]string{
			"status": "invalid",
		},
	}

	warnings := Validate(task, testColumns())

	found := false

	for _, w := range warnings {
		if w.Field == "status" {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected warning for invalid status value")
	}
}

func TestValidateEmptyColumns(t *testing.T) {
	task := &Task{
		Title:   "Test",
		ID:      "test",
		Created: getTestModTime(),
		Updated: getTestModTime(),
		Fields: map[string]string{
			"anything": "value",
		},
	}

	warnings := Validate(task, map[string]config.Column{})

	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateIDMismatch(t *testing.T) {
	task := &Task{
		Title:    "Test",
		ID:       "abc123",
		FileName: "xyz789.yaml",
		Created:  getTestModTime(),
		Updated:  getTestModTime(),
		Fields:   map[string]string{},
	}

	warnings := Validate(task, testColumns())

	found := false

	for _, w := range warnings {
		if w.Field == "id" {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected warning for id/filename mismatch")
	}
}

func TestValidateMultipleParents(t *testing.T) {
	task := &Task{
		Title:   "Test",
		ID:      "test",
		Created: getTestModTime(),
		Updated: getTestModTime(),
		Fields:  map[string]string{},
		Refs: []Ref{
			{Type: RefParent, ID: "abc123"},
			{Type: RefParent, ID: "def789"},
		},
	}

	warnings := Validate(task, testColumns())

	found := false

	for _, w := range warnings {
		if w.Field == fieldRefs {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected warning for multiple parent refs")
	}
}

func TestValidateMissingTimestamps(t *testing.T) {
	task := &Task{
		Title:  "Test",
		ID:     "test",
		Fields: map[string]string{},
	}

	warnings := Validate(task, testColumns())

	createdFound := false
	updatedFound := false

	for _, w := range warnings {
		if w.Field == "created" {
			createdFound = true
		}

		if w.Field == "updated" {
			updatedFound = true
		}
	}

	if !createdFound {
		t.Error("expected warning for missing created")
	}

	if !updatedFound {
		t.Error("expected warning for missing updated")
	}
}

func TestValidateSkippedRefs(t *testing.T) {
	task := &Task{
		Title:       "Test",
		ID:          "test",
		Created:     getTestModTime(),
		Updated:     getTestModTime(),
		Fields:      map[string]string{},
		SkippedRefs: 2,
	}

	warnings := Validate(task, testColumns())

	found := false

	for _, w := range warnings {
		if w.Field == fieldRefs {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected warning for skipped refs")
	}
}

func TestValidateDanglingRefs(t *testing.T) {
	task := &Task{
		Title:   "Test",
		ID:      "test",
		Created: getTestModTime(),
		Updated: getTestModTime(),
		Fields:  map[string]string{},
		Refs: []Ref{
			{Type: RefBlockedBy, ID: "exists"},
			{Type: RefRelatesTo, ID: "missing"},
		},
	}

	allIDs := map[string]bool{"test": true, "exists": true}
	warnings := ValidateDanglingRefs(task, allIDs)

	if len(warnings) != 1 {
		t.Fatalf("expected 1 dangling ref warning, got %d", len(warnings))
	}

	if warnings[0].Field != fieldRefs {
		t.Errorf("warning field = %q, want %q", warnings[0].Field, fieldRefs)
	}
}

func TestValidateSlugFilenameMatches(t *testing.T) {
	task := &Task{
		Title:    "Setup CI pipeline",
		ID:       "Sd2k7x",
		FileName: "Sd2k7x-setup-ci-pipeline.yaml",
		Fields:   map[string]string{"status": "open"},
		Created:  getTestModTime(),
		Updated:  getTestModTime(),
	}

	warnings := Validate(task, testColumns())

	for _, w := range warnings {
		if w.Field == "id" {
			t.Errorf("unexpected id warning for matching slugged filename: %v", w)
		}
	}
}

func TestValidateSlugFilenameMismatch(t *testing.T) {
	task := &Task{
		Title:    "Some task",
		ID:       "abc123",
		FileName: "xyz789-some-task.yaml",
		Fields:   map[string]string{"status": "open"},
		Created:  getTestModTime(),
		Updated:  getTestModTime(),
	}

	warnings := Validate(task, testColumns())

	found := false

	for _, w := range warnings {
		if w.Field == "id" {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected warning for slugged filename with mismatched id")
	}
}

func TestWarningString(t *testing.T) {
	w := Warning{Field: "status", Message: "invalid value"}
	got := w.String()
	want := "status: invalid value"

	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
