package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/auth"
)

func TestAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	mw := AuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc")
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_BearerButTamperedToken(t *testing.T) {
	mw := AuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	tok, _ := auth.GenerateToken("secret", "u1", false, "access", time.Minute)
	req.Header.Set("Authorization", "Bearer "+tok+"x")
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
