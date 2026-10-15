package env

import (
	"fmt"
	"net"
	"strings"
)

const trustedProxiesKey string = "TRUSTED_PROXIES"

// nil trusts no proxies.
var trustedProxiesDefault []string

func TrustedProxies() []string {
	return lookup(trustedProxiesKey, trustedProxiesDefault, parseTrustedProxies)
}

func parseTrustedProxies(s string) ([]string, error) {
	var proxies []string
	for _, entry := range strings.Split(s, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if strings.Contains(entry, "/") {
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return nil, fmt.Errorf("invalid trusted proxy %q: %w", entry, err)
			}
		} else if net.ParseIP(entry) == nil {
			return nil, fmt.Errorf("invalid trusted proxy %q, want an IP or CIDR", entry)
		}

		proxies = append(proxies, entry)
	}
	return proxies, nil
}
