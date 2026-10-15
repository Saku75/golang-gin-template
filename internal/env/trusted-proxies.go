package env

import (
	"fmt"
	"net"
	"strings"
)

const trustedProxiesKey string = "TRUSTED_PROXIES"

// trustedProxiesDefault trusts no proxies, so X-Forwarded-For is ignored and
// the client IP is always the direct peer.
var trustedProxiesDefault []string

func TrustedProxies() []string {
	return lookup(trustedProxiesKey, trustedProxiesDefault, parseTrustedProxies)
}

// parseTrustedProxies splits a comma-separated list of IPs and CIDRs, the
// forms gin.Engine.SetTrustedProxies accepts, and rejects anything else so a
// typo falls back to trusting nothing rather than everything.
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
