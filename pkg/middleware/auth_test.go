package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/onurceri/botla-co/internal/auth"
)

func TestAuthMiddleware_MissingHeader(t *testing.T) {
    mw := AuthMiddleware("secret")
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized { t.Fatalf("expected 401, got %d", rr.Code) }
}

func TestAuthMiddleware_InvalidTokenType(t *testing.T) {
    mw := AuthMiddleware("secret")
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    tok, _ := auth.GenerateToken("secret", "u1", "refresh", time.Minute)
    req.Header.Set("Authorization", "Bearer "+tok)
    mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized { t.Fatalf("expected 401, got %d", rr.Code) }
}

func TestAuthMiddleware_ValidAccess(t *testing.T) {
    mw := AuthMiddleware("secret")
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    tok, _ := auth.GenerateToken("secret", "u1", "access", time.Minute)
    req.Header.Set("Authorization", "Bearer "+tok)
    mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
    if rr.Code != http.StatusOK { t.Fatalf("expected 200, got %d", rr.Code) }
}

