package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/OWNER/PROJECT_NAME/internal/httpapi"
	"github.com/OWNER/PROJECT_NAME/internal/web"
)

// Service represents a long-running application service.
type Service interface {
	Start(ctx context.Context) error
	Ready() <-chan struct{}
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Addr() string
}

// BackendService is the HTTP API server.
type BackendService struct {
	server   *http.Server
	listener net.Listener
	ready    chan struct{}
}

// NewBackendService creates the HTTP backend service.
func NewBackendService(cfg BackendConfig, deps *Deps) (*BackendService, error) {
	renderer, err := web.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("create renderer: %w", err)
	}

	handlers := web.NewHandlers(renderer, deps.UserRepo)
	router := httpapi.NewRouter(handlers, deps.Logger)

	return &BackendService{
		server: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second, //nolint:mnd // server constant
		},
		ready: make(chan struct{}),
	}, nil
}

// Start begins listening on the configured port.
func (s *BackendService) Start(ctx context.Context) error {
	var lc net.ListenConfig

	ln, err := lc.Listen(ctx, "tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.listener = ln
	close(s.ready)

	return nil
}

// Ready returns a channel that closes when the service is ready.
func (s *BackendService) Ready() <-chan struct{} {
	return s.ready
}

// Run serves HTTP requests until the context is cancelled.
func (s *BackendService) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}

		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

// Stop gracefully shuts down the HTTP server.
func (s *BackendService) Stop(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}

// Addr returns the listener address.
func (s *BackendService) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}

	return s.server.Addr
}
