package env

import (
	"fmt"
	"strings"
)

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

func (m *mode) UnmarshalText(text []byte) error {
	switch parsed := mode(strings.ToLower(string(text))); parsed {
	case ModeDevelopment, ModeStaging, ModeRelease:
		*m = parsed
		return nil
	default:
		return fmt.Errorf("unknown mode %q, want %q, %q or %q", text, ModeDevelopment, ModeStaging, ModeRelease)
	}
}
