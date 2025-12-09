package integration

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/onurceri/botla-co/pkg/config"
)

func TestNewTestMux_CORSPreflightAndAuth(t *testing.T) {
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "http://example.com")
	_ = os.Setenv("JWT_SECRET", "test-secret")
	cfg := config.LoadConfig()
	db := &sql.DB{}

	h := NewTestMux(cfg, db, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/api/v1/chatbots/abc/chat", nil)
	req.Header.Set("Origin", "http://example.com")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("preflight request failed: %v", err)
	}
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("got %d want %d", res.StatusCode, http.StatusNoContent)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Fatalf("cors origin header missing or wrong: %q", res.Header.Get("Access-Control-Allow-Origin"))
	}

	req2, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/chatbots/abc", nil)
	res2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("protected request failed: %v", err)
	}
	if res2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d want %d", res2.StatusCode, http.StatusUnauthorized)
	}
}
