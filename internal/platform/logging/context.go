package logging

import (
	"context"

	"github.com/rs/zerolog"
)

type contextKey struct{}

// WithLogger stores a logger in the context.
func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// LoggerFrom retrieves the logger from context, or returns the default.
func LoggerFrom(ctx context.Context) zerolog.Logger {
	if l, ok := ctx.Value(contextKey{}).(zerolog.Logger); ok {
		return l
	}

	return defaultLogger
}

// With adds a string field to the context logger.
func With(ctx context.Context, key, value string) context.Context {
	l := LoggerFrom(ctx).With().Str(key, value).Logger()

	return WithLogger(ctx, l)
}

// WithFields adds multiple string fields to the context logger.
func WithFields(ctx context.Context, fields map[string]string) context.Context {
	c := LoggerFrom(ctx).With()
	for k, v := range fields {
		c = c.Str(k, v)
	}

	return WithLogger(ctx, c.Logger())
}
