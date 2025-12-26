package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/onurceri/botla-co/pkg/policy"
)

func TestPublic_SecureEmbed_Enforcement(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Create chatbot
	token := authToken(t, te.Server.URL, "secure_owner@example.com")

	// Enable secure embed for free plan (or upgrade user) to allow testing the feature logic
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code=$1`, policy.PlanFree.String())
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	secret := "my-secret-key-123"
	create := map[string]any{
		"name": "Secure Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update to enable secure embed
	upd := map[string]any{
		"secure_embed_enabled": true,
		"embed_secret":         secret,
		"allowed_domains":      "example.com",
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

	// Refresh bot struct to be sure? Not strictly needed as we have ID.
	// But let's verify it persisted if we were paranoid. The other test covers persistence.

	// 1. Request without token -> Should fail
	cr := chatReq{Message: "hi", SessionID: "s1"}
	crb, _ := json.Marshal(cr)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Content-Type", "application/json")
	// Set Origin to allowed domain
	req1.Header.Set("Origin", "https://example.com")
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", res1.StatusCode)
	}
	res1.Body.Close()

	// 2. Request with invalid token -> Should fail
	req2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Origin", "https://example.com")
	req2.Header.Set("X-Embed-Token", "invalid")
	res2, _ := http.DefaultClient.Do(req2)
	if res2.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", res2.StatusCode)
	}
	res2.Body.Close()

	// 3. Request with valid token -> Should pass
	validToken := generateEmbedToken(t, secret, bot.ID)
	req3, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Origin", "https://example.com")
	req3.Header.Set("X-Embed-Token", validToken)
	res3, _ := http.DefaultClient.Do(req3)
	if res3.StatusCode != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", res3.StatusCode)
	}
	res3.Body.Close()

	// 4. Request with valid token but wrong origin -> Should fail (if domain restriction is implemented)
	req4, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Origin", "https://evil.com")
	req4.Header.Set("X-Embed-Token", validToken)
	res4, _ := http.DefaultClient.Do(req4)
	if res4.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 with wrong origin, got %d", res4.StatusCode)
	}
	res4.Body.Close()
}

func generateEmbedToken(t *testing.T, secret, botID string) string {
	claims := jwt.MapClaims{
		"chatbot_id": botID,
		"exp":        time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign error: %v", err)
	}
	return s
}
