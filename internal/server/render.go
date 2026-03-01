package server

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/web"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const (
	// dictPairSize is the number of elements per key-value pair in dict.
	dictPairSize = 2
	// hoursPerDay is used for relative time calculations.
	hoursPerDay = 24
)

// Renderer errors.
var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrDictArgs         = errors.New("dict requires even number of arguments")
	ErrDictKeyType      = errors.New("dict key must be string")
)

// ColumnInfo pairs a column name with its config for template iteration.
type ColumnInfo struct {
	Name   string
	Column config.Column
}

// Renderer manages parsed HTML templates.
type Renderer struct {
	templates map[string]*template.Template
	md        goldmark.Markdown
}

// NewRenderer creates a Renderer by parsing all embedded templates.
func NewRenderer() (*Renderer, error) {
	r := &Renderer{
		templates: make(map[string]*template.Template),
		md:        goldmark.New(goldmark.WithExtensions(extension.GFM)),
	}

	if err := r.loadTemplates(); err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}

	return r, nil
}

// Render renders a full page with layout to the response writer.
func (r *Renderer) Render(w http.ResponseWriter, name string, data any) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}

	var buf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		return fmt.Errorf("execute template %s: %w", name, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

// RenderPartial renders a named block without the layout.
func (r *Renderer) RenderPartial(w http.ResponseWriter, templateName, blockName string, data any) error {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTemplateNotFound, templateName)
	}

	var buf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buf, blockName, data); err != nil {
		return fmt.Errorf("execute block %s/%s: %w", templateName, blockName, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("write partial: %w", err)
	}

	return nil
}

func (r *Renderer) renderMarkdown(s string) template.HTML {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(s), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(s)) //nolint:gosec // HTML-escaped fallback is safe
	}

	return template.HTML(buf.String()) //nolint:gosec // markdown output is trusted content from task YAML files
}

func (r *Renderer) loadTemplates() error {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006 3:04 PM")
		},
		"relativeTime":        relativeTime,
		"lower":               strings.ToLower,
		"upper":               strings.ToUpper,
		"dict":                dictFunc,
		"hasPrefix":           strings.HasPrefix,
		"sortParam":           sortParam,
		"sortIndicator":       sortIndicator,
		"inSlice":             slices.Contains[[]string, string],
		"queryParams":         queryParams,
		"queryParamsWithSort": queryParamsWithSort,
		"queryParamsWithout":  queryParamsWithout,
		"joinComma": func(s []string) string {
			return strings.Join(s, ",")
		},
		"markdown": r.renderMarkdown,
	}

	layoutBytes, err := web.TemplateFS.ReadFile("templates/layout.html")
	if err != nil {
		return fmt.Errorf("read layout: %w", err)
	}

	partialBytes, err := collectDir("templates/partials")
	if err != nil {
		return fmt.Errorf("collect partials: %w", err)
	}

	componentBytes, err := collectDir("templates/components")
	if err != nil {
		return fmt.Errorf("collect components: %w", err)
	}

	err = fs.WalkDir(web.TemplateFS, "templates/pages", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}

		pageBytes, readErr := web.TemplateFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read page %s: %w", path, readErr)
		}

		combined := string(layoutBytes) + string(partialBytes) + string(componentBytes) + string(pageBytes)
		name := strings.TrimSuffix(filepath.Base(path), ".html")

		tmpl, parseErr := template.New("layout").Funcs(funcMap).Parse(combined)
		if parseErr != nil {
			return fmt.Errorf("parse template %s: %w", name, parseErr)
		}

		r.templates[name] = tmpl

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk pages: %w", err)
	}

	return nil
}

func collectDir(dir string) ([]byte, error) {
	var result []byte

	err := fs.WalkDir(web.TemplateFS, dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}

		b, readErr := web.TemplateFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		result = append(result, b...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect templates in %s: %w", dir, err)
	}

	return result, nil
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}

		return fmt.Sprintf("%d minutes ago", m)
	case d < time.Hour*hoursPerDay:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}

		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / hoursPerDay)
		if days == 1 {
			return "1 day ago"
		}

		return fmt.Sprintf("%d days ago", days)
	}
}

func dictFunc(pairs ...any) (map[string]any, error) {
	if len(pairs)%dictPairSize != 0 {
		return nil, ErrDictArgs
	}

	m := make(map[string]any, len(pairs)/dictPairSize)

	var key string

	for idx, p := range pairs {
		if idx%dictPairSize == 0 {
			var ok bool

			key, ok = p.(string)
			if !ok {
				return nil, fmt.Errorf("%w: got %T", ErrDictKeyType, p)
			}
		} else {
			m[key] = p
		}
	}

	return m, nil
}

// SortedColumns returns columns sorted by order for consistent display.
func SortedColumns(columns map[string]config.Column) []ColumnInfo {
	result := make([]ColumnInfo, 0, len(columns))
	for name, col := range columns {
		result = append(result, ColumnInfo{Name: name, Column: col})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Column.Order != result[j].Column.Order {
			return result[i].Column.Order < result[j].Column.Order
		}

		return result[i].Name < result[j].Name
	})

	return result
}
