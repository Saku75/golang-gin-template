package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type exampleOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello from v1!"`
	}
}

func registerExample(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "v1-example",
		Method:      http.MethodGet,
		Path:        "/example",
		Summary:     "Example v1 route",
		Tags:        []string{"v1"},
	}, func(_ context.Context, _ *struct{}) (*exampleOutput, error) {
		out := &exampleOutput{}
		out.Body.Message = "Hello from v1!"
		return out, nil
	})
}
