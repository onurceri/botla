package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestCORS_PublicEndpoints(t *testing.T) {
	t.Parallel() // Now safe - no t.Setenv()

	// Test CORS middleware directly with the specific origin
	origins := []string{"http://thirdparty.local"}
	cors := middleware.CORSMiddlewareAllowOrigins(origins)

	// Simple public endpoint mock
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/00000000-0000-0000-0000-000000000000", nil)
	req.Header.Set("Origin", "http://thirdparty.local")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "http://thirdparty.local" {
		t.Fatalf("missing allow origin header, got: %s", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}
