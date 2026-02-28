package httpapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/platform/logging"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// responseRecorder captures status code and bytes written.
type responseRecorder struct {
	http.ResponseWriter

	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.statusCode = code
		r.wroteHeader = true
	}

	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += n

	if err != nil {
		return n, fmt.Errorf("write response: %w", err)
	}

	return n, nil
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// RecovererMiddleware recovers from panics and logs the error.
func RecovererMiddleware(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID := middleware.GetReqID(r.Context())
					logger.Error().
						Str("request_id", reqID).
						Interface("panic", err).
						Msg("recovered from panic")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ContextLoggerMiddleware adds a request-scoped logger to the context.
func ContextLoggerMiddleware(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.GetReqID(r.Context())
			reqLogger := logger.With().Str("request_id", reqID).Logger()
			ctx := logging.WithLogger(r.Context(), reqLogger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AccessLogMiddleware logs completed requests.
func AccessLogMiddleware(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)

			statusThreshold := 400
			serverErrThreshold := 500

			level := zerolog.InfoLevel
			if rec.statusCode >= serverErrThreshold {
				level = zerolog.ErrorLevel
			} else if rec.statusCode >= statusThreshold {
				level = zerolog.WarnLevel
			}

			logger.WithLevel(level).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rec.statusCode).
				Int("bytes", rec.bytesWritten).
				Dur("duration", duration).
				Str("remote", r.RemoteAddr).
				Msg("request completed")
		})
	}
}

// NoCacheMiddleware sets no-cache headers.
func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}

// VendorCacheMiddleware sets aggressive caching for vendor assets.
func VendorCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
