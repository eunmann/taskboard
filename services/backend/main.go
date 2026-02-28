// Package main provides the entry point for the backend API server.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/app"
	"github.com/OWNER/PROJECT_NAME/internal/platform/logging"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	healthCheck := flag.Bool("health", false, "Run health check and exit")

	flag.Parse()

	if *healthCheck {
		runHealthCheck()

		return
	}

	// Initialize logger from environment
	logger := logging.SetupFromEnv()
	logging.SetDefaultLogger(logger)

	logger.Info().Str("version", version).Str("commit", commit).Msg("starting backend")

	// Load configuration
	cfg := app.ConfigFromEnv()

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Add logger to context
	ctx = logging.WithLogger(ctx, logger)

	// Create shared dependencies
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

	// Start service
	if err := backend.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start backend service")
	}

	// Wait for ready
	select {
	case <-backend.Ready():
		logger.Info().Str("addr", backend.Addr()).Msg("backend server ready")
	case <-ctx.Done():
		logger.Fatal().Msg("context cancelled before server ready")
	}

	// Run until shutdown
	if err := backend.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("backend server error")
	}

	logger.Info().Msg("backend server stopped")
}

func runHealthCheck() {
	exitCode := doHealthCheck()
	os.Exit(exitCode)
}

func doHealthCheck() int {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd // service constant
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:"+port+"/health", http.NoBody)
	if err != nil {
		return 1
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 1
	}

	statusCode := resp.StatusCode
	_ = resp.Body.Close()

	if statusCode != http.StatusOK {
		return 1
	}

	return 0
}
