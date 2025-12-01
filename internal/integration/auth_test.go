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
	Token string `json:"token"`
}

func TestAuth_Register_Login_Protected(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "user1+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User One"}
	b, _ := json.Marshal(regBody)
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var tr tokenResp
	_ = json.NewDecoder(res.Body).Decode(&tr)
	res.Body.Close()
	if tr.Token == "" {
		t.Fatalf("token empty")
	}

	// re-register conflict
	res2, _ := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if res2.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res2.StatusCode)
	}
	res2.Body.Close()

	// login
	loginBody := map[string]string{"email": email, "password": "pass1234"}
	lb, _ := json.Marshal(loginBody)
	res3, _ := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res3.StatusCode)
	}
	var tr2 tokenResp
	_ = json.NewDecoder(res3.Body).Decode(&tr2)
	res3.Body.Close()
	if tr2.Token == "" {
		t.Fatalf("login token empty")
	}

	// protected without token
	res4, _ := http.Get(te.Server.URL + "/api/v1/protected")
	if res4.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res4.StatusCode)
	}
	res4.Body.Close()

	// protected with token
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tr2.Token)
	res5, _ := http.DefaultClient.Do(req)
	if res5.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res5.StatusCode)
	}
	res5.Body.Close()
}
