// Package config handles parsing and validation of .tasks/config.yaml.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigFile is the name of the configuration file within a tasks directory.
const ConfigFile = "config.yaml"

const (
	// priorityColumnOrder is the display order for the priority column.
	priorityColumnOrder = 2
	// typeColumnOrder is the display order for the type column.
	typeColumnOrder = 3
	// dirPermissions is the permission mode for the tasks directory.
	dirPermissions = 0o750
	// filePermissions is the permission mode for config files.
	filePermissions = 0o600
)

// Validation errors.
var (
	ErrMissingProject = errors.New("project name is required")
	ErrNoColumns      = errors.New("at least one column is required")
	ErrReservedField  = errors.New("column name conflicts with reserved field")
	ErrEmptyValueName = errors.New("column value has empty name")
	ErrDuplicateValue = errors.New("column has duplicate value")
)

// Config holds the project configuration including column definitions.
type Config struct {
	Project string            `yaml:"project"`
	Columns map[string]Column `yaml:"columns"`
}

// Column defines an enumerated field with ordered values.
type Column struct {
	Order  int     `yaml:"order"`
	Values []Value `yaml:"values"`
}

// Value defines one allowed value for a column with its display color.
type Value struct {
	Name  string `yaml:"name"`
	Color string `yaml:"color"`
}

// Load reads and validates config from the given tasks directory.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, ConfigFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// ValueNames returns the list of valid value names for a column.
func (c Column) ValueNames() []string {
	names := make([]string, len(c.Values))
	for i, v := range c.Values {
		names[i] = v.Name
	}

	return names
}

// ColorFor returns the color for a given value name, or empty string if not found.
func (c Column) ColorFor(name string) string {
	for _, v := range c.Values {
		if v.Name == name {
			return v.Color
		}
	}

	return ""
}

func reservedFields() []string {
	return []string{"id", "title", "description", "tags", "refs", "created", "updated"}
}

func (cfg *Config) validate() error {
	if cfg.Project == "" {
		return ErrMissingProject
	}

	if len(cfg.Columns) == 0 {
		return ErrNoColumns
	}

	for name := range cfg.Columns {
		if err := cfg.validateColumn(name); err != nil {
			return err
		}
	}

	return nil
}

func (cfg *Config) validateColumn(name string) error {
	lower := strings.ToLower(name)
	if slices.Contains(reservedFields(), lower) {
		return fmt.Errorf("column %q: %w", name, ErrReservedField)
	}

	col := cfg.Columns[name]
	if len(col.Values) == 0 {
		return fmt.Errorf("column %q: %w", name, ErrNoColumns)
	}

	seen := make(map[string]bool)

	for _, v := range col.Values {
		if v.Name == "" {
			return fmt.Errorf("column %q: %w", name, ErrEmptyValueName)
		}

		if seen[v.Name] {
			return fmt.Errorf("column %q value %q: %w", name, v.Name, ErrDuplicateValue)
		}

		seen[v.Name] = true
	}

	return nil
}

// DefaultConfig returns a starter configuration matching the spec §6.5.
func DefaultConfig() *Config {
	return &Config{
		Project: "Taskboard",
		Columns: map[string]Column{
			"status": {
				Order: 1,
				Values: []Value{
					{Name: "open", Color: "#22c55e"},
					{Name: "in-progress", Color: "#3b82f6"},
					{Name: "review", Color: "#f59e0b"},
					{Name: "done", Color: "#6b7280"},
					{Name: "cancelled", Color: "#ef4444"},
				},
			},
			"priority": {
				Order: priorityColumnOrder,
				Values: []Value{
					{Name: "critical", Color: "#dc2626"},
					{Name: "high", Color: "#f97316"},
					{Name: "medium", Color: "#eab308"},
					{Name: "low", Color: "#6b7280"},
				},
			},
			"type": {
				Order: typeColumnOrder,
				Values: []Value{
					{Name: "task"},
					{Name: "bug", Color: "#ef4444"},
					{Name: "feature", Color: "#8b5cf6"},
					{Name: "epic", Color: "#06b6d4"},
					{Name: "spike", Color: "#64748b"},
				},
			},
		},
	}
}

// WriteDefault creates the tasks directory and writes the default config file.
func WriteDefault(dir string) error {
	if err := os.MkdirAll(dir, dirPermissions); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	cfg := DefaultConfig()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path := filepath.Join(dir, ConfigFile)

	if err := os.WriteFile(path, data, filePermissions); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
