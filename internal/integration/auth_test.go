package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/auth"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

type tokenResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// errResp matches the new API error response format
type errResp struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
}

func hashTokenForTest(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func TestAuth_Register_Login_Protected(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := "user1+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User One"}
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
	var er errResp
	if err = json.NewDecoder(res2.Body).Decode(&er); err != nil {
		t.Fatalf("decode conflict response failed: %v", err)
	}
	res2.Body.Close()
	if er.Code != "ERR_EMAIL_EXISTS" {
		t.Fatalf("expected 'ERR_EMAIL_EXISTS', got %v", er.Code)
	}

	loginBody := map[string]string{"email": email, "password": "Test@123"}
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

	var userID string
	if err = te.DB.QueryRow("SELECT id FROM users WHERE email=$1", email).Scan(&userID); err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}
	claims, err := auth.VerifyToken(te.Cfg.JWT_SECRET, tr2.Token)
	if err != nil {
		t.Fatalf("failed to verify access token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected user id %s in token, got %s", userID, claims.UserID)
	}
	if claims.ExpiresAt == nil {
		t.Fatalf("expected exp claim on access token")
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		t.Fatalf("expected access token expiry in the future, got %v", claims.ExpiresAt.Time)
	}
	if ttl < 55*time.Minute || ttl > 65*time.Minute {
		t.Fatalf("expected access token TTL around 60m, got %v", ttl)
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
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	regBody := map[string]string{"email": "invalid-email", "password": "Test@123", "full_name": "Invalid Email"}
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

	var er errResp
	if err = json.NewDecoder(res.Body).Decode(&er); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if er.Code != "ERR_INVALID_EMAIL_FORMAT" {
		t.Fatalf("expected 'ERR_INVALID_EMAIL_FORMAT', got %v", er.Code)
	}
}

func TestAuth_Register_MissingEmail(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	regBody := map[string]string{"password": "Test@123", "full_name": "Missing Email"}
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

	var er errResp
	if err = json.NewDecoder(res.Body).Decode(&er); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if er.Code != "ERR_EMAIL_REQUIRED" {
		t.Fatalf("expected 'ERR_EMAIL_REQUIRED', got %v", er.Code)
	}
}

func TestAuth_Register_MissingPassword(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

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

	var er errResp
	if err = json.NewDecoder(res.Body).Decode(&er); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if er.Code != "ERR_PASSWORD_TOO_SHORT" {
		t.Fatalf("expected 'ERR_PASSWORD_TOO_SHORT', got %v", er.Code)
	}
}

func TestAuth_Login_InvalidEmail(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	loginBody := map[string]string{"email": "doesnotexist+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com", "password": "Test@123"}
	lb, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body failed: %v", err)
	}
	res, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}

	var er errResp
	if err = json.NewDecoder(res.Body).Decode(&er); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if er.Code != "ERR_INVALID_CREDENTIALS" {
		t.Fatalf("expected 'ERR_INVALID_CREDENTIALS', got %v", er.Code)
	}
}

func TestAuth_Login_InvalidPassword(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := "login-bad-pass+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User One"}
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
	resReg.Body.Close()

	loginBody := map[string]string{"email": email, "password": "wrongpass"}
	lb, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body failed: %v", err)
	}
	resLogin, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resLogin.Body.Close()

	if resLogin.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resLogin.StatusCode)
	}
	var er errResp
	if err = json.NewDecoder(resLogin.Body).Decode(&er); err != nil {
		t.Fatalf("decode error response failed: %v", err)
	}
	if er.Code != "ERR_INVALID_CREDENTIALS" {
		t.Fatalf("expected 'ERR_INVALID_CREDENTIALS', got %v", er.Code)
	}
}

func TestAuth_Login_EmptyEmail(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	loginBody := map[string]string{"email": "", "password": "Test@123"}
	lb, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body failed: %v", err)
	}
	res, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestAuth_Login_EmptyPassword(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	loginBody := map[string]string{"email": "empty-pass+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com", "password": ""}
	lb, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body failed: %v", err)
	}
	res, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestAuth_Login_CaseInsensitiveEmail(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	baseEmail := fmt.Sprintf("TestCI+%d@Example.com", time.Now().UnixNano())
	regBody := map[string]string{"email": baseEmail, "password": "Test@123", "full_name": "Case User"}
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
	resReg.Body.Close()

	lowerEmail := strings.ToLower(baseEmail)
	upperEmail := strings.ToUpper(baseEmail)

	loginBodyLower := map[string]string{"email": lowerEmail, "password": "Test@123"}
	lbLower, err := json.Marshal(loginBodyLower)
	if err != nil {
		t.Fatalf("marshal lower login body failed: %v", err)
	}
	resLower, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lbLower))
	if err != nil {
		t.Fatalf("login lower failed: %v", err)
	}
	if resLower.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for lower-case login, got %d", resLower.StatusCode)
	}
	resLower.Body.Close()

	loginBodyUpper := map[string]string{"email": upperEmail, "password": "Test@123"}
	lbUpper, err := json.Marshal(loginBodyUpper)
	if err != nil {
		t.Fatalf("marshal upper login body failed: %v", err)
	}
	resUpper, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lbUpper))
	if err != nil {
		t.Fatalf("login upper failed: %v", err)
	}
	if resUpper.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for upper-case login, got %d", resUpper.StatusCode)
	}
	resUpper.Body.Close()
}

