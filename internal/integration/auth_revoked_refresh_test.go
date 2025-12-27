package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAuth_RevokedRefreshToken_401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "revoke+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	reg := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	rb, _ := json.Marshal(reg)
	resReg, _ := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
	if resReg.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resReg.StatusCode)
	}
	var tp tokenPair
	json.NewDecoder(resReg.Body).Decode(&tp)
	resReg.Body.Close()
	if tp.RefreshToken == "" {
		t.Fatalf("missing refresh token")
	}

	// Manually revoke the refresh token in DB (lookup by hash)
	h := sha256.Sum256([]byte(tp.RefreshToken))
	tokenHash := hex.EncodeToString(h[:])
	_, err = te.DB.Exec("UPDATE refresh_tokens SET revoked=true WHERE token_hash=$1", tokenHash)
	if err != nil {
		t.Fatalf("db revoke failed: %v", err)
	}

	// Attempt refresh with revoked token → expect 401
	rr := refreshReq{RefreshToken: tp.RefreshToken}
	rrj, _ := json.Marshal(rr)
	resRef, _ := http.Post(te.Server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(rrj))
	if resRef.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resRef.StatusCode)
	}
	resRef.Body.Close()
}
