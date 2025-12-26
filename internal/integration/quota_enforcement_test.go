package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/policy"
)

// Helper to get user ID from email
func getUserIdFromToken(t *testing.T, pool *sql.DB, email string) string {
	var id string
	err := pool.QueryRow("SELECT id FROM users WHERE email=$1", email).Scan(&id)
	if err != nil {
		t.Fatalf("failed to get user id for %s: %v", email, err)
	}
	return id
}

// QTA-001: Chat when monthly tokens exceeded
func TestQuota_ChatTokensExceeded(t *testing.T) {
	// Setup with mock OpenAI to avoid real calls
	oai := NewLLMMock(t)
	defer oai.Close()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	qd := startQdrantStub()
	defer qd.Close()
	t.Setenv("QDRANT_URL", qd.URL)

	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Update free plan to have very low token limit
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '100'::jsonb) WHERE code=$1`, policy.PlanFree.String())

	token := authToken(t, te.Server.URL, "chatquota@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "chatquota@example.com")

	// Create chatbot
	create := map[string]any{"name": "Chat Quota Bot", "max_tokens": 50}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// First chat under quota should succeed
	firstReq := chatReq{Message: "Hello before limit", SessionID: "sess-ok"}
	firstBody, _ := json.Marshal(firstReq)
	reqOK, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(firstBody))
	reqOK.Header.Set("Authorization", "Bearer "+token)
	reqOK.Header.Set("Content-Type", "application/json")
	resOK, _ := http.DefaultClient.Do(reqOK)
	if resOK.StatusCode != http.StatusOK {
		t.Fatalf("QTA-001: expected first chat 200 OK, got %d", resOK.StatusCode)
	}
	var okResp chatResp
	json.NewDecoder(resOK.Body).Decode(&okResp)
	resOK.Body.Close()
	if okResp.TokensUsed <= 0 {
		t.Fatalf("QTA-001: expected tokens_used > 0, got %d", okResp.TokensUsed)
	}

	// Manually insert usage to exceed limit
	// We need 150 tokens to exceed 100 limit
	// Use IncrementChatTokens to update usage_ingestions (authoritative source for quota)
	userID := getUserIdFromToken(t, te.DB, "chatquota@example.com")
	err = db.IncrementChatTokens(context.Background(), te.DB, userID, 150)
	if err != nil {
		t.Fatalf("failed to increment usage: %v", err)
	}

	// Try chat after limit exceeded
	overReq := chatReq{Message: "Hello over limit", SessionID: "sess1"}
	cb, _ := json.Marshal(overReq)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)

	if res.StatusCode != http.StatusPaymentRequired {
		t.Errorf("QTA-001: expected 402 Payment Required, got %d", res.StatusCode)
	}

	var errResp map[string]any
	json.NewDecoder(res.Body).Decode(&errResp)
	res.Body.Close()

	if code, ok := errResp["code"].(string); !ok || code != "ERR_MONTHLY_TOKENS_EXCEEDED" {
		t.Errorf("expected error code ERR_MONTHLY_TOKENS_EXCEEDED, got %v", errResp)
	}
}

// QTA-003: Refresh when monthly limit exceeded
func TestQuota_RefreshExceeded(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Update free plan to have refresh enabled and low limit
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{refresh,enabled}', 'true'::jsonb) WHERE code=$1`, policy.PlanFree.String())
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{refresh,max_monthly}', '1'::jsonb) WHERE code=$1`, policy.PlanFree.String())

	token := authToken(t, te.Server.URL, "refreshquota@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "refreshquota@example.com")
	userID := getUserIdFromToken(t, te.DB, "refreshquota@example.com")

	// Manually increment refresh count (manual refresh uses refresh_count column)
	err = db.IncrementRefreshCount(context.Background(), te.DB, userID, time.Now())
	if err != nil {
		t.Fatalf("failed to increment refresh: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "Refresh Quota Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add a dummy source directly to DB to refresh
	var sourceID string
	err = te.DB.QueryRow(`INSERT INTO data_sources (chatbot_id, source_type, source_url, hash, status) VALUES ($1, 'url', 'http://example.com', 'hash', 'processed') RETURNING id`, bot.ID).Scan(&sourceID)
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Try refresh
	reqR, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/sources/"+sourceID+"/refresh", nil)
	reqR.Header.Set("Authorization", "Bearer "+token)
	resR, _ := http.DefaultClient.Do(reqR)

	if resR.StatusCode != http.StatusPaymentRequired {
		t.Errorf("QTA-003: expected 402, got %d", resR.StatusCode)
	}

	var errResp map[string]any
	json.NewDecoder(resR.Body).Decode(&errResp)
	resR.Body.Close()

	if code, ok := errResp["code"].(string); !ok || code != "ERR_MONTHLY_REFRESH_EXCEEDED" {
		t.Errorf("expected error code ERR_MONTHLY_REFRESH_EXCEEDED, got %v", errResp)
	}
}

// QTA-002: Ingestion when monthly limit exceeded
func TestQuota_IngestionExceeded(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Update free plan to have low ingestion limit (count)
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{max_monthly_ingestions}', '1'::jsonb) WHERE code=$1`, policy.PlanFree.String())

	token := authToken(t, te.Server.URL, "ingestionquota@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "ingestionquota@example.com")
	userID := getUserIdFromToken(t, te.DB, "ingestionquota@example.com")

	// Manually increment sources count
	err = db.IncrementSuccessfulIngestion(context.Background(), te.DB, userID, time.Now(), 1)
	if err != nil {
		t.Fatalf("failed to increment ingestion: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "Ingestion Quota Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Try to add a source using Multipart Form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	_ = w.WriteField("source_type", "text")
	_ = w.WriteField("text", "This is a test content.")
	w.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", w.FormDataContentType())
	resS, _ := http.DefaultClient.Do(reqS)

	// Note: If ingestion quota is exceeded, it should return 402 Payment Required
	// or 403 Forbidden depending on implementation. checkIngestionQuota returns quotaError which usually maps to 402.
	// But let's check what checkIngestionQuota returns.
	// In source_create.go: http.Error(w, err.Error(), http.StatusPaymentRequired)
	if resS.StatusCode != http.StatusPaymentRequired {
		t.Errorf("QTA-002: expected 402, got %d", resS.StatusCode)
	}

	// The body is plain text from http.Error? Or JSON?
	// source_create.go uses http.Error(w, err.Error(), ...) which returns text/plain.
	// It does NOT use WriteLocalizedError.
	// So we can't check JSON code "ERR_MONTHLY_INGESTION_EXCEEDED".
	// We should just check status code or text.
	// However, for consistency, let's just check status code first.
}

// QTA-004: Race condition check for token quota (Double-Spend)
func TestQuota_RaceCondition_DoubleSpend(t *testing.T) {
	// Setup mocks
	oai := NewLLMMock(t)
	defer oai.Close()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	qd := startQdrantStub()
	defer qd.Close()
	t.Setenv("QDRANT_URL", qd.URL)

	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// 1. Configure Plan with specific limit
	// Limit: 1000 tokens
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '1000'::jsonb) WHERE code=$1`, policy.PlanFree.String())

	// 2. Create User & Token
	email := "race_quota@example.com"
	token := authToken(t, te.Server.URL, email)
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), email)
	userID := getUserIdFromToken(t, te.DB, email)

	// 3. Create Chatbot with specific max_tokens
	create := map[string]any{
		"name":       "Race Quota Bot",
		"max_tokens": 40,
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	if bot.ID == "" {
		t.Fatal("failed to create bot")
	}

	// 4. Set Initial State: 900 used
	err = db.IncrementChatTokens(context.Background(), te.DB, userID, 900)
	if err != nil {
		t.Fatalf("failed to set initial tokens: %v", err)
	}

	// 5. Run Concurrent Requests
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	responses := make([]int, concurrency)

	chatBody := chatReq{Message: "Race test", SessionID: "sess-race"}
	chatBytes, _ := json.Marshal(chatBody)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(chatBytes))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			responses[idx] = res.StatusCode
			res.Body.Close()
		}(i)
	}

	wg.Wait()

	// 6. Analyze Results
	successCount := 0
	for _, code := range responses {
		if code == http.StatusOK {
			successCount++
		}
	}

	if successCount > 2 {
		t.Errorf("Race condition detected! Expected max 2 successes, got %d", successCount)
	}
	if successCount == 0 {
		t.Errorf("Too few successes, expected around 2, got 0")
	}
}
