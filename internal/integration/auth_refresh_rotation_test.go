package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type tokenPair struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func TestAuth_RefreshRotationAndLogout(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// register & login
	email := "rot+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	rb, _ := json.Marshal(regBody)
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

	// refresh: use initial refresh to get new pair and revoke old
	rr := refreshReq{RefreshToken: tp.RefreshToken}
	rrj, _ := json.Marshal(rr)
	resRef1, _ := http.Post(te.Server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(rrj))
	if resRef1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resRef1.StatusCode)
	}
	var tp2 tokenPair
	json.NewDecoder(resRef1.Body).Decode(&tp2)
	resRef1.Body.Close()
	if tp2.RefreshToken == "" {
		t.Fatalf("missing new refresh token")
	}
	if tp2.RefreshToken == tp.RefreshToken {
		t.Fatalf("expected rotated refresh token to differ from original")
	}

	// second refresh with old token should fail (revoked)
	resRefOld, _ := http.Post(te.Server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(rrj))
	if resRefOld.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked, got %d", resRefOld.StatusCode)
	}
	resRefOld.Body.Close()

	// logout current refresh and then try refresh → should 401
	lr := refreshReq{RefreshToken: tp2.RefreshToken}
	lrj, _ := json.Marshal(lr)
	resLogout, _ := http.Post(te.Server.URL+"/api/v1/auth/logout", "application/json", bytes.NewReader(lrj))
	if resLogout.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on logout, got %d", resLogout.StatusCode)
	}
	resLogout.Body.Close()

	resRef3, _ := http.Post(te.Server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(lrj))
	if resRef3.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", resRef3.StatusCode)
	}
	resRef3.Body.Close()
}
