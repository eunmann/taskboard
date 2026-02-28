// Package httpapi provides HTTP routing and middleware.
package httpapi

import (
	"net/http"

	"github.com/OWNER/PROJECT_NAME/internal/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// NewRouter creates the application HTTP router with all middleware and routes.
func NewRouter(handlers *web.Handlers, logger zerolog.Logger) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(RecovererMiddleware(logger))
	r.Use(middleware.RequestID)
	r.Use(ContextLoggerMiddleware(logger))
	r.Use(AccessLogMiddleware(logger))

	// Health check (no auth)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Static assets
	r.Handle("/app/static/*", web.StaticHandler())

	// Vendor assets (aggressive caching)
	r.Group(func(r chi.Router) {
		r.Use(VendorCacheMiddleware)
		r.Handle("/app/static/vendor/*", web.StaticHandler())
	})

	// Page routes
	r.Group(func(r chi.Router) {
		r.Use(NoCacheMiddleware)

		r.Get("/", handlers.Home)
		r.Get("/about", handlers.About)

		// HTMX fragment endpoints
		r.Get("/fragment/greeting", handlers.GreetingFragment)
		r.Get("/fragment/users", handlers.UsersFragment)
	})

	return r
}
