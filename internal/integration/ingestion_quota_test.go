package integration

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/policy"
)

func TestMonthlyIngestionQuota_AndDuplicateURL(t *testing.T) {
	t.Parallel()

	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	page := startHTMLStub()
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Allow localhost URLs for testing
	te.SourcesHandlers.SSRFValidator.SetAllowPrivate(true)

	// Set plan config with low ingestion quota and cooldown
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_build_object(
        'max_monthly_ingestions', 1,
        'max_monthly_embedding_tokens', 250000,
        'min_readd_cooldown_minutes', 60,
        'scraping', jsonb_build_object('dynamic_enabled', false, 'max_urls_per_bot', 5, 'max_pages_per_crawl', 0),
        'files', jsonb_build_object('max_size_mb', 10, 'max_files_per_bot', 5, 'max_files_total', 100, 'total_storage_mb', 100),
        'chat', jsonb_build_object('allowed_models', jsonb_build_array('gpt-4o-mini'), 'max_monthly_tokens', 100000, 'rag', jsonb_build_object('top_k',3,'max_context_tokens',1024))
    ) WHERE code=$1`, policy.PlanFree.String())

	token := authToken(t, te.Server.URL, "quota@example.com")
	// Ensure user is on free plan
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "quota@example.com")

	// Create chatbot
	create := map[string]any{"name": "Quota Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// First URL ingest (should succeed)
	var body1 strings.Builder
	mw1 := multipart.NewWriter(&body1)
	mw1.WriteField("source_type", "url")
	mw1.WriteField("source_url", page.URL)
	mw1.Close()
	reqS1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body1.String()))
	reqS1.Header.Set("Authorization", "Bearer "+token)
	reqS1.Header.Set("Content-Type", mw1.FormDataContentType())
	resS1, _ := testHTTPClient().Do(reqS1)
	if resS1.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resS1.StatusCode)
	}
	resS1.Body.Close()

	// Duplicate URL ingest should 409
	var bodyDup strings.Builder
	mwD := multipart.NewWriter(&bodyDup)
	mwD.WriteField("source_type", "url")
	mwD.WriteField("source_url", page.URL)
	mwD.Close()
	reqDup, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(bodyDup.String()))
	reqDup.Header.Set("Authorization", "Bearer "+token)
	reqDup.Header.Set("Content-Type", mwD.FormDataContentType())
	resDup, _ := testHTTPClient().Do(reqDup)
	if resDup.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resDup.StatusCode)
	}
	resDup.Body.Close()

	// List sources and wait until the first is processed to ensure monthly counter increments
	reqL, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", nil)
	reqL.Header.Set("Authorization", "Bearer "+token)
	resL, _ := testHTTPClient().Do(reqL)
	var items []map[string]any
	json.NewDecoder(resL.Body).Decode(&items)
	resL.Body.Close()
	sid := items[0]["id"].(string)
	// poll status
	for i := 0; i < 200; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/sources/"+sid, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := testHTTPClient().Do(reqG)
		if resG.StatusCode == http.StatusOK {
			var st map[string]any
			json.NewDecoder(resG.Body).Decode(&st)
			resG.Body.Close()
			s := st["status"].(string)
			if s != "pending" && s != "processing" {
				break
			}
		} else {
			resG.Body.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}

	reqDel, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/sources/"+sid, nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	resDel, _ := testHTTPClient().Do(reqDel)
	if resDel.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resDel.StatusCode)
	}
	resDel.Body.Close()

	// Re-add same URL immediately should hit cooldown 429
	var body2 strings.Builder
	mw2 := multipart.NewWriter(&body2)
	mw2.WriteField("source_type", "url")
	mw2.WriteField("source_url", page.URL)
	mw2.Close()
	reqS2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body2.String()))
	reqS2.Header.Set("Authorization", "Bearer "+token)
	reqS2.Header.Set("Content-Type", mw2.FormDataContentType())
	resS2, _ := testHTTPClient().Do(reqS2)
	if resS2.StatusCode != http.StatusTooManyRequests && resS2.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("expected 429 or 402, got %d", resS2.StatusCode)
	}
	resS2.Body.Close()

	// Also, a second distinct URL should exceed monthly ingestion and return 402 since first completed counts
	var body3 strings.Builder
	mw3 := multipart.NewWriter(&body3)
	mw3.WriteField("source_type", "url")
	mw3.WriteField("source_url", page.URL+"/another")
	mw3.Close()
	reqS3, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body3.String()))
	reqS3.Header.Set("Authorization", "Bearer "+token)
	reqS3.Header.Set("Content-Type", mw3.FormDataContentType())
	resS3, _ := testHTTPClient().Do(reqS3)
	if resS3.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", resS3.StatusCode)
	}
	resS3.Body.Close()
}
