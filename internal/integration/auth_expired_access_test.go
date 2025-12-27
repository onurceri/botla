package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/auth"
)

func TestAuth_ExpiredAccessToken_Me401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// register to obtain user id via protected ping
	email := "exp+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	rb, _ := json.Marshal(regBody)
	resReg, _ := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
	if resReg.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resReg.StatusCode)
	}
	resReg.Body.Close()

	// login to get valid token and confirm user id
	lb := map[string]string{"email": email, "password": "Test@123"}
	lbj, _ := json.Marshal(lb)
	resLog, _ := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	if resLog.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resLog.StatusCode)
	}
	var tr tokenResp
	json.NewDecoder(resLog.Body).Decode(&tr)
	resLog.Body.Close()

	// protected ping with valid token
	reqPing, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	reqPing.Header.Set("Authorization", "Bearer "+tr.Token)
	resPing, _ := http.DefaultClient.Do(reqPing)
	if resPing.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resPing.StatusCode)
	}
	resPing.Body.Close()

	expired, _ := auth.GenerateToken(te.Cfg.JWT_SECRET, "expired-user", false, "access", -time.Minute)
	reqExp, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	reqExp.Header.Set("Authorization", "Bearer "+expired)
	resExp, _ := http.DefaultClient.Do(reqExp)
	if resExp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resExp.StatusCode)
	}
	var errBody map[string]string
	if err := json.NewDecoder(resExp.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errBody["error"] != "Token expired" {
		t.Fatalf("expected error message %q, got %q", "Token expired", errBody["error"])
	}
	resExp.Body.Close()
}
