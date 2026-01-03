package integration

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

// RFR-002: Refresh unchanged (hash match)
func TestSourceRefresh_Unchanged_Skipped(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	page := startHTMLStub()
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
		cfg.OPENAI_API_KEY = "test-key"
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	// Allow localhost URLs for SSRF validation in tests
	te.SourcesHandlers.SSRFValidator.SetAllowPrivate(true)

	// Apply migration for refresh columns
	_, _ = te.DB.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_refreshed_at TIMESTAMPTZ`)
	_, _ = te.DB.Exec(`ALTER TABLE usage_ingestions ADD COLUMN IF NOT EXISTS refresh_count INT DEFAULT 0`)

	// Create pro user with refresh enabled
	token := authToken(t, te.Server.URL, "refresh_hash@example.com")
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{refresh}', '{"enabled": true, "max_monthly": 10}'::jsonb) WHERE code = 'free'`)

	// Create chatbot
	create := map[string]any{"name": "Hash Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add URL source
	var b strings.Builder
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()
	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(b.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := testHTTPClient().Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	sourceID := sid["id"]

	// Wait for source to complete
	statusPath := "/api/v1/sources/" + url.PathEscape(sourceID)
	waitForStatus(t, te.Server.URL+statusPath, token, "completed")

	// Get user ID
	var botUserID string
	te.DB.QueryRow(`SELECT user_id FROM chatbots WHERE id=$1`, bot.ID).Scan(&botUserID)

	// Get initial token usage
	var initialTokens int
	te.DB.QueryRow(`SELECT COALESCE(SUM(embedding_tokens), 0) FROM usage_ingestions WHERE user_id=$1`, botUserID).Scan(&initialTokens)

	// Refresh the source (content hasn't changed)
	refreshPath := "/api/v1/sources/" + url.PathEscape(sourceID) + "/refresh"
	reqR, _ := http.NewRequest(http.MethodPost, te.Server.URL+refreshPath, nil)
	reqR.Header.Set("Authorization", "Bearer "+token)
	resR, _ := testHTTPClient().Do(reqR)
	if resR.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resR.StatusCode)
	}
	resR.Body.Close()

	// Wait for refresh to complete
	waitForStatus(t, te.Server.URL+statusPath, token, "completed")

	// Check token usage again
	var finalTokens int
	te.DB.QueryRow(`SELECT COALESCE(SUM(embedding_tokens), 0) FROM usage_ingestions WHERE user_id=$1`, botUserID).Scan(&finalTokens)

	// Since content is unchanged, tokens should NOT increase (or increase by 0)
	// Note: Refresh count increases, but embedding tokens should not if skipped.
	if finalTokens > initialTokens {
		t.Errorf("expected tokens to remain %d, got %d (skipped processing failed)", initialTokens, finalTokens)
	}
}

func waitForStatus(t *testing.T, url, token, status string) {
	t.Helper()
	for i := 0; i < 100; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, url, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := testHTTPClient().Do(reqG)
		if resG.StatusCode != http.StatusOK {
			resG.Body.Close()
			time.Sleep(50 * time.Millisecond)
			continue
		}
		var st map[string]any
		json.NewDecoder(resG.Body).Decode(&st)
		resG.Body.Close()
		if st["status"].(string) == status {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for status %s", status)
}
