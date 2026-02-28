// Package server provides the HTTP server for the taskboard web UI.
package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/eunmann/taskboard/internal/index"
	"github.com/eunmann/taskboard/internal/web"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 5 * time.Second
)

// Server is the taskboard HTTP server.
type Server struct {
	httpServer *http.Server
	addr       string
	onReady    func(addr net.Addr)
}

// New creates a new Server with all routes registered.
func New(idx *index.Index, addr string) (*Server, error) {
	renderer, err := NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("create renderer: %w", err)
	}

	mux := http.NewServeMux()

	staticSub, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("static fs: %w", err)
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	mux.HandleFunc("GET /{$}", handleList(idx, renderer))
	mux.HandleFunc("GET /task/{id}", handleDetail(idx, renderer))
	mux.HandleFunc("GET /partials/table", handleTablePartial(idx, renderer))

	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		addr: addr,
	}, nil
}

// OnReady sets a callback invoked when the server starts listening.
func (s *Server) OnReady(fn func(addr net.Addr)) {
	s.onReady = fn
}

// ListenAndServe starts the server and blocks until ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	lc := net.ListenConfig{}

	ln, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	errCh := make(chan error, 1)

	go func() {
		if serveErr := s.httpServer.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}

		close(errCh)
	}()

	if s.onReady != nil {
		s.onReady(ln.Addr())
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}

		return nil
	case err := <-errCh:
		return err
	}
}
