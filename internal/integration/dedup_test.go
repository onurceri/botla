package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"
)

func TestPDFDeduplication_Integration(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := fmt.Sprintf("dedup-%d@example.com", time.Now().UnixNano())
	token := authTokenDedup(t, te.Server.URL, email)
	chatbotID := createChatbotForDedup(t, te.Server.URL, token, "Dedup Bot")

	// Relax PDF limit for this test to avoid ERR_PDF_LIMIT_REACHED
	_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{files,max_files_per_bot}', '10'::jsonb) WHERE code = 'free'`)

	// Create a simple PDF
	pdfContent := []byte("%PDF-1.4 dummy content for test")

	// Upload first time - should succeed
	resp1 := uploadPDF(t, te.Server.URL, token, chatbotID, pdfContent, "test.pdf")
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first upload failed: %d", resp1.StatusCode)
	}

	// Upload same content again - should fail with 409
	resp2 := uploadPDF(t, te.Server.URL, token, chatbotID, pdfContent, "test2.pdf")
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("second upload should be 409 Conflict, got %d", resp2.StatusCode)
	}

	// Verify error code
	body, _ := io.ReadAll(resp2.Body)
	if !bytes.Contains(body, []byte("ERR_DUPLICATE_CONTENT")) {
		t.Errorf("expected ERR_DUPLICATE_CONTENT error, got %s", string(body))
	}
}

func TestTextDeduplication_Integration(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := fmt.Sprintf("textdedup-%d@example.com", time.Now().UnixNano())
	token := authTokenDedup(t, te.Server.URL, email)
	chatbotID := createChatbotForDedup(t, te.Server.URL, token, "Text Dedup Bot")

	textContent := "This is some unique text content for testing"

	// Upload first time
	resp1 := uploadText(t, te.Server.URL, token, chatbotID, textContent)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first upload failed: %d", resp1.StatusCode)
	}

	// Upload same text again
	resp2 := uploadText(t, te.Server.URL, token, chatbotID, textContent)
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("second upload should be 409, got %d", resp2.StatusCode)
	}
}

func TestDeduplication_DifferentChatbots_Allowed(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := fmt.Sprintf("crossbot-%d@example.com", time.Now().UnixNano())
	token := authTokenDedup(t, te.Server.URL, email)
	chatbot1 := createChatbotForDedup(t, te.Server.URL, token, "Bot 1")
	chatbot2 := createChatbotForDedup(t, te.Server.URL, token, "Bot 2")

	textContent := "Same content for different bots"

	// Upload to first chatbot
	resp1 := uploadText(t, te.Server.URL, token, chatbot1, textContent)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("upload to bot1 failed: %d", resp1.StatusCode)
	}

	// Upload same content to second chatbot - should SUCCEED
	resp2 := uploadText(t, te.Server.URL, token, chatbot2, textContent)
	if resp2.StatusCode != http.StatusCreated {
		t.Errorf("upload to bot2 should succeed, got %d", resp2.StatusCode)
	}
}

func createChatbotForDedup(t *testing.T, baseURL, token, name string) string {
	create := map[string]any{"name": name, "language": "en-US"}
	cb, _ := json.Marshal(create)
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/chatbots", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create chatbot: %d", res.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	json.NewDecoder(res.Body).Decode(&created)
	return created.ID
}

func uploadPDF(t *testing.T, baseURL, token, chatbotID string, content []byte, filename string) *http.Response {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(content)
	_ = writer.WriteField("source_type", "pdf")
	_ = writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/sources", &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func uploadText(t *testing.T, baseURL, token, chatbotID, text string) *http.Response {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("source_type", "text")
	_ = writer.WriteField("text", text)
	_ = writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/sources", &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func authTokenDedup(t *testing.T, base string, email string) string {
	t.Helper()
	// Register
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	resp, err := http.Post(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("auth register request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("auth register failed: %d %s", resp.StatusCode, string(body))
	}

	// Login
	lb := map[string]string{"email": email, "password": "Test@123"}
	lbj, _ := json.Marshal(lb)
	resp, err = http.Post(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	if err != nil {
		t.Fatalf("auth login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("auth login failed: %d %s", resp.StatusCode, string(body))
	}

	var tr struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		t.Fatalf("decode token failed: %v", err)
	}
	if tr.Token == "" {
		t.Fatal("empty token received")
	}
	return tr.Token
}
