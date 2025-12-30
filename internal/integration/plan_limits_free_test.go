package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/policy"
)

type chatbotBranding struct {
	ID           string `json:"id"`
	HideBranding bool   `json:"hide_branding"`
}

func startDynamicHTMLStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html><html><body><script>document.body.innerHTML='<p>Merhaba Dinamik</p>'</script></body></html>`))
	})
	return httptest.NewServer(h)
}

func TestFreePlan_URLLimit_PerChatbot(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	page := startHTMLStub()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	_, _ = te.DB.Exec(`UPDATE plans SET config = '{
  "scraping": {
    "dynamic_enabled": false,
    "max_urls_per_bot": 1,
    "max_pages_per_crawl": 0
  },
  "files": {
    "ocr_enabled": false,
    "max_size_mb": 5,
    "max_files_per_bot": 1,
    "total_storage_mb": 10
  },
  "chat": {
    "allowed_models": ["` + policy.ModelGPT4oMini.String() + `"],
    "max_monthly_tokens": 100000,
    "rag": {
      "top_k": 3,
      "max_context_tokens": 2000
    }
  }
}'::jsonb WHERE code = '` + policy.PlanFree.String() + `'`)

	token := authToken(t, te.Server.URL, "url_limit_free@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, "url_limit_free@example.com")

	create := map[string]any{"name": "URL Limit Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	var body1 strings.Builder
	mw1 := multipart.NewWriter(&body1)
	mw1.WriteField("source_type", "url")
	mw1.WriteField("source_url", page.URL+"/1")
	mw1.Close()

	reqS1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body1.String()))
	reqS1.Header.Set("Authorization", "Bearer "+token)
	reqS1.Header.Set("Content-Type", mw1.FormDataContentType())
	resS1, _ := http.DefaultClient.Do(reqS1)
	if resS1.StatusCode != http.StatusCreated {
		t.Fatalf("expected first URL 201, got %d", resS1.StatusCode)
	}
	resS1.Body.Close()

	var body2 strings.Builder
	mw2 := multipart.NewWriter(&body2)
	mw2.WriteField("source_type", "url")
	mw2.WriteField("source_url", page.URL+"/2")
	mw2.Close()

	reqS2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body2.String()))
	reqS2.Header.Set("Authorization", "Bearer "+token)
	reqS2.Header.Set("Content-Type", mw2.FormDataContentType())
	resS2, _ := http.DefaultClient.Do(reqS2)
	if resS2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected second URL 403, got %d", resS2.StatusCode)
	}

	var errResp map[string]any
	json.NewDecoder(resS2.Body).Decode(&errResp)
	resS2.Body.Close()

	if code, ok := errResp["code"].(string); !ok || code != "ERR_URL_LIMIT_REACHED" {
		t.Fatalf("expected error code ERR_URL_LIMIT_REACHED, got %v", errResp)
	}
}

func TestFreePlan_URLLimit_AllowsNewURLAfterDelete(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	page := startHTMLStub()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	_, _ = te.DB.Exec(`UPDATE plans SET config = '{
  "scraping": {
    "dynamic_enabled": false,
    "max_urls_per_bot": 1,
    "max_pages_per_crawl": 0
  },
  "files": {
    "ocr_enabled": false,
    "max_size_mb": 5,
    "max_files_per_bot": 1,
    "total_storage_mb": 10
  },
  "chat": {
    "allowed_models": ["` + policy.ModelGPT4oMini.String() + `"],
    "max_monthly_tokens": 100000,
    "rag": {
      "top_k": 3,
      "max_context_tokens": 2000
    }
  }
}'::jsonb WHERE code = '` + policy.PlanFree.String() + `'`)

	token := authToken(t, te.Server.URL, "url_limit_delete@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, "url_limit_delete@example.com")

	create := map[string]any{"name": "URL Limit Delete Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	var body1 strings.Builder
	mw1 := multipart.NewWriter(&body1)
	mw1.WriteField("source_type", "url")
	mw1.WriteField("source_url", page.URL+"/1")
	mw1.Close()

	reqS1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body1.String()))
	reqS1.Header.Set("Authorization", "Bearer "+token)
	reqS1.Header.Set("Content-Type", mw1.FormDataContentType())
	resS1, _ := http.DefaultClient.Do(reqS1)
	if resS1.StatusCode != http.StatusCreated {
		t.Fatalf("expected first URL 201, got %d", resS1.StatusCode)
	}
	var first map[string]string
	json.NewDecoder(resS1.Body).Decode(&first)
	resS1.Body.Close()
	sourceID := first["id"]

	reqDel, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/sources/"+sourceID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	resDel, _ := http.DefaultClient.Do(reqDel)
	if resDel.StatusCode != http.StatusNoContent {
		t.Fatalf("expected delete 204, got %d", resDel.StatusCode)
	}
	resDel.Body.Close()

	var body2 strings.Builder
	mw2 := multipart.NewWriter(&body2)
	mw2.WriteField("source_type", "url")
	mw2.WriteField("source_url", page.URL+"/2")
	mw2.Close()

	reqS2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body2.String()))
	reqS2.Header.Set("Authorization", "Bearer "+token)
	reqS2.Header.Set("Content-Type", mw2.FormDataContentType())
	resS2, _ := http.DefaultClient.Do(reqS2)
	if resS2.StatusCode != http.StatusCreated {
		t.Fatalf("expected second URL 201 after delete, got %d", resS2.StatusCode)
	}
	resS2.Body.Close()
}

