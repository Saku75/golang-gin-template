package env

import (
	"fmt"
	"time"
)

const (
	shutdownDelayKey     string        = "SHUTDOWN_DELAY"
	shutdownDelayDefault time.Duration = 15 * time.Second
)

func ShutdownDelay() time.Duration {
	return lookup(shutdownDelayKey, shutdownDelayDefault, parseShutdownDelay)
}

func parseShutdownDelay(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("shutdown delay %q must not be negative", s)
	}
	return d, nil
}
