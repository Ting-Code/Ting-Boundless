// Package httpx provides a small HTTP server bootstrap with graceful shutdown,
// health endpoints, and baseline middleware shared by all services.
package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server wraps http.Server with graceful shutdown.
type Server struct {
	addr string
	log  *slog.Logger
	http *http.Server
}

// New creates a Server listening on addr with the given handler.
func New(addr string, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		addr: addr,
		log:  log,
		http: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

// Run starts the server and blocks until SIGINT/SIGTERM, then shuts down
// gracefully (stop accepting new requests, finish in-flight, then exit).
func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("server starting", slog.String("addr", s.addr))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.http.Shutdown(shutdownCtx)
}

// Env reads an environment variable with a fallback default.
func Env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
