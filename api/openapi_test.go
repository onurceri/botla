package api

import (
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPISpec_Valid(t *testing.T) {
	spec, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("failed to read spec: %v", err)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	// Validate the spec
	if err := doc.Validate(loader.Context); err != nil {
		t.Errorf("spec validation failed: %v", err)
	}
}

func TestOpenAPISpec_HasRequiredPaths(t *testing.T) {
	spec, _ := os.ReadFile("openapi.yaml")
	loader := openapi3.NewLoader()
	doc, _ := loader.LoadFromData(spec)

	requiredPaths := []string{
		"/auth/login",
		"/auth/register",
		"/chatbots",
		"/chatbots/{id}",
		"/chatbots/{id}/sources",
		"/sources/{id}/job",
	}

	for _, path := range requiredPaths {
		if doc.Paths.Find(path) == nil {
			t.Errorf("missing required path: %s", path)
		}
	}
}
