// Package server runs the HTTP server and handles its graceful shutdown.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const (
	// readHeaderTimeout bounds how long a client may take to send the request
	// headers, which closes the door on slowloris-style connections.
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 1 * time.Minute
	// writeTimeout caps the time from the end of the request headers to the
	// end of the response. Raise it, or remove it, for streaming endpoints.
	writeTimeout = 1 * time.Minute
	idleTimeout  = 2 * time.Minute
	// shutdownTimeout is how long in-flight requests get to finish once the
	// listener has closed before their connections are cut.
	shutdownTimeout = 15 * time.Second
)

// Config describes what Run serves and how it shuts down.
type Config struct {
	Addr    string
	Handler http.Handler
	// DrainDelay is how long to keep serving after a shutdown has been
	// requested before closing the listener. The readiness probe already
	// fails during this window, so load balancers can stop routing new
	// requests here before connections are refused. Zero skips the wait.
	DrainDelay time.Duration
}

// Run serves cfg.Handler on cfg.Addr until ctx is cancelled. It then keeps
// serving for cfg.DrainDelay, stops accepting connections, and waits up to
// shutdownTimeout for in-flight requests to finish. It returns nil after a
// clean shutdown, and the listen or shutdown error otherwise.
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
		// ListenAndServe only returns before a shutdown when it failed.
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
	<-errc // ListenAndServe returns http.ErrServerClosed once Shutdown begins.

	slog.Info("server stopped")
	return nil
}
