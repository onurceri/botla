package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

// SRC-006: Upload exceeding size limit
func TestSources_SizeLimitExceeded(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Set max size to 1MB for free plan
	_, _ = te.DB.Exec(`UPDATE plans SET config = config || '{"files": {"max_size_mb": 1}}'::jsonb WHERE code = 'free'`)

	token := authToken(t, te.Server.URL, "sizelimit@example.com")

	// Create chatbot
	create := map[string]any{"name": "Size Limit Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
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
	resS, _ := http.DefaultClient.Do(reqS)

	// Expect 413 Payload Too Large
	if resS.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resS.StatusCode)
	}
	resS.Body.Close()
}