func TestFreePlan_DynamicScraping_Disabled_StaticOnly(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	page := startDynamicHTMLStub()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	_, _ = te.DB.Exec(`UPDATE plans SET config = '{
  "scraping": {
    "dynamic_enabled": false,
    "max_urls_per_bot": 1,
    "max_pages_per_crawl": 0
  },
  "files": {
    "ocr_enabled": false,
    "max_size_mb": 5,
    "max_files_per_bot": 1,
    "total_storage_mb": 10
  },
  "chat": {
    "allowed_models": ["` + policy.ModelGPT4oMini.String() + `"],
    "max_monthly_tokens": 100000,
    "rag": {
      "top_k": 3,
      "max_context_tokens": 2000
    }
  }
}'::jsonb WHERE code = '` + policy.PlanFree.String() + `'`)

	email := "dynamic_free@example.com"
	token := authToken(t, te.Server.URL, email)
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, email)

	create := map[string]any{"name": "Dynamic Disabled Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	var body strings.Builder
	mw := multipart.NewWriter(&body)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := http.DefaultClient.Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected url source 201, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	sourceID := sid["id"]

	waitForProcessing(t, te, token, sourceID)

	src, err := db.GetSourceByID(context.Background(), te.DB, sourceID)
	if err != nil {
		t.Fatalf("failed to load source: %v", err)
	}
	if src == nil {
		t.Fatalf("expected source to exist")
	}
	if src.ChunkCount != 0 {
		t.Fatalf("expected zero chunks for dynamic-disabled free plan, got %d", src.ChunkCount)
	}
}

