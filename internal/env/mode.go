package env

import (
	"fmt"
	"strings"
)

// mode selects how the application behaves at runtime. The type is unexported
// so that Mode can be the accessor, matching Port, LogLevel and LogFormat;
// callers compare the result against the Mode* constants.
type mode string

const (
	ModeDevelopment mode = "development"
	ModeStaging     mode = "staging"
	ModeRelease     mode = "release"
)

const (
	modeKey     string = "MODE"
	modeDefault mode   = ModeDevelopment
)

func Mode() mode {
	return lookup(modeKey, modeDefault, func(s string) (mode, error) {
		var m mode
		err := m.UnmarshalText([]byte(s))
		return m, err
	})
}

// UnmarshalText parses a mode name case-insensitively, mirroring slog.Level.
func (m *mode) UnmarshalText(text []byte) error {
	switch parsed := mode(strings.ToLower(string(text))); parsed {
	case ModeDevelopment, ModeStaging, ModeRelease:
		*m = parsed
		return nil
	default:
		return fmt.Errorf("unknown mode %q, want %q, %q or %q", text, ModeDevelopment, ModeStaging, ModeRelease)
	}
}
