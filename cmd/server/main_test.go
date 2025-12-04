package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/onurceri/botla-co/internal/api/handlers"
)

func TestChatbotsDispatchHandler_Routes(t *testing.T) {
    ch := &handlers.ChatbotHandlers{}
    sh := &handlers.SourcesHandlers{}
    chh := &handlers.ChatHandlers{}
    h := chatbotsDispatchHandler("secret", ch, sh, chh)

    cases := []struct{ path string; code int }{
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
