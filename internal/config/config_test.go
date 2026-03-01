package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `
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

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Project != "Test" {
		t.Errorf("Project = %q, want %q", cfg.Project, "Test")
	}

	col, ok := cfg.Columns["status"]
	if !ok {
		t.Fatal("missing column 'status'")
	}

	if len(col.Values) != 2 {
		t.Fatalf("Values count = %d, want 2", len(col.Values))
	}

	if col.Values[0].Name != "open" {
		t.Errorf("Values[0].Name = %q, want %q", col.Values[0].Name, "open")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `{{{invalid`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestValidateMissingProject(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `
project: ""
columns:
  status:
    values:
      - name: open
        color: green
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing project name")
	}
}

func TestValidateNoColumns(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `
project: Test
columns: {}
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty columns")
	}
}

func TestValidateReservedField(t *testing.T) {
	reserved := []string{"id", "title", "description", "tags", "refs", "created", "updated"}
	for _, name := range reserved {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeYAML(t, dir, `
project: Test
columns:
  `+name+`:
    values:
      - name: foo
        color: blue
`)

			_, err := Load(dir)
			if err == nil {
				t.Fatalf("expected error for reserved field %q", name)
			}

			if !errors.Is(err, ErrReservedField) {
				t.Errorf("error = %v, want ErrReservedField", err)
			}
		})
	}
}

func TestValidateDuplicateValue(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `
project: Test
columns:
  status:
    values:
      - name: open
        color: green
      - name: open
        color: red
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for duplicate value name")
	}

	if !errors.Is(err, ErrDuplicateValue) {
		t.Errorf("error = %v, want ErrDuplicateValue", err)
	}
}

func TestValidateEmptyValueName(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `
project: Test
columns:
  status:
    values:
      - name: ""
        color: green
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty value name")
	}

	if !errors.Is(err, ErrEmptyValueName) {
		t.Errorf("error = %v, want ErrEmptyValueName", err)
	}
}

func TestValidateEmptyValues(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `
project: Test
columns:
  status:
    values: []
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty values list")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.validate(); err != nil {
		t.Fatalf("DefaultConfig() is invalid: %v", err)
	}

	if cfg.Project == "" {
		t.Error("DefaultConfig().Project is empty")
	}

	if len(cfg.Columns) < 2 {
		t.Errorf("DefaultConfig() has %d columns, want >= 2", len(cfg.Columns))
	}
}

func TestWriteDefaultAndLoad(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".taskboard")

	if err := WriteDefault(dir); err != nil {
		t.Fatalf("WriteDefault() error = %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() after WriteDefault error = %v", err)
	}

	if cfg.Project == "" {
		t.Error("loaded config has empty project name")
	}
}

func TestColumnValueNames(t *testing.T) {
	col := Column{
		Values: []Value{
			{Name: "a", Color: "red"},
			{Name: "b", Color: "blue"},
		},
	}

	names := col.ValueNames()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("ValueNames() = %v, want [a b]", names)
	}
}

func TestColumnColorFor(t *testing.T) {
	col := Column{
		Values: []Value{
			{Name: "open", Color: "green"},
			{Name: "closed", Color: "red"},
		},
	}

	if got := col.ColorFor("open"); got != "green" {
		t.Errorf("ColorFor(open) = %q, want %q", got, "green")
	}

	if got := col.ColorFor("missing"); got != "" {
		t.Errorf("ColorFor(missing) = %q, want empty", got)
	}
}

func writeYAML(t *testing.T, dir, content string) {
	t.Helper()

	path := filepath.Join(dir, ConfigFile)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test config: %v", err)
	}
}
