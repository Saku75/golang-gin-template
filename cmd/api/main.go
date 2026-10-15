package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"

	"gin_template/internal/env"
	"gin_template/internal/logger"
)

type rootOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!"`
	}
}

func main() {
	log := logger.Setup()

	// Must run before gin.New so the mode applies to the engine's setup.
	switch env.Mode() {
	case env.ModeRelease, env.ModeStaging:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(env.TrustedProxies()); err != nil {
		slog.Error("invalid trusted proxies", slog.Any("error", err))
		os.Exit(1)
	}
	router.Use(sloggin.New(log), gin.Recovery())

	api := humagin.New(router, huma.DefaultConfig("Gin Template API", "1.0.0"))

	huma.Get(api, "/", func(_ context.Context, _ *struct{}) (*rootOutput, error) {
		out := &rootOutput{}
		out.Body.Message = "Hello, world!"
		return out, nil
	})

	addr := fmt.Sprintf(":%d", env.Port())
	slog.Info("starting server", slog.String("addr", addr))
	if err := router.Run(addr); err != nil {
		slog.Error("server stopped", slog.Any("error", err))
		os.Exit(1)
	}
}
