package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestSecureEmbed_DomainOnlyRestriction tests that domain restriction works without token
func TestSecureEmbed_DomainOnlyRestriction(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Enable secure embed for free plan
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code='free'`)
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	token := authToken(t, te.Server.URL, "domain_test@example.com")

	// Create chatbot
	create := map[string]any{"name": "Domain Only Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update: secure embed ON, allowed_domains set, but NO embed_secret
	upd := map[string]any{
		"secure_embed_enabled": true,
		"allowed_domains":      "allowed.com",
		// No embed_secret - token should NOT be required
	}
	updB, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)

	// Test 1: Request from allowed domain WITHOUT token -> Should PASS (no secret = no token needed)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://allowed.com")
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from allowed domain without token, got %d", res1.StatusCode)
	}
	res1.Body.Close()

	// Test 2: Request from disallowed domain -> Should FAIL with 403
	req2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Origin", "https://evil.com")
	res2, _ := http.DefaultClient.Do(req2)
	if res2.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 from disallowed domain, got %d", res2.StatusCode)
	}
	res2.Body.Close()

	// Test 3: Request without Origin header -> Should FAIL with 403
	req3, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req3.Header.Set("Content-Type", "application/json")
	// No Origin header
	res3, _ := http.DefaultClient.Do(req3)
	if res3.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 without Origin header, got %d", res3.StatusCode)
	}
	res3.Body.Close()
}

// TestSecureEmbed_TokenOnlyRestriction tests token validation without domain restriction
func TestSecureEmbed_TokenOnlyRestriction(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Enable secure embed for free plan
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code='free'`)
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	userToken := authToken(t, te.Server.URL, "token_only@example.com")
	secret := "token-only-secret-456"

	// Create chatbot
	create := map[string]any{"name": "Token Only Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+userToken)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update: secure embed ON, embed_secret set, but NO allowed_domains
	upd := map[string]any{
		"secure_embed_enabled": true,
		"embed_secret":         secret,
		// No allowed_domains - any origin should be ok if token is valid
	}
	updB, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB))
	reqU.Header.Set("Authorization", "Bearer "+userToken)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)

	// Test 1: Request WITHOUT token -> Should FAIL with 401
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://any-site.com")
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", res1.StatusCode)
	}
	res1.Body.Close()

	// Test 2: Request WITH valid token -> Should PASS
	validToken := generateValidToken(secret, bot.ID, time.Hour)
	req2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Origin", "https://any-site.com")
	req2.Header.Set("X-Embed-Token", validToken)
	res2, _ := http.DefaultClient.Do(req2)
	if res2.StatusCode != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

// TestSecureEmbed_ExpiredToken tests that expired tokens are rejected
func TestSecureEmbed_ExpiredToken(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Enable secure embed for free plan
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code='free'`)
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	userToken := authToken(t, te.Server.URL, "expired_test@example.com")
	secret := "expired-token-secret"

	// Create chatbot
	create := map[string]any{"name": "Expired Token Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+userToken)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update to enable token validation
	upd := map[string]any{
		"secure_embed_enabled": true,
		"embed_secret":         secret,
	}
	updB, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB))
	reqU.Header.Set("Authorization", "Bearer "+userToken)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)

	// Generate an expired token (expired 1 hour ago)
	expiredToken := generateValidToken(secret, bot.ID, -time.Hour)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://some-site.com")
	req1.Header.Set("X-Embed-Token", expiredToken)
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with expired token, got %d", res1.StatusCode)
	}
	res1.Body.Close()
}

// TestSecureEmbed_WrongChatbotIdInToken tests that tokens with wrong chatbot_id are rejected
func TestSecureEmbed_WrongChatbotIdInToken(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Enable secure embed for free plan
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code='free'`)
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	userToken := authToken(t, te.Server.URL, "wrong_id_test@example.com")
	secret := "wrong-id-secret"

	// Create chatbot
	create := map[string]any{"name": "Wrong ID Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+userToken)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update to enable token validation
	upd := map[string]any{
		"secure_embed_enabled": true,
		"embed_secret":         secret,
	}
	updB, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB))
	reqU.Header.Set("Authorization", "Bearer "+userToken)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)

	// Generate token with wrong chatbot_id
	wrongIdToken := generateValidToken(secret, "wrong-chatbot-id-12345", time.Hour)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://some-site.com")
	req1.Header.Set("X-Embed-Token", wrongIdToken)
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong chatbot_id in token, got %d", res1.StatusCode)
	}
	res1.Body.Close()
}

