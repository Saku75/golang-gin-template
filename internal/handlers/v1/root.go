package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type rootOutput struct {
	Body struct {
		Version string `json:"version" enum:"v1" doc:"API version served under this prefix."`
	}
}

func registerRoot(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "v1-root",
		Method:      http.MethodGet,
		Path:        "", // Empty: the route is the /v1 prefix itself.
		Summary:     "API v1 root",
		Tags:        []string{"v1"},
	}, func(_ context.Context, _ *struct{}) (*rootOutput, error) {
		out := &rootOutput{}
		out.Body.Version = "v1"
		return out, nil
	})
}
