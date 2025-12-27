package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAuth_Register_SQLInjectionEmail(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "'; DROP TABLE users;--"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "SQL Injection"}
	b, err := json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for SQL injection email, got %d", res.StatusCode)
	}

	var totalUsers int
	if err = te.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers); err != nil {
		t.Fatalf("users table query failed (integrity check): %v", err)
	}

	var count int
	if err = te.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email=$1", email).Scan(&count); err != nil && err != sql.ErrNoRows {
		t.Fatalf("query for SQL injection email failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no user created with SQL injection email, got %d", count)
	}
}

func TestAuth_Register_XSSFullName(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "xss+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	fullName := "<script>alert('xss')</script>"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": fullName}
	b, err := json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", res.StatusCode)
	}

	var tr tokenResp
	if err = json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	res.Body.Close()
	if tr.Token == "" {
		t.Fatalf("token empty")
	}

	req, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("failed to create /me request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tr.Token)

	meRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("me request failed: %v", err)
	}
	defer meRes.Body.Close()

	if meRes.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /me, got %d", meRes.StatusCode)
	}

	var meBody struct {
		FullName string `json:"full_name"`
	}
	if err = json.NewDecoder(meRes.Body).Decode(&meBody); err != nil {
		t.Fatalf("failed to decode /me response: %v", err)
	}

	if meBody.FullName == "" {
		t.Fatalf("expected full_name in /me response")
	}
	if strings.Contains(meBody.FullName, "<script") || strings.Contains(meBody.FullName, "</script>") {
		t.Fatalf("expected sanitized full_name without raw script tags, got %q", meBody.FullName)
	}
}
