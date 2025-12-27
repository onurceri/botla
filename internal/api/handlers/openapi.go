package handlers

import (
	"net/http"
	"os"
)

// ServeOpenAPI serves the OpenAPI specification
func ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	spec, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		http.Error(w, "OpenAPI spec not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(spec)
}
