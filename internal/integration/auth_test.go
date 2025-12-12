package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type tokenResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func TestAuth_Register_Login_Protected(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "user1+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User One"}
	var b []byte
	b, err = json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var tr tokenResp
	if err = json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Fatalf("decode register response failed: %v", err)
	}
	res.Body.Close()
	if tr.Token == "" {
		t.Fatalf("token empty")
	}
	if tr.RefreshToken == "" {
		t.Fatalf("refresh token empty")
	}

	res2, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("re-register failed: %v", err)
	}
	if res2.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res2.StatusCode)
	}
	var errResp map[string]string
	if err = json.NewDecoder(res2.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode conflict response failed: %v", err)
	}
	res2.Body.Close()
	if errResp["error"] != "Email already exists" {
		t.Fatalf("expected 'Email already exists', got %v", errResp["error"])
	}

	loginBody := map[string]string{"email": email, "password": "pass1234"}
	lb, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body failed: %v", err)
	}
	res3, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res3.StatusCode)
	}
	var tr2 tokenResp
	if err = json.NewDecoder(res3.Body).Decode(&tr2); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	res3.Body.Close()
	if tr2.Token == "" {
		t.Fatalf("login token empty")
	}
	if tr2.RefreshToken == "" {
		t.Fatalf("login refresh token empty")
	}

	res4, err := http.Get(te.Server.URL + "/api/v1/protected")
	if err != nil {
		t.Fatalf("protected without token request failed: %v", err)
	}
	if res4.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res4.StatusCode)
	}
	res4.Body.Close()

	req, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/protected", nil)
	if err != nil {
		t.Fatalf("create protected request failed: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tr2.Token)
	res5, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("protected with token request failed: %v", err)
	}
	if res5.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res5.StatusCode)
	}
	res5.Body.Close()
}

func TestAuth_Register_InvalidEmailFormat(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	regBody := map[string]string{"email": "invalid-email", "password": "pass1234", "full_name": "Invalid Email"}
	var b []byte
	b, err = json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}

	var errResp map[string]string
	if err = json.NewDecoder(res.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if errResp["error"] != "Invalid email format" {
		t.Fatalf("expected 'Invalid email format', got %v", errResp["error"])
	}
}

func TestAuth_Register_MissingEmail(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	regBody := map[string]string{"password": "pass1234", "full_name": "Missing Email"}
	var b []byte
	b, err = json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}

	var errResp map[string]string
	if err = json.NewDecoder(res.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if errResp["error"] != "Email is required" {
		t.Fatalf("expected 'Email is required', got %v", errResp["error"])
	}
}

func TestAuth_Register_MissingPassword(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "nopassword+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "full_name": "Missing Password"}
	var b []byte
	b, err = json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}

	var errResp map[string]string
	if err = json.NewDecoder(res.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if errResp["error"] != "Password must be at least 8 characters long" {
		t.Fatalf("expected 'Password must be at least 8 characters long', got %v", errResp["error"])
	}
}
