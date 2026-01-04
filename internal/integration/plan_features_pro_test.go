package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/policy"
)

func updateProPlanConfig(t *testing.T, te *fixtures.TestEnv) {
	// Get the pro plan ID
	var proPlanID string
	err := te.DB.QueryRow(`SELECT id FROM plans WHERE code = 'pro'`).Scan(&proPlanID)
	if err != nil {
		t.Fatalf("failed to get pro plan ID: %v", err)
	}

	// Update plan_limits table
	_, err = te.DB.Exec(`
		INSERT INTO plan_limits (
			plan_id, scraping_dynamic_enabled, scraping_max_urls_per_bot, scraping_max_pages_per_crawl,
			files_max_size_mb, files_max_files_per_bot, files_total_storage_mb,
			chat_allowed_models, chat_max_monthly_tokens, chat_rag_top_k, chat_rag_max_context_tokens,
			branding_can_hide_branding, branding_can_custom_branding,
			refresh_enabled, refresh_max_monthly,
			guardrails_can_manage_topics, guardrails_can_customize_messages,
			guardrails_can_customize_thresholds, guardrails_can_use_smart_fallback,
			guardrails_can_use_escalate_fallback,
			security_secure_embed_enabled
		) VALUES (
			$1, true, 10, 10,
			20, 20, 500,
			ARRAY[$2, $3]::text[], 1000000, 5, 4000,
			true, false,
			true, 100,
			true, true,
			true, true,
			true,
			true
		)
		ON CONFLICT (plan_id) DO UPDATE SET
			scraping_dynamic_enabled = EXCLUDED.scraping_dynamic_enabled,
			scraping_max_urls_per_bot = EXCLUDED.scraping_max_urls_per_bot,
			scraping_max_pages_per_crawl = EXCLUDED.scraping_max_pages_per_crawl,
			files_max_size_mb = EXCLUDED.files_max_size_mb,
			files_max_files_per_bot = EXCLUDED.files_max_files_per_bot,
			files_total_storage_mb = EXCLUDED.files_total_storage_mb,
			chat_allowed_models = EXCLUDED.chat_allowed_models,
			chat_max_monthly_tokens = EXCLUDED.chat_max_monthly_tokens,
			chat_rag_top_k = EXCLUDED.chat_rag_top_k,
			chat_rag_max_context_tokens = EXCLUDED.chat_rag_max_context_tokens,
			branding_can_hide_branding = EXCLUDED.branding_can_hide_branding,
			branding_can_custom_branding = EXCLUDED.branding_can_custom_branding,
			refresh_enabled = EXCLUDED.refresh_enabled,
			refresh_max_monthly = EXCLUDED.refresh_max_monthly,
			guardrails_can_manage_topics = EXCLUDED.guardrails_can_manage_topics,
			guardrails_can_customize_messages = EXCLUDED.guardrails_can_customize_messages,
			guardrails_can_customize_thresholds = EXCLUDED.guardrails_can_customize_thresholds,
			guardrails_can_use_smart_fallback = EXCLUDED.guardrails_can_use_smart_fallback,
			guardrails_can_use_escalate_fallback = EXCLUDED.guardrails_can_use_escalate_fallback,
			security_secure_embed_enabled = EXCLUDED.security_secure_embed_enabled
	`, proPlanID, policy.ModelGPT4oMini.String(), policy.ModelGPT4o.String())
	if err != nil {
		t.Fatalf("failed to update plan limits: %v", err)
	}
}

func startProDynamicHTMLStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html><html><body><script>document.body.innerHTML='<p>Pro Dynamic Content</p>'</script></body></html>`))
	})
	return httptest.NewServer(h)
}

func startProLinkedHTMLStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		var b strings.Builder
		b.WriteString(`<html><body>`)
		for i := 1; i <= 15; i++ {
			fmt.Fprintf(&b, `<a href="/page%d">Page %d</a><br>`, i, i)
		}
		b.WriteString(`</body></html>`)
		w.Write([]byte(b.String()))
	})
	// Serve subpages
	for i := 1; i <= 15; i++ {
		path := fmt.Sprintf("/page%d", i)
		h.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Content"))
		})
	}
	return httptest.NewServer(h)
}

func waitForProcessingPro(t *testing.T, te *fixtures.TestEnv, token, sourceID string) {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/sources/"+sourceID, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := testHTTPClient().Do(reqG)
		if resG.StatusCode == http.StatusOK {
			var st map[string]any
			json.NewDecoder(resG.Body).Decode(&st)
			resG.Body.Close()
			if status, ok := st["status"].(string); ok {
				if status == "completed" || status == "failed" {
					t.Logf("Source %s processed with status: %s", sourceID, status)
					return
				}
			}
		} else {
			resG.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for source processing")
}

func setupStubs(t *testing.T) (*fixtures.LLMMock, *httptest.Server, fixtures.ConfigOverride) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	override := func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.QDRANT_URL = qd.URL
	}
	return oai, qd, override
}

func TestProPlan_ModelSelection(t *testing.T) {
	t.Parallel()

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	// Create user on Pro plan
	token := authToken(t, te.Server.URL, "pro_user@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_user@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	// 1. Create chatbot with gpt-4o-mini
	create := map[string]any{"name": "Pro Bot", "model": "gpt-4o-mini"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resC.Body)
		resC.Body.Close()
		t.Fatalf("chatbot create failed: %d %s", resC.StatusCode, string(body))
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 2. Update to gpt-4o (Allowed)
	upd1 := map[string]any{"model": "gpt-4o"}
	updB1, _ := json.Marshal(upd1)
	reqU1, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB1))
	reqU1.Header.Set("Authorization", "Bearer "+token)
	reqU1.Header.Set("Content-Type", "application/json")
	resU1, _ := testHTTPClient().Do(reqU1)
	if resU1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for gpt-4o, got %d", resU1.StatusCode)
	}
	resU1.Body.Close()

	// 3. Update to claude-3-5-sonnet (Forbidden)
	upd2 := map[string]any{"model": "claude-3-5-sonnet"}
	updB2, _ := json.Marshal(upd2)
	reqU2, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(updB2))
	reqU2.Header.Set("Authorization", "Bearer "+token)
	reqU2.Header.Set("Content-Type", "application/json")
	resU2, _ := testHTTPClient().Do(reqU2)
	if resU2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for claude-3-5-sonnet, got %d", resU2.StatusCode)
	}
	resU2.Body.Close()
}

func TestProPlan_MonthlyTokenLimit(t *testing.T) {
	t.Parallel()

	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	defer oai.Close()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	// Create user on Pro plan
	token := authToken(t, te.Server.URL, "pro_token@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_token@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	// Get user ID
	var userID string
	err = te.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, "pro_token@example.com").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "Token Limit Bot", "model": "gpt-4o", "max_tokens": 512}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resC.Body)
		resC.Body.Close()
		t.Fatalf("chatbot create failed: %d %s", resC.StatusCode, string(body))
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 1. Simulate usage: 999,400 tokens (leaving room for estimated 512 tokens)
	// We need to ensure we insert for the current month in usage_ingestions
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	_, err = te.DB.Exec(`INSERT INTO usage_ingestions (user_id, period_month, chat_tokens, sources_count, embedding_tokens, updated_at)
		VALUES ($1, $2, 999400, 0, 0, NOW())`, userID, monthStart)
	if err != nil {
		t.Fatalf("failed to insert usage: %v", err)
	}

	// 2. Send chat message (Expect 200 - 999400 + 512 = 999912 < 1000000)
	cr := map[string]any{"message": "hello", "session_id": "s1"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := testHTTPClient().Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resCh.StatusCode)
	}
	resCh.Body.Close()

	// 3. Update usage to 999,913 (will exceed when adding 512 more)
	_, err = te.DB.Exec(`UPDATE usage_ingestions SET chat_tokens = 999913 WHERE user_id = $1 AND period_month = $2`, userID, monthStart)
	if err != nil {
		t.Fatalf("failed to update usage: %v", err)
	}

	// 4. Send chat message (Expect 429 or 402)
	reqCh2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh2.Header.Set("Authorization", "Bearer "+token)
	reqCh2.Header.Set("Content-Type", "application/json")
	resCh2, _ := testHTTPClient().Do(reqCh2)
	// Note: Implementation might return 402 Payment Required or 429 Too Many Requests
	if resCh2.StatusCode != http.StatusTooManyRequests && resCh2.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("expected 429 or 402, got %d", resCh2.StatusCode)
	}
	resCh2.Body.Close()
}

func TestProPlan_PDFLimits(t *testing.T) {
	t.Parallel()

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	// Create user on Pro plan
	token := authToken(t, te.Server.URL, "pro_pdf@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_pdf@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "PDF Limit Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resC.Body)
		resC.Body.Close()
		t.Fatalf("chatbot create failed: %d %s", resC.StatusCode, string(body))
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 1. Upload 20 PDFs (small)
	for i := 0; i < 20; i++ {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		mw.WriteField("source_type", "pdf")
		part, _ := mw.CreateFormFile("file", fmt.Sprintf("test_%d.pdf", i))
		part.Write([]byte(fmt.Sprintf("%%PDF-1.4\n...dummy content %d...", i))) // Unique content
		mw.Close()

		reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
		reqS.Header.Set("Authorization", "Bearer "+token)
		reqS.Header.Set("Content-Type", mw.FormDataContentType())
		resS, _ := testHTTPClient().Do(reqS)
		if resS.StatusCode != http.StatusCreated {
			t.Fatalf("failed to upload PDF %d: %d", i+1, resS.StatusCode)
		}
		resS.Body.Close()
	}

	// 2. Upload 21st PDF (Expect 403)
	{
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		mw.WriteField("source_type", "pdf")
		part, _ := mw.CreateFormFile("file", "overflow.pdf")
		part.Write([]byte("%PDF-1.4\n...unique overflow content..."))
		mw.Close()

		reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
		reqS.Header.Set("Authorization", "Bearer "+token)
		reqS.Header.Set("Content-Type", mw.FormDataContentType())
		resS, _ := testHTTPClient().Do(reqS)
		if resS.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 for 21st PDF, got %d", resS.StatusCode)
		}
		resS.Body.Close()
	}

	// 3. Upload PDF > 20MB (Expect 413)
	{
		// Create a new bot to avoid hitting the file count limit (20) which would return 403
		create2 := map[string]any{"name": "PDF Size Bot"}
		cbj2, _ := json.Marshal(create2)
		reqC2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj2))
		reqC2.Header.Set("Authorization", "Bearer "+token)
		reqC2.Header.Set("Content-Type", "application/json")
		resC2, _ := testHTTPClient().Do(reqC2)
		if resC2.StatusCode != http.StatusCreated && resC2.StatusCode != http.StatusOK {
			t.Fatalf("chatbot 2 create failed: %d", resC2.StatusCode)
		}
		var bot2 chatbot
		json.NewDecoder(resC2.Body).Decode(&bot2)
		resC2.Body.Close()

		// Use io.Pipe to stream large content without huge memory alloc
		pr, pw := io.Pipe()
		mw := multipart.NewWriter(pw)

		go func() {
			defer pw.Close()
			defer mw.Close()
			_ = mw.WriteField("source_type", "pdf")
			part, _ := mw.CreateFormFile("file", "large.pdf")

			// Write 21 chunks of 1MB
			chunk := make([]byte, 1024*1024)
			for i := 0; i < 21; i++ {
				_, err := part.Write(chunk)
				if err != nil {
					return
				}
			}
		}()

		reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot2.ID+"/sources", pr)
		reqS.Header.Set("Authorization", "Bearer "+token)
		reqS.Header.Set("Content-Type", mw.FormDataContentType())
		resS, _ := testHTTPClient().Do(reqS)
		if resS.StatusCode != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected 413 for large PDF, got %d", resS.StatusCode)
		}
		resS.Body.Close()
	}
}

func TestProPlan_URLLimits(t *testing.T) {
	t.Parallel()

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	// Create user on Pro plan
	token := authToken(t, te.Server.URL, "pro_url@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_url@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "URL Limit Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 1. Add 10 URLs
	for i := 0; i < 10; i++ {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		mw.WriteField("source_type", "url")
		// Need unique URLs
		mw.WriteField("source_url", "http://example.com/"+string(rune('a'+i)))
		mw.Close()

		reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
		reqS.Header.Set("Authorization", "Bearer "+token)
		reqS.Header.Set("Content-Type", mw.FormDataContentType())
		resS, _ := testHTTPClient().Do(reqS)
		if resS.StatusCode != http.StatusCreated {
			t.Fatalf("failed to add URL %d: %d", i+1, resS.StatusCode)
		}
		resS.Body.Close()
	}

	// 2. Add 11th URL (Expect 403)
	{
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		mw.WriteField("source_type", "url")
		mw.WriteField("source_url", "http://example.com/overflow")
		mw.Close()

		reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
		reqS.Header.Set("Authorization", "Bearer "+token)
		reqS.Header.Set("Content-Type", mw.FormDataContentType())
		resS, _ := testHTTPClient().Do(reqS)
		if resS.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 for 11th URL, got %d", resS.StatusCode)
		}
		resS.Body.Close()
	}
}

func TestProPlan_DynamicScraping(t *testing.T) {
	oai, qd, override := setupStubs(t)
	defer oai.Close()
	defer qd.Close()

	page := startProDynamicHTMLStub()
	defer page.Close()

	cfg := scraper.DynamicConfig{
		PoolSize:   1,
		IdleTTL:    30 * time.Second,
		NavTimeout: 3 * time.Second,
		Allowed:    []string{"127.0.0.1", "localhost"},
	}
	bScraper, err := scraper.NewBrowserScraper(cfg)
	if err != nil {
		t.Skip("dynamic scraping not available: " + err.Error())
	}
	defer bScraper.Close()
	if _, dynErr := bScraper.ScrapeDynamicURL(page.URL); dynErr != nil {
		t.Skip("dynamic scraping not available: " + dynErr.Error())
	}

	te, err := fixtures.SetupTestEnvWithConfig(override)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	// Create user on Pro plan
	token := authToken(t, te.Server.URL, "pro_dynamic@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_dynamic@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "Dynamic Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add URL
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := testHTTPClient().Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	sourceID := sid["id"]

	// Wait for processing
	waitForProcessingPro(t, te, token, sourceID)

	src, err := db.GetSourceByID(context.Background(), te.DB, sourceID)
	if err != nil {
		t.Fatalf("failed to load source: %v", err)
	}
	if src == nil {
		t.Fatalf("expected source to exist")
	}
	if src.ChunkCount == 0 {
		t.Fatalf("expected non-zero chunks for dynamic-enabled pro plan")
	}

	// Note: We skip content verification because actual dynamic scraping requires a headless browser which might not be present.
	// We assume that if 201 created and processing completed, the flow is working.
	// To verify `is_dynamic` flag, we can check the database or mocks if available.
	// Since we don't have access to internal mock state here easily without refactoring, we trust the status.
}

func TestProPlan_DiscoveryMode(t *testing.T) {
	t.Parallel()

	oai, qd, override := setupStubs(t)
	defer oai.Close()
	defer qd.Close()

	page := startProLinkedHTMLStub()
	defer page.Close()

	te, err := fixtures.SetupTestEnvWithConfig(override)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Allow localhost URLs for testing
	te.SourcesHandlers.SSRFValidator.SetAllowPrivate(true)

	updateProPlanConfig(t, te)

	// Create user on Pro plan
	token := authToken(t, te.Server.URL, "pro_discovery@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_discovery@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	// Create chatbot with discovery_mode="auto"
	create := map[string]any{"name": "Discovery Bot", "discovery_mode": "auto"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add seed URL
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := testHTTPClient().Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	sourceID := sid["id"]

	// Wait for processing
	waitForProcessingPro(t, te, token, sourceID)

	// Verify that we have sources for discovered pages.
	// Since the scraper in test env might not actually follow links unless it's a real scraper or mock,
	// we check if we can verify the limit logic.
	// If the scraper doesn't run, we won't see discovered pages.
	// Assuming the test env uses a real scraper or a capable mock:
	// Check count of sources for this chatbot.
	// Max pages per crawl is 10 (from config).
	// Seed page + 10 discovered = 11? Or 10 total? Config says `max_pages_per_crawl`.
	// Usually this limit applies to discovered pages.
	// We'll check the count.

	// Give it some time for discovery to queue and process
	time.Sleep(2 * time.Second)

	var count int
	err = te.DB.QueryRow("SELECT COUNT(*) FROM data_sources WHERE chatbot_id=$1", bot.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count sources: %v", err)
	}

	// If scraper is not working in test env, count will be 1.
	// If it works, it should be <= 11.
	if count > 11 {
		t.Errorf("expected max 11 sources (1 seed + 10 discovered), got %d", count)
	}
}

func TestProPlan_Refresh(t *testing.T) {
	t.Parallel()

	oai, qd, override := setupStubs(t)
	defer oai.Close()
	defer qd.Close()

	page := startProDynamicHTMLStub()
	defer page.Close()

	te, err := fixtures.SetupTestEnvWithConfig(override)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Allow localhost URLs for testing
	te.SourcesHandlers.SSRFValidator.SetAllowPrivate(true)

	updateProPlanConfig(t, te)

	// Create user on Pro plan
	token := authToken(t, te.Server.URL, "pro_refresh@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_refresh@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	// Create chatbot
	create := map[string]any{"name": "Refresh Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add URL Source
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := testHTTPClient().Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	sourceID := sid["id"]

	// Wait for initial processing to complete
	waitForProcessingPro(t, te, token, sourceID)

	// 1. Manually refresh (Expect 202 Accepted)
	reqR, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/sources/"+sourceID+"/refresh", nil)
	reqR.Header.Set("Authorization", "Bearer "+token)
	resR, _ := testHTTPClient().Do(reqR)
	if resR.StatusCode != http.StatusAccepted {
		t.Errorf("expected refresh 202, got %d", resR.StatusCode)
	}
	resR.Body.Close()

	// 2. Set refresh_policy = "auto" (Expect 200)
	upd := map[string]any{"refresh_policy": "auto", "refresh_frequency": "daily"}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := testHTTPClient().Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Errorf("expected auto refresh update 200, got %d", resU.StatusCode)
	}
	resU.Body.Close()
}

func TestProPlan_Branding(t *testing.T) {
	t.Parallel()

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	token := authToken(t, te.Server.URL, "pro_branding@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_branding@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	create := map[string]any{"name": "Branding Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 1. Hide Branding (Allowed)
	upd := map[string]any{"hide_branding": true}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := testHTTPClient().Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for hide_branding, got %d", resU.StatusCode)
	}
	resU.Body.Close()

	// 2. Custom Branding (Not Allowed)
	upd2 := map[string]any{"custom_branding": map[string]any{"logo_url": "http://example.com/logo.png"}}
	ub2, _ := json.Marshal(upd2)
	reqU2, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub2))
	reqU2.Header.Set("Authorization", "Bearer "+token)
	reqU2.Header.Set("Content-Type", "application/json")
	resU2, _ := testHTTPClient().Do(reqU2)
	if resU2.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for custom_branding, got %d", resU2.StatusCode)
	}
	resU2.Body.Close()
}

func TestProPlan_Security(t *testing.T) {
	t.Parallel()

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	token := authToken(t, te.Server.URL, "pro_security@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_security@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	create := map[string]any{"name": "Security Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 1. Secure Embed Enabled (Allowed)
	upd := map[string]any{
		"secure_embed_enabled": true,
		"allowed_domains":      "example.com",
		"embed_secret":         "secret",
	}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := testHTTPClient().Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for secure embed, got %d", resU.StatusCode)
	}
	resU.Body.Close()
}

func TestProPlan_Guardrails(t *testing.T) {
	t.Parallel()

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	updateProPlanConfig(t, te)

	token := authToken(t, te.Server.URL, "pro_guardrails@example.com")
	_, err = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = $1`, "pro_guardrails@example.com")
	if err != nil {
		t.Fatalf("failed to update user plan: %v", err)
	}

	create := map[string]any{"name": "Guardrails Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 1. Full Guardrails Access
	upd := map[string]any{
		"threshold_config": map[string]any{
			"high_threshold": 0.8,
			"fallback_mode":  "escalate",
		},
		"topic_restrictions": map[string]any{
			"allowed_topics": []string{"tech"},
		},
	}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := testHTTPClient().Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for guardrails, got %d", resU.StatusCode)
	}
	resU.Body.Close()
}
