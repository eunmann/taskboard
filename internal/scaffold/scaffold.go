// Package scaffold copies Claude Code slash commands and CLAUDE.md instructions
// into a project directory so any repo can use the taskboard workflow.
package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	fileMode    = 0o644
	dirMode     = 0o755
	commandsDir = "commands"
	claudeMDDir = ".claude/commands"
	claudeMD    = "CLAUDE.md"
	sentinel    = "<!-- taskboard:begin"
)

//go:embed commands/*.md
var commandsFS embed.FS

//go:embed claudemd.txt
var claudeMDSnippet string

// Result reports what Write did.
type Result struct {
	CommandsWritten  int
	CommandsSkipped  int
	ClaudeMDAppended bool
}

// Write copies embedded slash commands into projectDir/.claude/commands/
// and appends the taskboard CLAUDE.md snippet if not already present.
func Write(projectDir string) (*Result, error) {
	result := &Result{}

	if err := writeCommands(projectDir, result); err != nil {
		return result, fmt.Errorf("write commands: %w", err)
	}

	if err := appendClaudeMD(projectDir, result); err != nil {
		return result, fmt.Errorf("append CLAUDE.md: %w", err)
	}

	return result, nil
}

func writeCommands(projectDir string, result *Result) error {
	destDir := filepath.Join(projectDir, claudeMDDir)

	if err := os.MkdirAll(destDir, dirMode); err != nil {
		return fmt.Errorf("create %s: %w", destDir, err)
	}

	entries, err := fs.ReadDir(commandsFS, commandsDir)
	if err != nil {
		return fmt.Errorf("read embedded commands: %w", err)
	}

	for _, e := range entries {
		dest := filepath.Join(destDir, e.Name())

		if _, err := os.Stat(dest); err == nil {
			result.CommandsSkipped++

			continue
		}

		data, err := commandsFS.ReadFile(commandsDir + "/" + e.Name())
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", e.Name(), err)
		}

		if err := os.WriteFile(dest, data, fileMode); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}

		result.CommandsWritten++
	}

	return nil
}

func appendClaudeMD(projectDir string, result *Result) error {
	path := filepath.Join(projectDir, claudeMD)

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", claudeMD, err)
	}

	if strings.Contains(string(existing), sentinel) {
		return nil
	}

	content := claudeMDSnippet
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		content = "\n" + content
	}

	if err := appendToFile(path, content); err != nil {
		return fmt.Errorf("append %s: %w", claudeMD, err)
	}

	result.ClaudeMDAppended = true

	return nil
}

func appendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileMode)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	_, writeErr := f.WriteString(content)

	if closeErr := f.Close(); closeErr != nil && writeErr == nil {
		return fmt.Errorf("close: %w", closeErr)
	}

	if writeErr != nil {
		return fmt.Errorf("write: %w", writeErr)
	}

	return nil
}
