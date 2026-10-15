package env

import "log/slog"

const (
	logLevelKey     string     = "LOG_LEVEL"
	logLevelDefault slog.Level = slog.LevelInfo
)

func LogLevel() slog.Level {
	return lookup(logLevelKey, logLevelDefault, func(s string) (slog.Level, error) {
		var level slog.Level
		err := level.UnmarshalText([]byte(s))
		return level, err
	})
}
