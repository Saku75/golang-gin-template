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

// run wires the application together. Start here to add middleware, readiness
// checks, or anything that needs the application's lifetime context.
func run() error {
	log := logger.Setup()

	// ctx ends on SIGINT or SIGTERM, which starts the graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// gin.SetMode must run before gin.New.
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
	router.Use(sloggin.NewWithFilters(log, sloggin.IgnorePath("/healthz", "/readyz")), gin.Recovery())

	if origins := env.CORSAllowedOrigins(); len(origins) > 0 {
		cfg := cors.DefaultConfig()
		cfg.AllowOrigins = origins
		cfg.AllowHeaders = append(cfg.AllowHeaders, "Authorization")
		router.Use(cors.New(cfg))
	}

	cfg := huma.DefaultConfig("Gin Template API", "1.0.0")
	cfg.DocsPath = "" // Served from embedded files instead; see handlers/docs.
	api := humagin.New(router, cfg)
	// Readiness checks go here; see handlers.Check.
	handlers.Register(ctx, api)

	return server.Run(ctx, server.Config{
		Addr:       fmt.Sprintf(":%d", env.Port()),
		Handler:    router,
		DrainDelay: env.ShutdownDelay(),
	})
}
