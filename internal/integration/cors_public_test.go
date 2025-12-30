package integration

import (
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestCORS_PublicEndpoints(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://thirdparty.local")
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/public/chatbots/00000000-0000-0000-0000-000000000000", nil)
	req.Header.Set("Origin", "http://thirdparty.local")
	res, _ := http.DefaultClient.Do(req)
	if res.Header.Get("Access-Control-Allow-Origin") != "http://thirdparty.local" {
		t.Fatalf("missing allow origin header")
	}
	res.Body.Close()
}
