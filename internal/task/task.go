// Package task handles parsing YAML task files and task data structures.
package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eunmann/taskboard/internal/config"
	"gopkg.in/yaml.v3"
)

// Task represents a single task parsed from a YAML file.
type Task struct {
	ID          string
	Title       string
	Description string
	Tags        []string
	Refs        []Ref
	Fields      map[string]string
	Created     time.Time
	Updated     time.Time
	Warnings    []Warning
	FileName    string
	SkippedRefs int
}

// Parse parses task YAML data and validates against the given columns.
// The modTime is used as a fallback for missing created/updated timestamps.
func Parse(filename string, data []byte, columns map[string]config.Column, modTime time.Time) (*Task, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return parseError(filename, modTime, err), nil
	}

	t := &Task{
		Fields:   make(map[string]string),
		FileName: filename,
	}

	t.ID = extractString(raw, "id")
	if t.ID == "" {
		t.ID = idFromFilename(filename)
	}

	t.Title = extractString(raw, "title")
	t.Description = extractString(raw, "description")
	t.Tags = extractStringSlice(raw, "tags")
	t.Refs, t.SkippedRefs = extractRefs(raw)
	t.Created = extractTime(raw, "created")
	t.Updated = extractTime(raw, "updated")

	if t.Created.IsZero() && !modTime.IsZero() {
		t.Created = modTime
	}

	if t.Updated.IsZero() && !modTime.IsZero() {
		t.Updated = modTime
	}

	for key, val := range raw {
		if isKnownField(key) {
			continue
		}

		t.Fields[key] = fmt.Sprintf("%v", val)
	}

	t.Warnings = Validate(t, columns)

	return t, nil
}

// parseError creates a degraded task for files with YAML parse errors.
// The task appears in the UI with its filename as title and a warning badge.
func parseError(filename string, modTime time.Time, parseErr error) *Task {
	id := idFromFilename(filename)

	return &Task{
		ID:       id,
		Title:    "",
		Fields:   make(map[string]string),
		FileName: filename,
		Created:  modTime,
		Updated:  modTime,
		Warnings: []Warning{{
			Field:   "yaml",
			Message: fmt.Sprintf("failed to parse file: %v", parseErr),
		}},
	}
}

// ParseFile reads and parses a task file from disk.
func ParseFile(path string, columns map[string]config.Column) (*Task, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat task file: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task file: %w", err)
	}

	return Parse(filepath.Base(path), data, columns, info.ModTime())
}

func isKnownField(name string) bool {
	switch name {
	case "id", "title", "description", "tags", "refs", "created", "updated":
		return true
	default:
		return false
	}
}

func idFromFilename(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	if idx := strings.IndexByte(name, '-'); idx > 0 {
		return name[:idx]
	}

	return name
}

func extractString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}

	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}

	return s
}

func extractStringSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}

	items, ok := v.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}

	return result
}

func extractRefs(m map[string]any) ([]Ref, int) {
	v, ok := m["refs"]
	if !ok {
		return nil, 0
	}

	items, ok := v.([]any)
	if !ok {
		return nil, 0
	}

	var refs []Ref

	skipped := 0

	for _, item := range items {
		ref, ok := item.(map[string]any)
		if !ok {
			skipped++

			continue
		}

		refType, hasType := ref["type"]
		refID, hasID := ref["id"]

		if !hasType || !hasID {
			skipped++

			continue
		}

		r := Ref{
			Type: fmt.Sprintf("%v", refType),
			ID:   fmt.Sprintf("%v", refID),
		}
		refs = append(refs, r)
	}

	return refs, skipped
}

func extractTime(m map[string]any, key string) time.Time {
	v, ok := m[key]
	if !ok {
		return time.Time{}
	}

	switch val := v.(type) {
	case time.Time:
		return val
	case string:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return time.Time{}
		}

		return t
	default:
		return time.Time{}
	}
}
