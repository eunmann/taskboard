// Package web provides embedded static assets and templates for the taskboard UI.
package web

import "embed"

// StaticFS holds the embedded static assets (CSS, JS, vendor libraries).
//
//go:embed static/*
var StaticFS embed.FS

// TemplateFS holds the embedded HTML templates (layout, pages, partials, components).
//
//go:embed templates/*
var TemplateFS embed.FS
