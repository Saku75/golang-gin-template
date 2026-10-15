package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"

	"gin_template/internal/env"
	"gin_template/internal/handlers"
	"gin_template/internal/logger"
	"gin_template/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("exiting", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	log := logger.Setup()

	// SIGINT and SIGTERM cancel ctx, which fails the readiness probe and
	// starts a graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Must run before gin.New so the mode applies to the engine's setup.
	switch env.Mode() {
	case env.ModeRelease, env.ModeStaging:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(env.TrustedProxies()); err != nil {
		return fmt.Errorf("set trusted proxies: %w", err)
	}
	// Probes are polled constantly, so keep them out of the request log.
	router.Use(sloggin.NewWithFilters(log, sloggin.IgnorePath("/healthz", "/readyz")), gin.Recovery())

	// With no origins configured no CORS headers are sent, so browsers on
	// other origins are refused. "*" allows every origin.
	if origins := env.CORSAllowedOrigins(); len(origins) > 0 {
		cfg := cors.DefaultConfig()
		cfg.AllowOrigins = origins
		cfg.AllowHeaders = append(cfg.AllowHeaders, "Authorization")
		router.Use(cors.New(cfg))
	}

	api := humagin.New(router, huma.DefaultConfig("Gin Template API", "1.0.0"))
	handlers.Register(ctx, api)

	return server.Run(ctx, server.Config{
		Addr:       fmt.Sprintf(":%d", env.Port()),
		Handler:    router,
		DrainDelay: env.ShutdownDelay(),
	})
}
