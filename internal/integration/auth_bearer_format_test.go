package integration

import (
    "net/http"
    "testing"
)

func TestAuth_InvalidBearerFormat_Protected401(t *testing.T) {
    te, err := SetupTestEnv()
    if err != nil { t.Fatalf("setup failed: %v", err) }
    defer TeardownTestEnv(te)

    req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/protected", nil)
    req.Header.Set("Authorization", "Token abc.def")
    res, _ := http.DefaultClient.Do(req)
    if res.StatusCode != http.StatusUnauthorized { t.Fatalf("expected 401, got %d", res.StatusCode) }
    res.Body.Close()
}

