// Package main provides the all-in-one development runner that runs all services
// in a single process. This is useful for local development and integration testing.
package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/app"
	"github.com/OWNER/PROJECT_NAME/internal/platform/logging"
)

func main() {
	// Initialize logger
	logger := logging.SetupFromEnv()
	logging.SetDefaultLogger(logger)

	logger.Info().Msg("starting all-in-one development server")

	cfg := app.ConfigFromEnv()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctx = logging.WithLogger(ctx, logger)

	deps, err := app.NewDeps(ctx, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create dependencies")
	}

	defer func() { _ = deps.Close() }()

	// Create backend service
	backend, err := app.NewBackendService(cfg.Backend, deps)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create backend service")
	}

	// Start
	if err := backend.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start backend service")
	}

	// Wait for ready
	select {
	case <-backend.Ready():
		logger.Info().Str("addr", backend.Addr()).Msg("backend server ready")
	case <-ctx.Done():
		logger.Fatal().Msg("context cancelled while waiting for service")
	case <-time.After(30 * time.Second): //nolint:mnd // service constant
		logger.Fatal().Msg("timeout waiting for service to be ready")
	}

	logger.Info().Msg("all services ready")

	// Run until shutdown
	if err := backend.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("backend server error")
	}

	// Graceful stop
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd // service constant
	defer stopCancel()

	if err := backend.Stop(stopCtx); err != nil {
		logger.Error().Err(err).Msg("error stopping backend")
	}

	logger.Info().Msg("all services stopped")
}
