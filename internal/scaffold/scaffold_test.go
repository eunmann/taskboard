package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	dir := t.TempDir()

	result, err := Write(dir)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if result.CommandsWritten != 3 {
		t.Errorf("CommandsWritten = %d, want 3", result.CommandsWritten)
	}

	if result.CommandsSkipped != 0 {
		t.Errorf("CommandsSkipped = %d, want 0", result.CommandsSkipped)
	}

	if !result.ClaudeMDAppended {
		t.Error("ClaudeMDAppended = false, want true")
	}

	cmdDir := filepath.Join(dir, ".claude", "commands")

	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("commands dir has %d files, want 3", len(entries))
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile(CLAUDE.md) error: %v", err)
	}

	if !strings.Contains(string(data), sentinel) {
		t.Error("CLAUDE.md missing sentinel marker")
	}
}

func TestWriteSkipsExisting(t *testing.T) {
	dir := t.TempDir()

	cmdDir := filepath.Join(dir, ".claude", "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}

	existing := filepath.Join(cmdDir, "plan-work.md")
	original := []byte("original content")

	if err := os.WriteFile(existing, original, 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	result, err := Write(dir)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if result.CommandsWritten != 2 {
		t.Errorf("CommandsWritten = %d, want 2", result.CommandsWritten)
	}

	if result.CommandsSkipped != 1 {
		t.Errorf("CommandsSkipped = %d, want 1", result.CommandsSkipped)
	}

	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if string(got) != string(original) {
		t.Error("Write() overwrote existing file")
	}
}

func TestWriteIdempotent(t *testing.T) {
	dir := t.TempDir()

	if _, err := Write(dir); err != nil {
		t.Fatalf("first Write() error: %v", err)
	}

	result, err := Write(dir)
	if err != nil {
		t.Fatalf("second Write() error: %v", err)
	}

	if result.CommandsWritten != 0 {
		t.Errorf("second Write() CommandsWritten = %d, want 0", result.CommandsWritten)
	}

	if result.CommandsSkipped != 3 {
		t.Errorf("second Write() CommandsSkipped = %d, want 3", result.CommandsSkipped)
	}

	if result.ClaudeMDAppended {
		t.Error("second Write() ClaudeMDAppended = true, want false")
	}
}

func TestWriteExistingClaudeMDAppend(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "CLAUDE.md")
	original := "# My Project\n\nExisting content.\n"

	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	result, err := Write(dir)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if !result.ClaudeMDAppended {
		t.Error("ClaudeMDAppended = false, want true")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	content := string(data)

	if !strings.HasPrefix(content, original) {
		t.Error("Write() did not preserve existing CLAUDE.md content")
	}

	if !strings.Contains(content, sentinel) {
		t.Error("CLAUDE.md missing sentinel marker after append")
	}
}

func TestWriteExistingClaudeMDWithMarker(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "CLAUDE.md")
	content := "# My Project\n\n<!-- taskboard:begin -->\nold snippet\n<!-- taskboard:end -->\n"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	result, err := Write(dir)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if result.ClaudeMDAppended {
		t.Error("ClaudeMDAppended = true, want false (sentinel present)")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if string(data) != content {
		t.Error("Write() modified CLAUDE.md despite sentinel being present")
	}
}

func TestWriteBadDirectory(t *testing.T) {
	_, err := Write("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Write() with bad directory returned nil error")
	}
}

func TestCommandCount(t *testing.T) {
	entries, err := fs.ReadDir(commandsFS, commandsDir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("embedded FS has %d commands, want 3", len(entries))
	}
}

func TestCommandsHaveFrontMatter(t *testing.T) {
	entries, err := fs.ReadDir(commandsFS, commandsDir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	for _, e := range entries {
		data, err := commandsFS.ReadFile(commandsDir + "/" + e.Name())
		if err != nil {
			t.Fatalf("ReadFile(%s) error: %v", e.Name(), err)
		}

		content := string(data)

		if !strings.HasPrefix(content, "---\n") {
			t.Errorf("%s: does not start with front matter delimiter", e.Name())
		}

		if !strings.Contains(content, "description:") {
			t.Errorf("%s: missing description in front matter", e.Name())
		}
	}
}

func TestClaudeMDSnippetContainsMarker(t *testing.T) {
	if !strings.Contains(claudeMDSnippet, sentinel) {
		t.Error("claudeMDSnippet missing sentinel marker")
	}
}
