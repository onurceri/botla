package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/pkg/logger"
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
	db, _ := sql.Open("pgx", "postgres://localhost")
	shutdownServer(srv, log, db)
}
