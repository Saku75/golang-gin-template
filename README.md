# Go API template with Gin and Huma

A starting point for HTTP APIs in Go. Gin handles routing and middleware, Huma
adds typed request and response handling with an OpenAPI spec and docs
generated from the code.

## What's included

- Typed, validated configuration from environment variables, with logged
  fallbacks when a value is missing or invalid
- Structured logging with `log/slog`, in text or JSON, including Gin's own
  output and every request
- Liveness and readiness probes, with readiness checks you plug in, and a
  readiness probe that fails during shutdown so traffic drains cleanly
- Versioned handler layout with `/v1` as the first version
- Graceful shutdown that drains in-flight requests on SIGINT and SIGTERM
- CORS and trusted-proxy handling configured from the environment
- Live reload with Air, and a multi-stage Dockerfile producing a static,
  non-root image

## Requirements

- Go 1.27 or newer
- Docker, only for building the image

## Getting started

```sh
cp .env.example .env   # adjust as needed
go tool air            # build, run, and reload on change
```

Or run it once without reloading:

```sh
go run ./cmd/api
```

Note that `.env` is loaded by Air, not by the application. When running the
binary directly, set the variables in the shell.

The API then serves these routes:

| Route               | Purpose                                                   |
| ------------------- | --------------------------------------------------------- |
| `GET /healthz`      | Liveness probe, 200 while the process is up               |
| `GET /readyz`       | Readiness probe, 503 when a check fails or during shutdown |
| `GET /v1`           | Root of API version 1                                     |
| `GET /v1/example`   | Example route                                             |
| `GET /docs`         | Interactive API documentation                             |
| `GET /openapi.json` | Generated OpenAPI spec (also `.yaml`)                     |

## Configuration

Every variable has a default. A missing variable is logged at info level and
an invalid one at warn level, and the default is used in both cases.

| Variable               | Default       | Description                                                                 |
| ---------------------- | ------------- | --------------------------------------------------------------------------- |
| `MODE`                 | `development` | `development`, `staging`, or `release`. Development runs Gin in debug mode. |
| `PORT`                 | `8080`        | Port to listen on.                                                          |
| `LOG_LEVEL`            | `info`        | `debug`, `info`, `warn`, or `error`.                                        |
| `LOG_FORMAT`           | `text`        | `text` or `json`.                                                           |
| `TRUSTED_PROXIES`      | none          | Comma-separated IPs or CIDRs whose `X-Forwarded-For` header is trusted.     |
| `CORS_ALLOWED_ORIGINS` | none          | Comma-separated origins allowed from browsers, or `*`. Empty disables CORS. |
| `SHUTDOWN_DELAY`       | `15s`         | How long to keep serving after SIGTERM while `/readyz` fails. `0s` disables. |

## Project layout

```
cmd/api/           Entry point: wires config, logging, middleware, routes, and the server
internal/env/      One file per environment variable, built on a shared typed lookup
internal/logger/   Logger setup, including replay of anything logged before it was configured
internal/handlers/ HTTP handlers; probes at the top level, versioned routes in v1/
internal/server/   HTTP server with timeouts and graceful shutdown
```

### Adding a handler

Create a file in `internal/handlers/v1` with a register function and add a
call to it in `v1.Register`. Paths are relative to the `/v1` prefix:

```go
func registerThing(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "v1-get-thing",
		Method:      http.MethodGet,
		Path:        "/things/{id}",
		Summary:     "Get a thing",
		Tags:        []string{"v1"},
	}, func(ctx context.Context, in *thingInput) (*thingOutput, error) {
		// ...
	})
}
```

A new API version is a new subpackage next to `v1`, mounted from
`handlers.Register`.

### Adding a readiness check

Pass checks to `handlers.Register` in `cmd/api/main.go`. Each has a name and a
function that returns an error while its dependency is unavailable:

```go
handlers.Register(ctx, api, handlers.Check{Name: "db", Run: db.PingContext})
```

Checks run concurrently on every `/readyz` request, and a failing one makes the
probe answer 503 with the check named in the error details.

### Adding an environment variable

Copy one of the files in `internal/env`. Each declares the variable name, a
default, and a parser, and reads the value through `lookup`, which handles
the logging and fallback.

## Docker

```sh
docker build -t api .
docker run --rm -p 8080:8080 api
```

The image runs in release mode with JSON logs by default. Override any variable with `-e`, for example `-e LOG_LEVEL=debug`.

## Shutdown

On SIGINT or SIGTERM the server:

1. Fails `/readyz` immediately, while `/healthz` keeps answering 200.
2. Keeps serving for `SHUTDOWN_DELAY`, so load balancers that watch the
   readiness probe stop sending new requests before the listener closes.
3. Stops accepting connections and gives in-flight requests up to 15 seconds
   to finish, then exits.

On Kubernetes, set the pod's `terminationGracePeriodSeconds` above
`SHUTDOWN_DELAY` plus those 15 seconds, since the default of 30 is exactly the
sum with the default delay, and give the readiness probe a short
`periodSeconds` so the failure is noticed within the delay. The timeouts
live in `internal/server/server.go`.

## Using this as a template

Rename the module in `go.mod`, then update the `gin_template/...` import paths
to match:

```sh
go mod edit -module github.com/you/yourapi
grep -rl 'gin_template/' --include='*.go' . | xargs sed -i 's|gin_template/|github.com/you/yourapi/|g'
```