func TestAuth_Login_MultipleSessions(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := fmt.Sprintf("multisession+%d@example.com", time.Now().UnixNano())
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "Multi Session User"}
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
	resReg.Body.Close()

	loginBody := map[string]string{"email": email, "password": "Test@123"}
	lb, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body failed: %v", err)
	}

	resA, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login A failed: %v", err)
	}
	if resA.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for login A, got %d", resA.StatusCode)
	}
	var tokensA tokenResp
	if err = json.NewDecoder(resA.Body).Decode(&tokensA); err != nil {
		t.Fatalf("decode login A response failed: %v", err)
	}
	resA.Body.Close()

	resB, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login B failed: %v", err)
	}
	if resB.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for login B, got %d", resB.StatusCode)
	}
	var tokensB tokenResp
	if err = json.NewDecoder(resB.Body).Decode(&tokensB); err != nil {
		t.Fatalf("decode login B response failed: %v", err)
	}
	resB.Body.Close()

	if tokensA.RefreshToken == "" || tokensB.RefreshToken == "" {
		t.Fatalf("expected non-empty refresh tokens for both sessions")
	}
	if tokensA.RefreshToken == tokensB.RefreshToken {
		t.Fatalf("expected different refresh tokens for different sessions")
	}

	hashA := hashTokenForTest(tokensA.RefreshToken)
	hashB := hashTokenForTest(tokensB.RefreshToken)

	var revokedA bool
	if err = te.DB.QueryRow("SELECT revoked FROM refresh_tokens WHERE token_hash=$1", hashA).Scan(&revokedA); err != nil {
		t.Fatalf("query refresh token A failed: %v", err)
	}
	if revokedA {
		t.Fatalf("expected refresh token A to be valid (not revoked)")
	}

	var revokedB bool
	if err = te.DB.QueryRow("SELECT revoked FROM refresh_tokens WHERE token_hash=$1", hashB).Scan(&revokedB); err != nil {
		t.Fatalf("query refresh token B failed: %v", err)
	}
	if revokedB {
		t.Fatalf("expected refresh token B to be valid (not revoked)")
	}

	logoutReq := map[string]string{"refresh_token": tokensA.RefreshToken}
	logoutBody, err := json.Marshal(logoutReq)
	if err != nil {
		t.Fatalf("marshal logout body failed: %v", err)
	}
	resLogout, err := http.Post(te.Server.URL+"/api/v1/auth/logout", "application/json", bytes.NewReader(logoutBody))
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	if resLogout.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for logout, got %d", resLogout.StatusCode)
	}
	resLogout.Body.Close()

	if err = te.DB.QueryRow("SELECT revoked FROM refresh_tokens WHERE token_hash=$1", hashA).Scan(&revokedA); err != nil {
		t.Fatalf("re-query refresh token A failed: %v", err)
	}
	if !revokedA {
		t.Fatalf("expected refresh token A to be revoked after logout")
	}

	if err = te.DB.QueryRow("SELECT revoked FROM refresh_tokens WHERE token_hash=$1", hashB).Scan(&revokedB); err != nil {
		t.Fatalf("re-query refresh token B failed: %v", err)
	}
	if revokedB {
		t.Fatalf("expected refresh token B to remain valid after logout of A")
	}
}

func TestAuth_Login_RefreshTokenTracking(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := "refresh-track+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User One"}
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
	var regTokens tokenResp
	if err = json.NewDecoder(resReg.Body).Decode(&regTokens); err != nil {
		t.Fatalf("decode register response failed: %v", err)
	}
	resReg.Body.Close()

	loginBody := map[string]string{"email": email, "password": "Test@123"}
	lb, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body failed: %v", err)
	}
	resLogin, err := http.Post(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(lb))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resLogin.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resLogin.StatusCode)
	}
	var loginTokens tokenResp
	if err = json.NewDecoder(resLogin.Body).Decode(&loginTokens); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	resLogin.Body.Close()

	if loginTokens.RefreshToken == "" {
		t.Fatalf("expected non-empty refresh token from login")
	}

	var userID string
	if err = te.DB.QueryRow("SELECT id FROM users WHERE email=$1", email).Scan(&userID); err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}

	tokenHash := hashTokenForTest(loginTokens.RefreshToken)
	var dbUserID string
	var expiresAt time.Time
	if err = te.DB.QueryRow("SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash=$1", tokenHash).Scan(&dbUserID, &expiresAt); err != nil {
		t.Fatalf("failed to query refresh token: %v", err)
	}
	if dbUserID != userID {
		t.Fatalf("expected refresh token user_id %s, got %s", userID, dbUserID)
	}
	if expiresAt.Before(time.Now()) {
		t.Fatalf("expected refresh token expiry in the future, got %v", expiresAt)
	}
	ttl := time.Until(expiresAt)
	if ttl < 6*24*time.Hour || ttl > 8*24*time.Hour {
		t.Fatalf("expected refresh token TTL around 7d, got %v", ttl)
	}
}
