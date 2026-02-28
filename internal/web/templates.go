package web

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// ErrTemplateNotFound is returned when a requested template does not exist.
var ErrTemplateNotFound = errors.New("template not found")

//go:embed templates/*
var templateFS embed.FS

// Renderer manages parsed HTML templates.
type Renderer struct {
	templates map[string]*template.Template
}

// NewRenderer creates a Renderer by parsing all templates.
func NewRenderer() (*Renderer, error) {
	r := &Renderer{
		templates: make(map[string]*template.Template),
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
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

// RenderError renders an error page.
func (r *Renderer) RenderError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Error</title></head>
<body style="font-family:sans-serif;max-width:600px;margin:40px auto;padding:20px">
<h1>Error %d</h1><p>%s</p>
<a href="/">Back to Home</a>
</body></html>`, statusCode, message)
	_, _ = w.Write([]byte(html))
}

func (r *Renderer) loadTemplates() error {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006 3:04PM")
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	}

	layoutBytes, err := templateFS.ReadFile("templates/shared/layouts/main.html")
	if err != nil {
		return fmt.Errorf("read layout: %w", err)
	}

	partialBytes, err := collectTemplateDir("templates/shared/partials")
	if err != nil {
		return fmt.Errorf("walk partials: %w", err)
	}

	componentBytes, err := collectTemplateDir("templates/shared/components")
	if err != nil {
		return fmt.Errorf("walk components: %w", err)
	}

	err = fs.WalkDir(templateFS, "templates/pages", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}

		pageBytes, readErr := templateFS.ReadFile(path)
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

func collectTemplateDir(dir string) ([]byte, error) {
	var result []byte

	err := fs.WalkDir(templateFS, dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}

		b, readErr := templateFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		result = append(result, b...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, err)
	}

	return result, nil
}
