// Package handlers holds the API's HTTP handlers. Each handler lives in its
// own file with a register function, versioned handlers live in a subpackage
// per version, and Register mounts them all.
package handlers

import (
	"github.com/danielgtaylor/huma/v2"

	"gin_template/internal/handlers/v1"
)

// Register mounts every handler on api. The checks are run by the readiness
// probe; pass one per dependency the service needs before it can take
// traffic.
func Register(api huma.API, checks ...Check) {
	registerHealthz(api)
	registerReadyz(api, checks)
	v1.Register(api)
}
