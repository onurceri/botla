package integration

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
)

func TestSourceRefresh_Success(t *testing.T) {
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

	// Apply migration for refresh columns
	_, _ = te.DB.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_refreshed_at TIMESTAMPTZ`)
	_, _ = te.DB.Exec(`ALTER TABLE usage_ingestions ADD COLUMN IF NOT EXISTS refresh_count INT DEFAULT 0`)

	// Create pro user with refresh enabled
	token := authToken(t, te.Server.URL, "refresh_pro@example.com")
	_ = te.UpdatePlanLimit("free", "refresh_enabled", true)
	_ = te.UpdatePlanLimit("free", "refresh_max_monthly", 5)

	// Create chatbot
	create := map[string]any{"name": "Refresh Bot"}
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
	completed := false
	for i := 0; i < 100; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+statusPath, nil)
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
		if st["status"].(string) == "completed" {
			completed = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !completed {
		t.Fatalf("source never completed")
	}

	// Refresh the source
	refreshPath := "/api/v1/sources/" + url.PathEscape(sourceID) + "/refresh"
	reqR, _ := http.NewRequest(http.MethodPost, te.Server.URL+refreshPath, nil)
	reqR.Header.Set("Authorization", "Bearer "+token)
	resR, _ := testHTTPClient().Do(reqR)
	if resR.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resR.StatusCode)
	}
	resR.Body.Close()

	// Wait for refresh to complete
	refreshCompleted := false
	for i := 0; i < 100; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+statusPath, nil)
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
		if st["status"].(string) == "completed" {
			refreshCompleted = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !refreshCompleted {
		t.Fatalf("refresh never completed")
	}
}

func TestSourceRefresh_FreePlanBlocked(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Apply migration for refresh columns
	_, _ = te.DB.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_refreshed_at TIMESTAMPTZ`)
	_, _ = te.DB.Exec(`ALTER TABLE usage_ingestions ADD COLUMN IF NOT EXISTS refresh_count INT DEFAULT 0`)

	// Ensure free plan has refresh disabled
	_ = te.UpdatePlanLimit("free", "refresh_enabled", false)
	_ = te.UpdatePlanLimit("free", "refresh_max_monthly", 0)

	token := authToken(t, te.Server.URL, "refresh_free@example.com")

	// Create chatbot
	create := map[string]any{"name": "Free Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add a URL source directly to DB (skip processing)
	var sourceID string
	url := "https://example.com"
	err = te.DB.QueryRow(`INSERT INTO data_sources (chatbot_id, source_type, source_url, status) VALUES ($1, 'url', $2, 'completed') RETURNING id`, bot.ID, url).Scan(&sourceID)
	if err != nil {
		t.Fatalf("insert source: %v", err)
	}

	// Try to refresh - should be forbidden
	refreshPath := "/api/v1/sources/" + sourceID + "/refresh"
	reqR, _ := http.NewRequest(http.MethodPost, te.Server.URL+refreshPath, nil)
	reqR.Header.Set("Authorization", "Bearer "+token)
	resR, _ := testHTTPClient().Do(reqR)
	if resR.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resR.StatusCode)
	}
	resR.Body.Close()
}

func TestSourceRefresh_NonURLBlocked(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Apply migration for refresh columns
	_, _ = te.DB.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_refreshed_at TIMESTAMPTZ`)
	_, _ = te.DB.Exec(`ALTER TABLE usage_ingestions ADD COLUMN IF NOT EXISTS refresh_count INT DEFAULT 0`)
	_ = te.UpdatePlanLimit("free", "refresh_enabled", true)
	_ = te.UpdatePlanLimit("free", "refresh_max_monthly", 5)

	token := authToken(t, te.Server.URL, "refresh_pdf@example.com")

	// Create chatbot
	create := map[string]any{"name": "PDF Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add a PDF source directly to DB
	var sourceID string
	err = te.DB.QueryRow(`INSERT INTO data_sources (chatbot_id, source_type, original_filename, status) VALUES ($1, 'pdf', 'test.pdf', 'completed') RETURNING id`, bot.ID).Scan(&sourceID)
	if err != nil {
		t.Fatalf("insert source: %v", err)
	}

	// Try to refresh PDF - should be bad request
	refreshPath := "/api/v1/sources/" + sourceID + "/refresh"
	reqR, _ := http.NewRequest(http.MethodPost, te.Server.URL+refreshPath, nil)
	reqR.Header.Set("Authorization", "Bearer "+token)
	resR, _ := testHTTPClient().Do(reqR)
	if resR.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resR.StatusCode)
	}
	resR.Body.Close()
}

func TestSourceRefresh_QuotaExceeded(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Apply migration for refresh columns
	_, _ = te.DB.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_refreshed_at TIMESTAMPTZ`)
	_, _ = te.DB.Exec(`ALTER TABLE usage_ingestions ADD COLUMN IF NOT EXISTS refresh_count INT DEFAULT 0`)
	_ = te.UpdatePlanLimit("free", "refresh_enabled", true)
	_ = te.UpdatePlanLimit("free", "refresh_max_monthly", 1)

	token := authToken(t, te.Server.URL, "refresh_quota@example.com")

	// Create chatbot
	create := map[string]any{"name": "Quota Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Get user ID
	var userID string
	te.DB.QueryRow(`SELECT user_id FROM chatbots WHERE id=$1`, bot.ID).Scan(&userID)

	// Add a URL source
	var sourceID string
	url := "https://example.com"
	err = te.DB.QueryRow(`INSERT INTO data_sources (chatbot_id, source_type, source_url, status) VALUES ($1, 'url', $2, 'completed') RETURNING id`, bot.ID, url).Scan(&sourceID)
	if err != nil {
		t.Fatalf("insert source: %v", err)
	}

	// Pre-fill quota
	te.DB.Exec(`INSERT INTO usage_ingestions (user_id, period_month, refresh_count) VALUES ($1, date_trunc('month', NOW())::date, 1)`, userID)

	// Try to refresh - should be payment required
	refreshPath := "/api/v1/sources/" + sourceID + "/refresh"
	reqR, _ := http.NewRequest(http.MethodPost, te.Server.URL+refreshPath, nil)
	reqR.Header.Set("Authorization", "Bearer "+token)
	resR, _ := testHTTPClient().Do(reqR)
	if resR.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", resR.StatusCode)
	}
	resR.Body.Close()
}
