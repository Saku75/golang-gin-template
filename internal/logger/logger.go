package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"gin_template/internal/env"
)

// Setup configures slog from LOG_FORMAT and LOG_LEVEL, installs it as the
// default logger, and returns it. Call it first: anything logged before it
// finishes is replayed through the configured logger.
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

	// Gin's own debug output goes through slog as well.
	gin.DebugPrintFunc = func(format string, values ...any) {
		l.Debug(strings.TrimSpace(fmt.Sprintf(format, values...)))
	}

	return l
}
