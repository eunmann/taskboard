// Package web provides HTTP handlers for the web UI.
package web

import (
	"net/http"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/platform/logging"
	"github.com/OWNER/PROJECT_NAME/internal/user"
)

// PageData holds data passed to page templates.
type PageData struct {
	Title   string
	Content any
}

// Handlers holds all web handler dependencies.
type Handlers struct {
	renderer *Renderer
	userRepo *user.Repo
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(renderer *Renderer, userRepo *user.Repo) *Handlers {
	return &Handlers{
		renderer: renderer,
		userRepo: userRepo,
	}
}

// Home renders the home page.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFrom(ctx)

	users, err := h.userRepo.List(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to list users")

		users = nil
	}

	data := PageData{
		Title:   "Home",
		Content: users,
	}

	if err := h.renderer.Render(w, "home", data); err != nil {
		logger.Error().Err(err).Msg("failed to render home page")
	}
}

// About renders the about page.
func (h *Handlers) About(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFrom(r.Context())

	data := PageData{
		Title: "About",
	}

	if err := h.renderer.Render(w, "about", data); err != nil {
		logger.Error().Err(err).Msg("failed to render about page")
	}
}

// GreetingFragment renders an HTMX greeting fragment.
func (h *Handlers) GreetingFragment(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFrom(r.Context())

	greeting := "Hello from the server! The time is " + time.Now().Format(time.Kitchen)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := h.renderer.RenderPartial(w, "home", "greeting", greeting)
	if err != nil {
		logger.Error().Err(err).Msg("failed to render greeting fragment")
	}
}

// UsersFragment renders the users table fragment (for HTMX refresh).
func (h *Handlers) UsersFragment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFrom(ctx)

	users, err := h.userRepo.List(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to list users")
		http.Error(w, "Failed to load users", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.renderer.RenderPartial(w, "home", "users_table", users); err != nil {
		logger.Error().Err(err).Msg("failed to render users fragment")
	}
}