// TestSecureEmbed_WrongSecretSignature tests that tokens signed with wrong secret are rejected
func TestSecureEmbed_WrongSecretSignature(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Enable secure embed for free plan
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code='free'`)
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	userToken := authToken(t, te.Server.URL, "wrong_secret_test@example.com")
	correctSecret := "correct-secret-abc"
	wrongSecret := "wrong-secret-xyz"

	// Create chatbot
	create := map[string]any{"name": "Wrong Secret Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+userToken)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update to enable token validation with correct secret
	upd := map[string]any{
		"secure_embed_enabled": true,
		"embed_secret":         correctSecret,
	}
	updB, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB))
	reqU.Header.Set("Authorization", "Bearer "+userToken)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)

	// Generate token signed with WRONG secret
	wrongSignatureToken := generateValidToken(wrongSecret, bot.ID, time.Hour)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://some-site.com")
	req1.Header.Set("X-Embed-Token", wrongSignatureToken)
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong secret signature, got %d", res1.StatusCode)
	}
	res1.Body.Close()
}

// TestSecureEmbed_NoSecureEmbedEnabled tests that requests pass when secure embed is disabled
func TestSecureEmbed_NoSecureEmbedEnabled(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	userToken := authToken(t, te.Server.URL, "no_secure@example.com")

	// Create chatbot (secure embed is disabled by default)
	create := map[string]any{"name": "Non-Secure Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+userToken)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)

	// Request from any origin without token should work (secure embed disabled)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://any-site.com")
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusOK {
		t.Errorf("expected 200 when secure embed disabled, got %d", res1.StatusCode)
	}
	res1.Body.Close()
}

// TestSecureEmbed_DomainAndTokenCombined tests both domain AND token restrictions together
func TestSecureEmbed_DomainAndTokenCombined(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Enable secure embed for free plan
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code='free'`)
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	userToken := authToken(t, te.Server.URL, "combined_test@example.com")
	secret := "combined-secret-789"

	// Create chatbot
	create := map[string]any{"name": "Combined Restrictions Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+userToken)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update: BOTH domain AND token restrictions
	upd := map[string]any{
		"secure_embed_enabled": true,
		"allowed_domains":      "trusted.com,partner.com",
		"embed_secret":         secret,
	}
	updB, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB))
	reqU.Header.Set("Authorization", "Bearer "+userToken)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)
	validToken := generateValidToken(secret, bot.ID, time.Hour)

	// Test 1: Allowed domain + valid token -> PASS
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Origin", "https://trusted.com")
	req1.Header.Set("X-Embed-Token", validToken)
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusOK {
		t.Errorf("expected 200 with allowed domain and valid token, got %d", res1.StatusCode)
	}
	res1.Body.Close()

	// Test 2: Allowed domain + NO token -> FAIL (token required since secret is set)
	req2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Origin", "https://trusted.com")
	res2, _ := http.DefaultClient.Do(req2)
	if res2.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with allowed domain but no token, got %d", res2.StatusCode)
	}
	res2.Body.Close()

	// Test 3: Disallowed domain + valid token -> FAIL (domain check comes first)
	req3, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Origin", "https://evil.com")
	req3.Header.Set("X-Embed-Token", validToken)
	res3, _ := http.DefaultClient.Do(req3)
	if res3.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 with disallowed domain and valid token, got %d", res3.StatusCode)
	}
	res3.Body.Close()

	// Test 4: Disallowed domain + no token -> FAIL
	req4, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Origin", "https://evil.com")
	res4, _ := http.DefaultClient.Do(req4)
	if res4.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 with disallowed domain and no token, got %d", res4.StatusCode)
	}
	res4.Body.Close()
}

// Helper function to generate a valid JWT token
func generateValidToken(secret, botID string, expOffset time.Duration) string {
	claims := jwt.MapClaims{
		"chatbot_id": botID,
		"exp":        time.Now().Add(expOffset).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}
