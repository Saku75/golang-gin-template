// Package docs serves the API reference at /docs from files embedded in the
// binary, so it works without internet access. go generate fetches them into
// assets/ before the first build; see fetch.go for the source and version.
package docs

import (
	"embed"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

//go:generate go run fetch.go

//go:embed assets/web-components.min.js assets/styles.min.css
var assets embed.FS

const (
	path        = "/docs"
	openAPIPath = "/openapi" // Must match huma.Config.OpenAPIPath.
)

const page = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="referrer" content="no-referrer">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{title}}</title>
    <link rel="stylesheet" href="{{path}}/styles.min.css">
    <script src="{{path}}/web-components.min.js"></script>
  </head>
  <body style="height: 100vh;">
    <elements-api
      apiDescriptionUrl="{{openapi}}.yaml"
      router="hash"
      layout="sidebar"
      tryItCredentialsPolicy="same-origin"
    ></elements-api>
  </body>
</html>`

func Register(api huma.API) {
	title := "API Reference"
	if info := api.OpenAPI().Info; info != nil && info.Title != "" {
		title = info.Title + " Reference"
	}
	html := []byte(strings.NewReplacer(
		"{{title}}", title,
		"{{path}}", path,
		"{{openapi}}", openAPIPath,
	).Replace(page))

	csp := strings.Join([]string{
		"default-src 'none'",
		"base-uri 'none'",
		"connect-src 'self'",
		"form-action 'none'",
		"frame-ancestors 'none'",
		"sandbox allow-same-origin allow-scripts allow-popups allow-popups-to-escape-sandbox allow-downloads",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
	}, "; ")

	adapter := api.Adapter()
	adapter.Handle(&huma.Operation{Method: http.MethodGet, Path: path}, func(ctx huma.Context) {
		ctx.SetHeader("Content-Security-Policy", csp)
		ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
		ctx.BodyWriter().Write(html)
	})
	serve(adapter, "styles.min.css", "text/css; charset=utf-8")
	serve(adapter, "web-components.min.js", "text/javascript; charset=utf-8")
}

func serve(adapter huma.Adapter, name, contentType string) {
	data, err := assets.ReadFile("assets/" + name)
	if err != nil {
		panic(err)
	}
	adapter.Handle(&huma.Operation{Method: http.MethodGet, Path: path + "/" + name}, func(ctx huma.Context) {
		ctx.SetHeader("Content-Type", contentType)
		ctx.SetHeader("Cache-Control", "public, max-age=86400")
		ctx.BodyWriter().Write(data)
	})
}
