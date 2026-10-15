package env

import (
	"log/slog"
	"os"
)

func lookup[T any](key string, def T, parse func(string) (T, error)) T {
	value := os.Getenv(key)
	if value == "" {
		slog.Info("environment variable not set, using default",
			slog.String("env", key),
			slog.Any("default", def),
		)
		return def
	}

	parsed, err := parse(value)
	if err != nil {
		slog.Warn("environment variable has invalid value, using default",
			slog.String("env", key),
			slog.String("value", value),
			slog.Any("error", err),
			slog.Any("default", def),
		)
		return def
	}

	return parsed
}
