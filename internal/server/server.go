// Package server runs the HTTP server with graceful shutdown.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 1 * time.Minute
	writeTimeout      = 1 * time.Minute // Caps response time; raise for streaming endpoints.
	idleTimeout       = 2 * time.Minute
	shutdownTimeout   = 15 * time.Second // Time in-flight requests get after the listener closes.
)

type Config struct {
	Addr    string
	Handler http.Handler
	// DrainDelay keeps the server serving after shutdown is requested, while
	// /readyz already fails, so load balancers can stop routing here first.
	DrainDelay time.Duration
}

// Run serves until ctx is cancelled, then drains and shuts down. The README's
// Shutdown section describes the sequence.
func Run(ctx context.Context, cfg Config) error {
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           cfg.Handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("server listening", slog.String("addr", cfg.Addr))
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return fmt.Errorf("listen on %s: %w", cfg.Addr, err)
	case <-ctx.Done():
	}

	if cfg.DrainDelay > 0 {
		slog.Info("server draining", slog.Duration("delay", cfg.DrainDelay))
		time.Sleep(cfg.DrainDelay)
	}

	slog.Info("server shutting down", slog.Duration("timeout", shutdownTimeout))
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	<-errc // ListenAndServe returns once Shutdown begins.

	slog.Info("server stopped")
	return nil
}
