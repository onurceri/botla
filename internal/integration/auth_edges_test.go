package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAuth_EdgeCases(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// method not allowed
	res, _ := http.Get(te.Server.URL + "/api/v1/auth/register")
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.StatusCode)
	}
	res.Body.Close()

	// bad json
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/auth/register", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	r1, _ := http.DefaultClient.Do(req)
	if r1.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", r1.StatusCode)
	}
	r1.Body.Close()

	// missing fields
	body := map[string]string{"email": " ", "password": " ", "full_name": " "}
	b, _ := json.Marshal(body)
	r2, _ := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if r2.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", r2.StatusCode)
	}
	r2.Body.Close()

	// invalid login
	lb := map[string]string{"email": "nouser@example.com", "password": "x"}
	lbj, _ := json.Marshal(lb)
	r3, _ := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	if r3.StatusCode != http.StatusUnauthorized && r3.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 401 or 500, got %d", r3.StatusCode)
	}
	r3.Body.Close()
}

func TestAuth_Register_SQLInjectionEmailSafe(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	var before int
	err = te.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&before)
	if err != nil {
		t.Fatalf("count before failed: %v", err)
	}

	body := map[string]string{
		"email":     `'; DROP TABLE users;--`,
		"password":  "testpassword123",
		"full_name": "SQL Injection Test",
	}
	b, _ := json.Marshal(body)
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if res.StatusCode == http.StatusInternalServerError {
		t.Fatalf("expected non-500 status, got %d", res.StatusCode)
	}
	res.Body.Close()

	var after int
	err = te.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&after)
	if err != nil {
		t.Fatalf("count after failed: %v", err)
	}
	if after < before {
		t.Fatalf("users table appears corrupted: before=%d after=%d", before, after)
	}
}
