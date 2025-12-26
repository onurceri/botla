package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAuth_Register_WeakPassword(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "weak+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	// Attempt to register with a password shorter than 8 characters
	regBody := map[string]string{"email": email, "password": "short", "full_name": "Weak User"}
	b, _ := json.Marshal(regBody)
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	// We expect 400 Bad Request
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for weak password, got %d", res.StatusCode)
	}
	var errResp struct {
		Code   string `json:"code"`
		Status int    `json:"status"`
	}
	_ = json.NewDecoder(res.Body).Decode(&errResp)
	res.Body.Close()
	if errResp.Code != "ERR_PASSWORD_TOO_SHORT" {
		t.Errorf("expected 'ERR_PASSWORD_TOO_SHORT', got %v", errResp.Code)
	}
}
