# Go API template with Gin and Huma

A starting point for HTTP APIs in Go. Gin handles routing and middleware, Huma
gives you typed handlers with an OpenAPI spec and docs generated from the code,
`log/slog` handles logging, and shutdown is graceful in a way Kubernetes
understands.

## Getting started

Requires Go 1.27 or newer.

1. Rename the module and its import paths:

   ```sh
   go mod edit -module github.com/you/yourapi
   grep -rl 'gin_template/' --include='*.go' . | xargs sed -i 's|gin_template/|github.com/you/yourapi/|g'
   ```

2. Run it with live reload:

   ```sh
   cp .env.example .env
   go tool air
   ```

   `.env` is loaded by Air only. For `go run ./cmd/api`, set the variables in
   the shell.

3. Open <http://localhost:8080/docs>.

Routes out of the box:

| Route               | Purpose                                                    |
| ------------------- | ---------------------------------------------------------- |
| `GET /healthz`      | Liveness probe, 200 while the process is up                |
| `GET /readyz`       | Readiness probe, 503 when a check fails or during shutdown |
| `GET /v1`           | Root of API version 1                                      |
| `GET /v1/example`   | Example route                                              |
| `GET /docs`         | Interactive API documentation                              |
| `GET /openapi.json` | Generated OpenAPI spec (also `.yaml`)                      |

## Where to look

```
cmd/api/main.go     Wires everything together. Add middleware and readiness checks here.
internal/env/       One file per environment variable. Copy one to add another.
internal/handlers/  Probes, plus one subpackage per API version. Add routes in v1/.
internal/logger/    slog setup from LOG_LEVEL and LOG_FORMAT.
internal/server/    HTTP server timeouts and the shutdown sequence.
```

## Adding things

### A route

Add a file in `internal/handlers/v1` and call its register function from
`v1.Register`. Paths are relative to `/v1`:

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

### A readiness check

Pass one to `handlers.Register` in `cmd/api/main.go` per dependency the
service needs before it can take traffic:

```go
handlers.Register(ctx, api, handlers.Check{Name: "db", Run: db.PingContext})
```

Checks run concurrently on every `/readyz` request. A failing one makes the
probe answer 503 with the check named in the error details.

### An environment variable

Copy a file in `internal/env`. Each declares the variable's name, default, and
parser, and reads it through `lookup`, which logs and falls back to the default
when the value is missing or invalid.

### An API version

Add a subpackage next to `internal/handlers/v1` and mount it from
`handlers.Register`.

## Configuration

| Variable               | Default       | Description                                                                  |
| ---------------------- | ------------- | ---------------------------------------------------------------------------- |
| `MODE`                 | `development` | `development`, `staging`, or `release`. Development runs Gin in debug mode.  |
| `PORT`                 | `8080`        | Port to listen on.                                                           |
| `LOG_LEVEL`            | `info`        | `debug`, `info`, `warn`, or `error`.                                         |
| `LOG_FORMAT`           | `text`        | `text` or `json`.                                                            |
| `TRUSTED_PROXIES`      | none          | Comma-separated IPs or CIDRs whose `X-Forwarded-For` header is trusted.      |
| `CORS_ALLOWED_ORIGINS` | none          | Comma-separated origins allowed from browsers, or `*`. Empty disables CORS.  |
| `SHUTDOWN_DELAY`       | `15s`         | How long to keep serving after SIGTERM while `/readyz` fails. `0s` disables. |

## Deploying

### Docker

```sh
docker build -t api .
docker run --rm -p 8080:8080 api
```

The image runs in release mode with JSON logs. Override any variable with
`-e`, for example `-e LOG_LEVEL=debug`.

### Shutdown

On SIGINT or SIGTERM the server:

1. Fails `/readyz` immediately, while `/healthz` keeps answering 200.
2. Keeps serving for `SHUTDOWN_DELAY`, so load balancers that watch the
   readiness probe stop sending new requests before the listener closes.
3. Stops accepting connections and gives in-flight requests up to 15 seconds
   to finish, then exits.

On Kubernetes, set `terminationGracePeriodSeconds` above `SHUTDOWN_DELAY` plus
those 15 seconds, since the default of 30 is exactly the sum with the default
delay, and give the readiness probe a short `periodSeconds` so the failure is
noticed within the delay. The timeouts live in `internal/server/server.go`.
