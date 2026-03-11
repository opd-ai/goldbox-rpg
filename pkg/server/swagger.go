package server

import (
	_ "embed"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed swagger_ui.html
var swaggerUITemplate string

// ServeSwaggerUI serves the Swagger UI interface for API documentation
func (s *RPCServer) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("swagger").Parse(swaggerUITemplate)
	if err != nil {
		http.Error(w, "Failed to parse Swagger UI template", http.StatusInternalServerError)
		return
	}

	data := struct {
		SpecURL string
	}{
		SpecURL: "/api/openapi.yaml",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render Swagger UI", http.StatusInternalServerError)
	}
}

// ServeOpenAPISpec serves the OpenAPI specification file
func (s *RPCServer) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	// Try to find the OpenAPI spec file
	specPath := filepath.Join("api", "openapi.yaml")
	
	// Check if file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	// Serve the file
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow CORS for Swagger UI
	http.ServeFile(w, r, specPath)
}
