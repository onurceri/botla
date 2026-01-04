package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/policy"
)

// TRK-002: Turkish special chars in chatbot name
func TestTurkish_ChatbotName(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "turkish-name@example.com")

	name := "Türkçe Chatbot Şğıöüç"
	create := map[string]any{"name": name}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)

	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create failed: %d", resC.StatusCode)
	}

	var bot struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	if bot.Name != name {
		t.Errorf("TRK-002: expected name %q, got %q", name, bot.Name)
	}
}

// TRK-004: Turkish chars in system prompt
func TestTurkish_SystemPrompt(t *testing.T) {
	t.Parallel()
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.OPENAI_API_KEY = "test-key"
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "turkish-prompt@example.com")

	// Create chatbot
	create := map[string]any{"name": "Prompt Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Update prompt
	prompt := "Her zaman Türkçe konuş. Şğıöüç."
	update := map[string]any{"custom_instruction": prompt}
	upj, _ := json.Marshal(update)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(upj))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := testHTTPClient().Do(reqU)

	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	// Verify in DB
	var storedPrompt string
	err = te.DB.QueryRow("SELECT custom_instruction FROM chatbots WHERE id=$1", bot.ID).Scan(&storedPrompt)
	if err != nil {
		t.Fatalf("failed to query prompt: %v", err)
	}
	if storedPrompt != prompt {
		t.Errorf("TRK-004: expected prompt %q, got %q", prompt, storedPrompt)
	}
}

// TRK-001: Turkish special chars in user message
func TestTurkish_UserMessage(t *testing.T) {
	t.Parallel()
	// Setup with mock OpenAI to verify prompt sent to LLM
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.OPENAI_API_KEY = "test-key"
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "turkish-msg@example.com")

	// Create chatbot
	create := map[string]any{"name": "Message Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	msg := "Merhaba dünya. Şğıöüç."
	chatReq := map[string]string{"message": msg, "session_id": "sess-trk"}
	cb, _ := json.Marshal(chatReq)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := testHTTPClient().Do(req)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("chat failed: %d", res.StatusCode)
	}
	drainBody(res)

	// Verify in DB
	var content string
	err = te.DB.QueryRow("SELECT content FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE chatbot_id=$1) AND role='user' ORDER BY created_at DESC LIMIT 1", bot.ID).Scan(&content)
	if err != nil {
		t.Fatalf("failed to query message: %v", err)
	}
	if content != msg {
		t.Errorf("TRK-001: expected message %q, got %q", msg, content)
	}
}

// TRK-010: Error ERR_MONTHLY_TOKENS_EXCEEDED
func TestTurkish_LocalizedError(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Update free plan to have very low token limit
	_ = te.UpdatePlanLimit(policy.PlanFree.String(), "chat_max_monthly_tokens", 100)

	email := "error-loc@example.com"
	token := authToken(t, te.Server.URL, email)
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), email)

	// Get user ID for quota increment
	var userID string
	err = te.DB.QueryRow("SELECT id FROM users WHERE email=$1", email).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "Error Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Increment chat tokens to exceed limit (use proper quota function)
	err = db.IncrementChatTokens(context.Background(), te.DB, userID, 500)
	if err != nil {
		t.Fatalf("failed to increment tokens: %v", err)
	}

	// Try chat
	chatReq := map[string]string{"message": "Hello", "session_id": "sess1"}
	cb, _ := json.Marshal(chatReq)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	// Set Accept-Language to tr-TR
	req.Header.Set("Accept-Language", "tr-TR")
	res, _ := testHTTPClient().Do(req)

	if res.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("TRK-010: expected 402, got %d", res.StatusCode)
	}

	var errResp map[string]any
	json.NewDecoder(res.Body).Decode(&errResp)
	drainBody(res)

	// API uses WriteErrorCode which returns {"code": "...", "status": ...}
	// The error message localization is handled by frontend, not backend
	code, _ := errResp["code"].(string)
	expected := "ERR_MONTHLY_TOKENS_EXCEEDED"
	if code != expected {
		t.Errorf("TRK-010: expected code %q, got %q (full response: %v)", expected, code, errResp)
	}
}

// TRK-007: JSON encoding of Turkish chars
func TestTurkish_JSONEncoding(t *testing.T) {
	// This test verifies that the API doesn't escape unicode characters in JSON response
	// Go's encoding/json escapes HTML chars by default (<, >, &), but usually not others unless configured?
	// Actually json.Marshal escapes by default? No, it produces UTF-8.
	// But sometimes SetEscapeHTML(true) is used.
	// We want to ensure we get "Ş" not "\u015e".

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "json-enc@example.com")

	name := "Şğıöüç"
	create := map[string]any{"name": name}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)

	// Read raw body
	buf := new(bytes.Buffer)
	buf.ReadFrom(resC.Body)
	raw := buf.String()
	resC.Body.Close()

	if !bytes.Contains(buf.Bytes(), []byte(name)) {
		// If raw contains unicode escape sequence, it fails this test check (assuming name has special chars)
		// But verify if it's actually escaped or not.
		// If it's escaped, we might see \u015e
		// TRK-007 expectation: "No escaped unicode"
		t.Errorf("TRK-007: expected raw JSON to contain %q, got %s", name, raw)
	}
}