func TestFreePlan_DiscoveryMode_Disabled_OnZeroCrawlLimit(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	page := startLinkedHTMLStub()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	_, _ = te.DB.Exec(`UPDATE plans SET config = '{
  "scraping": {
    "dynamic_enabled": false,
    "max_urls_per_bot": 1,
    "max_pages_per_crawl": 0
  },
  "files": {
    "ocr_enabled": false,
    "max_size_mb": 5,
    "max_files_per_bot": 1,
    "total_storage_mb": 10
  },
  "chat": {
    "allowed_models": ["` + policy.ModelGPT4oMini.String() + `"],
    "max_monthly_tokens": 100000,
    "rag": {
      "top_k": 3,
      "max_context_tokens": 2000
    }
  }
}'::jsonb WHERE code = '` + policy.PlanFree.String() + `'`)

	token := authToken(t, te.Server.URL, "disc_free@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, "disc_free@example.com")

	create := map[string]any{
		"name":           "Free Discovery Bot",
		"discovery_mode": "auto",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	upd := map[string]any{"discovery_mode": "auto"}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusForbidden {
		t.Fatalf("expected discovery_mode update 403, got %d", resU.StatusCode)
	}
	resU.Body.Close()

	var body strings.Builder
	mw := multipart.NewWriter(&body)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := http.DefaultClient.Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected root source 201, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	rootSourceID := sid["id"]

	waitForProcessing(t, te, token, rootSourceID)
	time.Sleep(200 * time.Millisecond)

	sources, err := db.ListSourcesByChatbotID(context.Background(), te.DB, bot.ID)
	if err != nil {
		t.Fatalf("failed to list sources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected only root source, got %d", len(sources))
	}
}

func TestFreePlan_RefreshPolicyAuto_Forbidden(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{refresh}', '{"enabled": false, "max_monthly": 0}'::jsonb, true) WHERE code = '` + policy.PlanFree.String() + `'`)

	token := authToken(t, te.Server.URL, "refresh_policy_free@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, "refresh_policy_free@example.com")

	create := map[string]any{"name": "Refresh Policy Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	upd := map[string]any{"refresh_policy": "auto"}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusForbidden {
		t.Fatalf("expected refresh_policy update 403, got %d", resU.StatusCode)
	}

	var errResp map[string]any
	json.NewDecoder(resU.Body).Decode(&errResp)
	resU.Body.Close()

	if feature, ok := errResp["feature"].(string); !ok || feature != "auto_refresh" {
		t.Fatalf("expected feature auto_refresh, got %v", errResp)
	}
}

func TestFreePlan_Branding_HideBrandingForbidden(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "branding_free@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, "branding_free@example.com")

	create := map[string]any{"name": "Branding Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbotBranding
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	upd := map[string]any{"hide_branding": true}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusForbidden {
		t.Fatalf("expected hide_branding update 403, got %d", resU.StatusCode)
	}

	var errResp map[string]any
	json.NewDecoder(resU.Body).Decode(&errResp)
	resU.Body.Close()

	if feature, ok := errResp["feature"].(string); !ok || feature != "hide_branding" {
		t.Fatalf("expected feature hide_branding, got %v", errResp)
	}

	reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID, nil)
	reqG.Header.Set("Authorization", "Bearer "+token)
	resG, _ := http.DefaultClient.Do(reqG)
	if resG.StatusCode != http.StatusOK {
		t.Fatalf("expected get chatbot 200, got %d", resG.StatusCode)
	}

	var fetched chatbotBranding
	json.NewDecoder(resG.Body).Decode(&fetched)
	resG.Body.Close()

	if fetched.HideBranding {
		t.Fatalf("expected hide_branding to remain false")
	}
}

func TestFreePlan_SecureEmbed_Forbidden(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": false}'::jsonb, true) WHERE code = '` + policy.PlanFree.String() + `'`)

	token := authToken(t, te.Server.URL, "secure_free@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, "secure_free@example.com")

	create := map[string]any{"name": "Secure Embed Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	updSecure := map[string]any{"secure_embed_enabled": true}
	usb, _ := json.Marshal(updSecure)
	reqU1, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(usb))
	reqU1.Header.Set("Authorization", "Bearer "+token)
	reqU1.Header.Set("Content-Type", "application/json")
	resU1, _ := http.DefaultClient.Do(reqU1)
	if resU1.StatusCode != http.StatusForbidden {
		t.Fatalf("expected secure_embed_enabled update 403, got %d", resU1.StatusCode)
	}

	var errResp1 map[string]any
	json.NewDecoder(resU1.Body).Decode(&errResp1)
	resU1.Body.Close()

	if feature, ok := errResp1["feature"].(string); !ok || feature != "secure_embed" {
		t.Fatalf("expected feature secure_embed, got %v", errResp1)
	}

	updDomains := map[string]any{"allowed_domains": "example.com"}
	udb, _ := json.Marshal(updDomains)
	reqU2, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(udb))
	reqU2.Header.Set("Authorization", "Bearer "+token)
	reqU2.Header.Set("Content-Type", "application/json")
	resU2, _ := http.DefaultClient.Do(reqU2)
	if resU2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected allowed_domains update 403, got %d", resU2.StatusCode)
	}

	var errResp2 map[string]any
	json.NewDecoder(resU2.Body).Decode(&errResp2)
	resU2.Body.Close()

	if feature, ok := errResp2["feature"].(string); !ok || feature != "secure_embed" {
		t.Fatalf("expected feature secure_embed for allowed_domains, got %v", errResp2)
	}
}

func TestFreePlan_Guardrails_Restrictions(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(
		COALESCE(config, '{}'::jsonb),
		'{guardrails}',
		'{
			"can_manage_topics": false,
			"can_customize_messages": false,
			"can_customize_thresholds": false,
			"can_use_smart_fallback": false,
			"can_use_escalate_fallback": false
		}'::jsonb,
		true
	) WHERE code = '` + policy.PlanFree.String() + `'`)

	token := authToken(t, te.Server.URL, "guardrails_free@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, "guardrails_free@example.com")

	create := map[string]any{"name": "Guardrails Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	updThresholds := map[string]any{
		"threshold_config": map[string]any{
			"high_threshold":   0.9,
			"medium_threshold": 0.4,
		},
	}
	tb, _ := json.Marshal(updThresholds)
	reqT, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(tb))
	reqT.Header.Set("Authorization", "Bearer "+token)
	reqT.Header.Set("Content-Type", "application/json")
	resT, _ := http.DefaultClient.Do(reqT)
	if resT.StatusCode != http.StatusForbidden {
		t.Fatalf("expected threshold update 403, got %d", resT.StatusCode)
	}
	resT.Body.Close()

	updSmart := map[string]any{
		"threshold_config": map[string]any{
			"fallback_mode": "smart",
		},
	}
	sb, _ := json.Marshal(updSmart)
	reqS, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(sb))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", "application/json")
	resS, _ := http.DefaultClient.Do(reqS)
	if resS.StatusCode != http.StatusForbidden {
		t.Fatalf("expected smart fallback update 403, got %d", resS.StatusCode)
	}
	resS.Body.Close()

	updTopics := map[string]any{
		"topic_restrictions": map[string]any{
			"allowed_topics": []string{"tech"},
		},
	}
	tpb, _ := json.Marshal(updTopics)
	reqTR, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(tpb))
	reqTR.Header.Set("Authorization", "Bearer "+token)
	reqTR.Header.Set("Content-Type", "application/json")
	resTR, _ := http.DefaultClient.Do(reqTR)
	if resTR.StatusCode != http.StatusForbidden {
		t.Fatalf("expected topic restrictions update 403, got %d", resTR.StatusCode)
	}
	resTR.Body.Close()
}

func TestFreePlan_PDF_OCRDisabled_NoTextExtracted(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENAI_API_KEY", "sk-test")
	t.Setenv("QDRANT_URL", qd.URL)

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	_, _ = te.DB.Exec(`UPDATE plans SET config = '{
  "scraping": {
    "dynamic_enabled": false,
    "max_urls_per_bot": 1,
    "max_pages_per_crawl": 0
  },
  "files": {
    "ocr_enabled": false,
    "max_size_mb": 5,
    "max_files_per_bot": 1,
    "total_storage_mb": 10
  },
  "chat": {
    "allowed_models": ["` + policy.ModelGPT4oMini.String() + `"],
    "max_monthly_tokens": 100000,
    "rag": {
      "top_k": 3,
      "max_context_tokens": 2000
    }
  }
}'::jsonb WHERE code = '` + policy.PlanFree.String() + `'`)

	email := "ocr_free@example.com"
	token := authToken(t, te.Server.URL, email)
	_, _ = te.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = '`+policy.PlanFree.String()+`') WHERE email = $1`, email)

	create := map[string]any{"name": "OCR Disabled Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated && resC.StatusCode != http.StatusOK {
		t.Fatalf("chatbot create failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	var body strings.Builder
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "image-only.pdf")
	fw.Write([]byte("%PDF-1.4\nstub"))
	mw.WriteField("source_type", "pdf")
	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(body.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := http.DefaultClient.Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected pdf source 201, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	sourceID := sid["id"]

	waitForProcessing(t, te, token, sourceID)

	src, err := db.GetSourceByID(context.Background(), te.DB, sourceID)
	if err != nil {
		t.Fatalf("failed to load source: %v", err)
	}
	if src == nil {
		t.Fatalf("expected source to exist")
	}
	if src.ChunkCount != 0 {
		t.Fatalf("expected zero chunks for OCR-disabled free plan, got %d", src.ChunkCount)
	}
}
