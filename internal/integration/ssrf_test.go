package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

type chatbotInfo struct {
	ID string `json:"id"`
}

func ssrfAuthToken(t *testing.T, base string, email string) string {
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	testHTTPPost(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": "Test@123"}
	lbj, _ := json.Marshal(lb)
	res, _ := testHTTPPost(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	var tr struct {
		Token string `json:"token"`
	}
	json.NewDecoder(res.Body).Decode(&tr)
	drainBody(res)
	return tr.Token
}

func ssrfCreateChatbot(t *testing.T, base string, token string, name string) string {
	cbBody := map[string]string{"name": name}
	b, _ := json.Marshal(cbBody)
	req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/chatbots", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := testHTTPClient().Do(req)
	if err != nil {
		t.Fatalf("create chatbot request failed: %v", err)
	}
	defer drainBody(res)
	var bot chatbotInfo
	json.NewDecoder(res.Body).Decode(&bot)
	return bot.ID
}

func createURLSource(t *testing.T, base string, token string, chatbotID string, sourceURL string) *http.Response {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("source_type", "url")
	_ = mw.WriteField("source_url", sourceURL)
	_ = mw.Close()

	req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/chatbots/"+chatbotID+"/sources", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	res, err := testHTTPClient().Do(req)
	if err != nil {
		t.Fatalf("create URL source request failed: %v", err)
	}
	return res
}

func TestSSRFProtection_Integration(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := ssrfAuthToken(t, te.Server.URL, "ssrf@example.com")
	chatbotID := ssrfCreateChatbot(t, te.Server.URL, token, "SSRF Test Bot")

	// Enforce strict SSRF protection for this test
	te.SourcesHandlers.SSRFValidator.SetAllowPrivate(false)

	blockedURLs := []string{
		"http://localhost/admin",
		"http://127.0.0.1:8080/internal",
		"http://192.168.1.1/router",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
	}

	for _, url := range blockedURLs {
		resp := createURLSource(t, te.Server.URL, token, chatbotID, url)

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("URL %s should be blocked with 403, got %d", url, resp.StatusCode)
		}

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		drainBody(resp)

		if body["code"] != "ERR_BLOCKED_URL" {
			t.Errorf("expected ERR_BLOCKED_URL code, got %s", body["code"])
		}
	}
}

func TestSSRFProtection_AllowsPublicURLs(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := ssrfAuthToken(t, te.Server.URL, "ssrfpublic@example.com")
	chatbotID := ssrfCreateChatbot(t, te.Server.URL, token, "Public URL Test")

	// This should succeed (though it might fail to scrape, it should pass SSRF check)
	resp := createURLSource(t, te.Server.URL, token, chatbotID, "https://example.com")
	defer drainBody(resp)

	// Should be 201 Created (might process successfully) or not 403 (SSRF blocked)
	if resp.StatusCode == http.StatusForbidden {
		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		if body["code"] == "ERR_BLOCKED_URL" {
			t.Error("public URL should not be SSRF blocked")
		}
	}
}
