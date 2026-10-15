// Package handlers holds the API's HTTP handlers. Each handler lives in its
// own file with a register function, versioned handlers live in a subpackage
// per version, and Register mounts them all.
package handlers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"gin_template/internal/handlers/v1"
)

// Register mounts every handler on api.
//
// lifetime is the application's context. Once it is cancelled the readiness
// probe answers 503, so load balancers stop routing new requests here while
// the server drains. The checks are also run by the readiness probe; pass one
// per dependency the service needs before it can take traffic.
func Register(lifetime context.Context, api huma.API, checks ...Check) {
	registerHealthz(api)
	registerReadyz(lifetime, api, checks)
	v1.Register(api)
}
