package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/auth"
	"github.com/onurceri/botla-app/internal/integration/fixtures"
)

type refreshTokenPair struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func TestAuth_Refresh_GeneratesNewAccessToken(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := "refresh-success+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	rb, err := json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal register body failed: %v", err)
	}
	resReg, err := testHTTPPost(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resReg.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resReg.StatusCode)
	}
	var regTokens refreshTokenPair
	if err = json.NewDecoder(resReg.Body).Decode(&regTokens); err != nil {
		t.Fatalf("decode register response failed: %v", err)
	}
	resReg.Body.Close()
	if regTokens.Token == "" || regTokens.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens from register")
	}

	reqBody := refreshRequest{RefreshToken: regTokens.RefreshToken}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal refresh body failed: %v", err)
	}
	resRef, err := testHTTPPost(te.Server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	if resRef.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resRef.StatusCode)
	}
	var refTokens refreshTokenPair
	if err = json.NewDecoder(resRef.Body).Decode(&refTokens); err != nil {
		t.Fatalf("decode refresh response failed: %v", err)
	}
	resRef.Body.Close()

	if refTokens.Token == "" {
		t.Fatalf("expected new access token")
	}
	if refTokens.RefreshToken == "" {
		t.Fatalf("expected new refresh token")
	}

	if refTokens.Token == regTokens.Token {
		t.Fatalf("expected access token to change after refresh")
	}

	_, err = auth.VerifyToken(te.Cfg.JWT_SECRET, refTokens.Token)
	if err != nil {
		t.Fatalf("expected new access token to be a valid JWT, got %v", err)
	}

	reqOld, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/protected", nil)
	reqOld.Header.Set("Authorization", "Bearer "+regTokens.Token)
	resOld, err := testHTTPClient().Do(reqOld)
	if err != nil {
		t.Fatalf("protected request with old access token failed: %v", err)
	}
	if resOld.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for old access token, got %d", resOld.StatusCode)
	}
	resOld.Body.Close()
}
