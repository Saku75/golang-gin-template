// Package v1 is version 1 of the API. Add a file per handler and register it
// here; paths are relative to /v1.
package v1

import "github.com/danielgtaylor/huma/v2"

func Register(api huma.API) {
	group := huma.NewGroup(api, "/v1")
	registerRoot(group)
	registerExample(group)
}
