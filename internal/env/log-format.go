package env

import (
	"fmt"
	"strings"
)

type logFormat string

const (
	LogFormatText logFormat = "text"
	LogFormatJSON logFormat = "json"
)

const (
	logFormatKey     string    = "LOG_FORMAT"
	logFormatDefault logFormat = LogFormatText
)

func LogFormat() logFormat {
	return lookup(logFormatKey, logFormatDefault, func(s string) (logFormat, error) {
		var format logFormat
		err := format.UnmarshalText([]byte(s))
		return format, err
	})
}

func (f *logFormat) UnmarshalText(text []byte) error {
	switch format := logFormat(strings.ToLower(string(text))); format {
	case LogFormatText, LogFormatJSON:
		*f = format
		return nil
	default:
		return fmt.Errorf("unknown log format %q, want %q or %q", text, LogFormatText, LogFormatJSON)
	}
}
