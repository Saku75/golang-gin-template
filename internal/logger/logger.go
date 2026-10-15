package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"gin_template/internal/env"
)

// Setup reads LOG_FORMAT and LOG_LEVEL from the environment, builds a logger
// that writes to stderr, and installs it as the slog default so that
// package-level calls such as slog.Info use it from then on. The logger is
// also returned so it can be handed to anything that takes one explicitly,
// such as request-logging middleware.
//
// Anything logged while the environment is read, such as the env package
// reporting a missing variable, is buffered and replayed through the new
// logger, so every line of output uses the configured format and level.
//
// Gin's own debug-mode output is routed through the logger at debug level,
// so it never bypasses the configured format.
func Setup() *slog.Logger {
	replay := newReplayHandler()
	slog.SetDefault(slog.New(replay))

	format, level := env.LogFormat(), env.LogLevel()
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch format {
	case env.LogFormatJSON:
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	l := slog.New(handler)
	slog.SetDefault(l)
	replay.replay(handler)

	gin.DebugPrintFunc = func(format string, values ...any) {
		l.Debug(strings.TrimSpace(fmt.Sprintf(format, values...)))
	}

	return l
}
