package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/policy"
)

// SRC-006: Upload exceeding size limit
func TestSources_SizeLimitExceeded(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Set max size to 1MB for free plan (update plan_limits table)
	_, _ = te.DB.Exec(`UPDATE plan_limits SET files_max_size_mb = 1 WHERE plan_id = (SELECT id FROM plans WHERE code = $1)`, policy.PlanFree.String())

	token := authToken(t, te.Server.URL, "sizelimit@example.com")

	// Create chatbot
	create := map[string]any{"name": "Size Limit Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Create 2MB content
	largeContent := make([]byte, 2*1024*1024)
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	// Upload large file
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "pdf")

	part, _ := mw.CreateFormFile("file", "large.pdf")
	part.Write(largeContent)

	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", &b)
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := testHTTPClient().Do(reqS)

	// Expect 413 Payload Too Large
	if resS.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resS.StatusCode)
	}
	resS.Body.Close()
}

func TestSources_PDFLimitPerBot(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	_, _ = te.DB.Exec(`UPDATE plans SET config = config || '{"files": {"max_size_mb": 10, "max_files_per_bot": 1}}'::jsonb WHERE code = $1`, policy.PlanFree.String())

	email := "pdflimit@example.com"
	token := authToken(t, te.Server.URL, email)
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), email)

	create := map[string]any{"name": "PDF Limit Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	makeBody := func(size int) (*bytes.Buffer, string) {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		_ = mw.WriteField("source_type", "pdf")
		part, _ := mw.CreateFormFile("file", "file.pdf")
		content := make([]byte, size)
		for i := range content {
			content[i] = 'a'
		}
		_, _ = part.Write(content)
		_ = mw.Close()
		return &b, mw.FormDataContentType()
	}

	body1, ct1 := makeBody(512 * 1024)
	reqS1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", body1)
	reqS1.Header.Set("Authorization", "Bearer "+token)
	reqS1.Header.Set("Content-Type", ct1)
	resS1, _ := testHTTPClient().Do(reqS1)
	if resS1.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resS1.StatusCode)
	}
	resS1.Body.Close()

	body2, ct2 := makeBody(512 * 1024)
	reqS2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", body2)
	reqS2.Header.Set("Authorization", "Bearer "+token)
	reqS2.Header.Set("Content-Type", ct2)
	resS2, _ := testHTTPClient().Do(reqS2)
	if resS2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resS2.StatusCode)
	}
	resS2.Body.Close()
}

func TestSources_StorageLimitExceeded(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Set storage limits (update plan_limits table)
	// Also set file count limit high enough to allow multiple uploads
	_, _ = te.DB.Exec(`UPDATE plan_limits SET files_total_storage_mb = 10, files_max_size_mb = 10, files_max_files_per_bot = 100 WHERE plan_id = (SELECT id FROM plans WHERE code = $1)`, policy.PlanFree.String())

	email := "storagelimit@example.com"
	token := authToken(t, te.Server.URL, email)
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), email)

	create := map[string]any{"name": "Storage Limit Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	makeBody := func(size int) (*bytes.Buffer, string) {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		_ = mw.WriteField("source_type", "pdf")
		part, _ := mw.CreateFormFile("file", "file.pdf")
		content := make([]byte, size)
		for i := range content {
			content[i] = 'a'
		}
		_, _ = part.Write(content)
		_ = mw.Close()
		return &b, mw.FormDataContentType()
	}

	body1, ct1 := makeBody(10 * 1024 * 1024)
	reqS1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", body1)
	reqS1.Header.Set("Authorization", "Bearer "+token)
	reqS1.Header.Set("Content-Type", ct1)
	resS1, _ := testHTTPClient().Do(reqS1)
	if resS1.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resS1.StatusCode)
	}
	resS1.Body.Close()

	body2, ct2 := makeBody(1024 * 1024)
	reqS2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", body2)
	reqS2.Header.Set("Authorization", "Bearer "+token)
	reqS2.Header.Set("Content-Type", ct2)
	resS2, _ := testHTTPClient().Do(reqS2)
	if resS2.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", resS2.StatusCode)
	}
	resS2.Body.Close()
}
