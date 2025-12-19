package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAuth_InvalidAccessToken_Me401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestAuth_TamperedAccessToken_Me401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer header.payload.signaturex")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestAuth_MissingAuthorizationHeader_Me401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestAuth_ValidAccessToken_Me200(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "me-access+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User"}
	rb, err := json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal register body failed: %v", err)
	}
	resReg, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resReg.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resReg.StatusCode)
	}
	var tokens struct {
		Token string `json:"token"`
	}
	if err = json.NewDecoder(resReg.Body).Decode(&tokens); err != nil {
		t.Fatalf("decode register response failed: %v", err)
	}
	resReg.Body.Close()
	if tokens.Token == "" {
		t.Fatalf("expected non-empty access token")
	}

	req, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Token)
	resMe, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("me request failed: %v", err)
	}
	if resMe.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resMe.StatusCode)
	}
	var me struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err = json.NewDecoder(resMe.Body).Decode(&me); err != nil {
		t.Fatalf("decode me response failed: %v", err)
	}
	resMe.Body.Close()
	if me.ID == "" {
		t.Fatalf("expected non-empty user id")
	}
	if me.Email != email {
		t.Fatalf("expected email %s, got %s", email, me.Email)
	}
}
