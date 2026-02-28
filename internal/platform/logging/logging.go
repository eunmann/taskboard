// Package logging provides structured logging utilities built on zerolog.
package logging

import (
	"os"
	"strings"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/platform/env"
	"github.com/rs/zerolog"
)

var defaultLogger zerolog.Logger //nolint:gochecknoglobals // package-level logger

func init() { //nolint:gochecknoinits // logger initialization
	defaultLogger = NewConsoleLogger(zerolog.InfoLevel)
}

// NewLogger creates a JSON-formatted logger.
func NewLogger(level zerolog.Level) zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().
		Logger()
}

// NewConsoleLogger creates a human-readable console logger.
func NewConsoleLogger(level zerolog.Level) zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Kitchen,
	}).
		Level(level).
		With().
		Timestamp().
		Logger()
}

// ParseLevel converts a string to a zerolog level.
func ParseLevel(levelStr string) zerolog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// SetupFromEnv configures a logger from LOG_LEVEL and LOG_FORMAT env vars.
func SetupFromEnv() zerolog.Logger {
	level := ParseLevel(env.GetOrDefault(env.LogLevel, "info"))
	format := env.GetOrDefault(env.LogFormat, "json")

	if format == "console" {
		return NewConsoleLogger(level)
	}

	return NewLogger(level)
}

// DefaultLogger returns the package-level default logger.
func DefaultLogger() *zerolog.Logger {
	return &defaultLogger
}

// SetDefaultLogger overrides the package-level default logger.
func SetDefaultLogger(logger zerolog.Logger) {
	defaultLogger = logger
}
