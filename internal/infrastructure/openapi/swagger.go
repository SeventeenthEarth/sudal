package openapi

import (
	"html/template"
	"net/http"

	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"go.uber.org/zap"
)

// SwaggerHandler provides Swagger UI for the OpenAPI documentation
type SwaggerHandler struct {
	specPath string
}

// NewSwaggerHandler creates a new Swagger UI handler
func NewSwaggerHandler(specPath string) *SwaggerHandler {
	return &SwaggerHandler{
		specPath: specPath,
	}
}

// ServeSwaggerUI serves the Swagger UI protocol
func (h *SwaggerHandler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	log.Info("Serving Swagger UI", zap.String("path", r.URL.Path))

	// Serve the main Swagger UI HTML page
	h.serveSwaggerHTML(w, r)
}

// serveSwaggerHTML serves the main Swagger UI HTML page
func (h *SwaggerHandler) serveSwaggerHTML(w http.ResponseWriter, r *http.Request) {
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Sudal API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '{{.SpecURL}}',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	tmpl, err := template.New("swagger").Parse(htmlTemplate)
	if err != nil {
		log.Error("Failed to parse swagger template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		SpecURL string
	}{
		SpecURL: "/api/openapi.yaml",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Error("Failed to execute swagger template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// ServeOpenAPISpec serves the OpenAPI specification file
func (h *SwaggerHandler) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	log.Info("Serving OpenAPI spec", zap.String("path", r.URL.Path))

	// Read the OpenAPI spec file
	http.ServeFile(w, r, h.specPath)
}
