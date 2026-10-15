// Package v1 holds the handlers for version 1 of the API. Register mounts
// them all under the /v1 prefix, so each handler declares only the path
// below that.
package v1

import "github.com/danielgtaylor/huma/v2"

// Register mounts every v1 handler on api under the /v1 prefix.
func Register(api huma.API) {
	group := huma.NewGroup(api, "/v1")
	registerRoot(group)
	registerExample(group)
}
