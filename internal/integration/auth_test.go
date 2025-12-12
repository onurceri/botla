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

	"github.com/onurceri/botla-co/internal/auth"
)

type tokenResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func hashTokenForTest(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
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
	if ttl < 10*time.Minute || ttl > 20*time.Minute {
		t.Fatalf("expected access token TTL around 15m, got %v", ttl)
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

func TestAuth_Login_InvalidEmail(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	loginBody := map[string]string{"email": "doesnotexist+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com", "password": "pass1234"}
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
}

func TestAuth_Login_InvalidPassword(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "login-bad-pass+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User One"}
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
}

func TestAuth_Login_EmptyEmail(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	loginBody := map[string]string{"email": "", "password": "pass1234"}
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
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

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

func TestAuth_Login_RefreshTokenTracking(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "refresh-track+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User One"}
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

	loginBody := map[string]string{"email": email, "password": "pass1234"}
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
