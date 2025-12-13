package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestChatbotsDispatchHandler_Routes(t *testing.T) {
	ch := &handlers.ChatbotHandlers{}
	sh := &handlers.SourcesHandlers{}
	chh := &handlers.ChatHandlers{}
	puh := &handlers.PendingURLsHandlers{}
	h := chatbotsDispatchHandler("secret", ch, sh, chh, puh)

	cases := []struct {
		path string
		code int
	}{
		{"/api/v1/chatbots/x/sources", http.StatusUnauthorized},
		{"/api/v1/chatbots/x/analytics/sources", http.StatusUnauthorized},
		{"/api/v1/chatbots/x/chat", http.StatusUnauthorized},
		{"/api/v1/chatbots/x", http.StatusUnauthorized},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != c.code {
			t.Fatalf("path %s: got %d want %d", c.path, w.Code, c.code)
		}
	}
}

func TestNewHTTPServer_Config(t *testing.T) {
	mux := http.NewServeMux()
	srv := newHTTPServer("8081", mux)
	if srv.Addr != ":8081" {
		t.Fatalf("addr got %s want %s", srv.Addr, ":8081")
	}
	if srv.Handler == nil {
		t.Fatalf("handler should be set")
	}
}

func TestStartAndShutdownServer(t *testing.T) {
	mux := http.NewServeMux()
	srv := newHTTPServer("0", mux)
	log := logger.New("ERROR")
	startServerAsync(srv, log, "0")
	db := testdb.OpenTestDB(t)
	defer db.Close()
	shutdownServer(srv, log, db)
}

// Backward-compatible dispatcher used by tests
func chatbotsDispatchHandler(secret string, ch *handlers.ChatbotHandlers, sh *handlers.SourcesHandlers, chh *handlers.ChatHandlers, puh *handlers.PendingURLsHandlers) http.Handler {
	rlSources := middleware.NewRateLimiterFromEnvWithPrefix("SOURCES")
	acth := &handlers.ActionHandlers{DB: ch.DB}
	hoh := &handlers.HandoffHandlers{DB: ch.DB}
	// Create handler
	h := chatbotsDispatchHandlerWithSourcesRL(secret, ch, sh, chh, puh, acth, hoh, nil, nil, rlSources)
	return h
}
