package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

// StaticHandler returns an HTTP handler that serves embedded static assets.
func StaticHandler() http.Handler {
	sub, _ := fs.Sub(staticFS, "static")

	return http.StripPrefix("/app/static/", http.FileServer(http.FS(sub)))
}
