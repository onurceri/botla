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

func TestAuth_RefreshWithExpiredToken(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "refexp+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	// Register to get a user ID
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "Ref Exp User"}
	b, _ := json.Marshal(regBody)
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	// We need the user ID to generate a token manually.
	// The response has tokens but not explicit user ID, but we can decode the token or just fetch user from DB.
	// Or we can just trust the token generation if we had the ID.
	// Let's get the user ID from DB.
	var userID string
	err = te.DB.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}
	res.Body.Close()

	// Generate an expired refresh token
	expiredRefresh, err := auth.GenerateToken(te.Cfg.JWT_SECRET, userID, false, "refresh", -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Try to refresh with it
	rr := map[string]string{"refresh_token": expiredRefresh}
	rrj, _ := json.Marshal(rr)
	resRef, err := http.Post(te.Server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(rrj))
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	if resRef.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired refresh token, got %d", resRef.StatusCode)
	}
	resRef.Body.Close()
}
