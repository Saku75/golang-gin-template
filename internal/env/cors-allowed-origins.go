package env

import (
	"fmt"
	"net/url"
	"strings"
)

const corsAllowedOriginsKey string = "CORS_ALLOWED_ORIGINS"

// corsAllowedOriginsDefault allows no cross-origin requests: with no origins
// configured the CORS middleware is not installed at all.
var corsAllowedOriginsDefault []string

func CORSAllowedOrigins() []string {
	return lookup(corsAllowedOriginsKey, corsAllowedOriginsDefault, parseCORSAllowedOrigins)
}

// parseCORSAllowedOrigins splits a comma-separated list of origins. Each must
// be "*" or scheme://host[:port] with an http or https scheme and nothing
// after the host, which is the form browsers send in the Origin header.
func parseCORSAllowedOrigins(s string) ([]string, error) {
	var origins []string
	for _, entry := range strings.Split(s, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if entry == "*" {
			origins = append(origins, entry)
			continue
		}

		u, err := url.Parse(entry)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
			u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
			return nil, fmt.Errorf("invalid CORS origin %q, want * or scheme://host[:port] with http or https", entry)
		}
		origins = append(origins, entry)
	}
	return origins, nil
}
