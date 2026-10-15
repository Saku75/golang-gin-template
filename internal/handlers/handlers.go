// Package handlers holds the HTTP handlers. The probes live here; routes live
// in one subpackage per API version, starting with v1.
package handlers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"gin_template/internal/handlers/v1"
)

// Register mounts every handler. Cancelling lifetime makes /readyz answer 503
// while the server drains. Pass one Check per dependency the service needs
// before it can take traffic.
func Register(lifetime context.Context, api huma.API, checks ...Check) {
	registerHealthz(api)
	registerReadyz(lifetime, api, checks)
	v1.Register(api)
}
