package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/stretchr/testify/mock"
)

func TestNewTestMux_CORSPreflightAndAuth(t *testing.T) {
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "http://example.com")
	_ = os.Setenv("JWT_SECRET", "test-secret")
	cfg := config.LoadConfig()
	db := testdb.OpenParallelTestDB(t)

	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)
	mockLLM := &rag.MockFullClient{}

	h, _ := NewTestMux(cfg, db, nil, mockLLM, mockVC)
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
