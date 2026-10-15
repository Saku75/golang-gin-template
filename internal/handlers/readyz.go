package handlers

import (
	"context"
	"net/http"
	"sync"

	"github.com/danielgtaylor/huma/v2"
)

// Check is one readiness check. Run returns nil while its dependency is
// ready and should respect ctx. Names must be unique.
type Check struct {
	Name string
	Run  func(ctx context.Context) error
}

type readyzOutput struct {
	Body struct {
		Status string            `json:"status" enum:"ok" doc:"ok when every check passed."`
		Checks map[string]string `json:"checks,omitempty" doc:"Result of each readiness check by name."`
	}
}

func registerReadyz(lifetime context.Context, api huma.API, checks []Check) {
	huma.Register(api, huma.Operation{
		OperationID: "readyz",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "Readiness probe",
		Description: "Runs every readiness check. Returns 200 when all pass, or 503 listing each check that failed. Also returns 503 once the service has started shutting down.",
		Tags:        []string{"Health"},
		Errors:      []int{http.StatusServiceUnavailable},
	}, func(ctx context.Context, _ *struct{}) (*readyzOutput, error) {
		if lifetime.Err() != nil {
			return nil, huma.Error503ServiceUnavailable("service is shutting down")
		}

		results := make([]error, len(checks))
		var wg sync.WaitGroup
		for i, check := range checks {
			wg.Go(func() { results[i] = check.Run(ctx) })
		}
		wg.Wait()

		var failures []error
		out := &readyzOutput{}
		out.Body.Status = "ok"
		out.Body.Checks = make(map[string]string, len(checks))
		for i, check := range checks {
			if err := results[i]; err != nil {
				out.Body.Checks[check.Name] = err.Error()
				failures = append(failures, &huma.ErrorDetail{
					Location: "checks." + check.Name,
					Message:  err.Error(),
				})
				continue
			}
			out.Body.Checks[check.Name] = "ok"
		}

		if len(failures) > 0 {
			return nil, huma.Error503ServiceUnavailable("service is not ready", failures...)
		}
		return out, nil
	})
}
