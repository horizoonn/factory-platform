package swaggerui

import (
	"html/template"
	"io"
	"net/http"
)

const swaggerUIVersion = "5.32.8"

var pageTemplate = template.Must(template.New("swagger-ui").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function () {
      SwaggerUIBundle({
        url: {{.SpecURL}},
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`))

type Config struct {
	Title           string
	UIPath          string
	SpecPath        string
	Spec            string
	SpecContentType string
}

type Handler struct {
	config Config
}

func NewHandler(config Config) *Handler {
	return &Handler{config: config}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case h.config.UIPath:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.Method == http.MethodHead {
			return
		}

		data := struct {
			Title   string
			SpecURL string
		}{
			Title:   h.config.Title,
			SpecURL: h.config.SpecPath,
		}
		if err := pageTemplate.Execute(w, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	case h.config.SpecPath:
		w.Header().Set("Content-Type", h.config.SpecContentType)
		if r.Method == http.MethodGet {
			if _, err := io.WriteString(w, h.config.Spec); err != nil {
				return
			}
		}
	default:
		http.NotFound(w, r)
	}
}
