package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

// Multi-tenant isolation tests verify that users cannot access resources
// belonging to other users, ensuring data security in a SaaS environment.

// Multi-tenant auth helper - creates unique users per test to avoid collisions
func mtAuthToken(t *testing.T, base string, email string) string {
	t.Helper()
	regBody := map[string]string{"email": email, "password": fixtures.TestPassword, "full_name": "User"}
	b, _ := json.Marshal(regBody)
	_, _ = testHTTPPost(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": fixtures.TestPassword}
	lbj, _ := json.Marshal(lb)
	res, err := testHTTPPost(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer drainBody(res)
	var tr tokenResp
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	return tr.Token
}

// mtCreateChatbot creates a chatbot and returns its ID
func mtCreateChatbot(t *testing.T, baseURL, token, name string) string {
	t.Helper()

	payload := map[string]string{"name": name}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create chatbot: %v", err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create chatbot, status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode chatbot response: %v", err)
	}

	id, ok := result["id"].(string)
	if !ok {
		t.Fatal("chatbot id not found in response")
	}

	return id
}

// mtCreateTextSource creates a text source and returns its ID
func mtCreateTextSource(t *testing.T, baseURL, token, chatbotID, content string) string {
	t.Helper()

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	_ = mw.WriteField("source_type", "text")
	_ = mw.WriteField("text", content)
	_ = mw.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/sources", &b)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create source, status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode source response: %v", err)
	}

	id, ok := result["id"].(string)
	if !ok {
		t.Fatal("source id not found in response")
	}

	return id
}

// mtListChatbots lists all chatbots for a user
func mtListChatbots(t *testing.T, baseURL, token string) []map[string]interface{} {
	t.Helper()

	req, _ := http.NewRequest("GET", baseURL+"/api/v1/chatbots", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to list chatbots: %v", err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to list chatbots, status: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode chatbots response: %v", err)
	}

	return result
}

func TestMultiTenant_ChatbotIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot
	tokenA := mtAuthToken(t, te.Server.URL, "mt_cb_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "User A Bot")

	// User B tries to access User A's chatbot
	tokenB := mtAuthToken(t, te.Server.URL, "mt_cb_userB@test.com")

	t.Run("User B cannot GET User A's chatbot", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botA, nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		// Should be 403 Forbidden or 404 Not Found
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 403/404, got %d", resp.StatusCode)
		}
	})

	t.Run("User B cannot UPDATE User A's chatbot", func(t *testing.T) {
		update := map[string]any{"name": "Hacked Bot"}
		uj, _ := json.Marshal(update)
		req, _ := http.NewRequest("PUT", te.Server.URL+"/api/v1/chatbots/"+botA, bytes.NewReader(uj))
		req.Header.Set("Authorization", "Bearer "+tokenB)
		req.Header.Set("Content-Type", "application/json")

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 403/404, got %d", resp.StatusCode)
		}
	})

	t.Run("User B cannot DELETE User A's chatbot", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", te.Server.URL+"/api/v1/chatbots/"+botA, nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 403/404, got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_SourceIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot and source
	tokenA := mtAuthToken(t, te.Server.URL, "mt_src_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "Bot A")
	sourceA := mtCreateTextSource(t, te.Server.URL, tokenA, botA, "Secret content from User A")

	// User B tries to access
	tokenB := mtAuthToken(t, te.Server.URL, "mt_src_userB@test.com")

	t.Run("User B cannot access User A's source via source endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceA, nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("source access leak: got %d", resp.StatusCode)
		}
	})

	t.Run("User B cannot delete User A's source", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", te.Server.URL+"/api/v1/sources/"+sourceA, nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("could delete other user's source: got %d", resp.StatusCode)
		}
	})

	t.Run("User B cannot add source to User A's chatbot", func(t *testing.T) {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		_ = mw.WriteField("source_type", "text")
		_ = mw.WriteField("text", "Malicious content")
		_ = mw.Close()

		req, _ := http.NewRequest("POST", te.Server.URL+"/api/v1/chatbots/"+botA+"/sources", &b)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		req.Header.Set("Content-Type", mw.FormDataContentType())

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("could add source to other user's chatbot: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_JobIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot and source (which creates a job)
	tokenA := mtAuthToken(t, te.Server.URL, "mt_job_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "Job Bot")
	sourceA := mtCreateTextSource(t, te.Server.URL, tokenA, botA, "Content for job test")

	// User B tries to access job
	tokenB := mtAuthToken(t, te.Server.URL, "mt_job_userB@test.com")

	t.Run("User B cannot access User A's job status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceA+"/job", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("job status leak: got %d", resp.StatusCode)
		}
	})

	t.Run("User B cannot retry User A's job", func(t *testing.T) {
		req, _ := http.NewRequest("POST", te.Server.URL+"/api/v1/sources/"+sourceA+"/job/retry", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("could retry other user's job: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_ListingIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates resources
	tokenA := mtAuthToken(t, te.Server.URL, "mt_list_userA@test.com")
	_ = mtCreateChatbot(t, te.Server.URL, tokenA, "A's Bot 1")
	_ = mtCreateChatbot(t, te.Server.URL, tokenA, "A's Bot 2")

	// User B creates resources
	tokenB := mtAuthToken(t, te.Server.URL, "mt_list_userB@test.com")
	_ = mtCreateChatbot(t, te.Server.URL, tokenB, "B's Bot")

	t.Run("User B lists only their own chatbots", func(t *testing.T) {
		bots := mtListChatbots(t, te.Server.URL, tokenB)

		// Should only see their own
		for _, bot := range bots {
			if name, ok := bot["name"].(string); ok {
				if strings.Contains(name, "A's Bot") {
					t.Error("User B can see User A's chatbots")
				}
			}
		}

		// Should have exactly 1 bot
		if len(bots) != 1 {
			t.Errorf("expected User B to see 1 chatbot, got %d", len(bots))
		}
	})

	t.Run("User A lists only their own chatbots", func(t *testing.T) {
		bots := mtListChatbots(t, te.Server.URL, tokenA)

		// Should only see their own
		for _, bot := range bots {
			if name, ok := bot["name"].(string); ok {
				if strings.Contains(name, "B's Bot") {
					t.Error("User A can see User B's chatbots")
				}
			}
		}

		// Should have exactly 2 bots
		if len(bots) != 2 {
			t.Errorf("expected User A to see 2 chatbots, got %d", len(bots))
		}
	})
}

func TestMultiTenant_SourceListingIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot and sources
	tokenA := mtAuthToken(t, te.Server.URL, "mt_srclist_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "A's Bot for Sources")
	_ = mtCreateTextSource(t, te.Server.URL, tokenA, botA, "A's secret content 1")
	_ = mtCreateTextSource(t, te.Server.URL, tokenA, botA, "A's secret content 2")

	// User B creates chatbot and source
	tokenB := mtAuthToken(t, te.Server.URL, "mt_srclist_userB@test.com")
	botB := mtCreateChatbot(t, te.Server.URL, tokenB, "B's Bot for Sources")
	_ = mtCreateTextSource(t, te.Server.URL, tokenB, botB, "B's content")

	t.Run("User B cannot list sources of User A's chatbot", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botA+"/sources", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("source listing leak: got %d", resp.StatusCode)
		}
	})

	t.Run("User A cannot list sources of User B's chatbot", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botB+"/sources", nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("source listing leak: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_AnalyticsIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot
	tokenA := mtAuthToken(t, te.Server.URL, "mt_analytics_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "A's Analytics Bot")

	// User B tries to access A's analytics
	tokenB := mtAuthToken(t, te.Server.URL, "mt_analytics_userB@test.com")

	t.Run("User B cannot access User A's chatbot analytics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botA+"/analytics", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("analytics leak: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_ActionIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot
	tokenA := mtAuthToken(t, te.Server.URL, "mt_action_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "A's Action Bot")

	// User B tries to access A's actions
	tokenB := mtAuthToken(t, te.Server.URL, "mt_action_userB@test.com")

	t.Run("User B cannot list User A's chatbot actions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botA+"/actions", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("actions leak: got %d", resp.StatusCode)
		}
	})

	t.Run("User B cannot create action on User A's chatbot", func(t *testing.T) {
		action := map[string]any{
			"name":        "malicious_action",
			"description": "Trying to add action to other user's bot",
			"parameters":  map[string]any{},
		}
		actionJ, _ := json.Marshal(action)
		req, _ := http.NewRequest("POST", te.Server.URL+"/api/v1/chatbots/"+botA+"/actions", bytes.NewReader(actionJ))
		req.Header.Set("Authorization", "Bearer "+tokenB)
		req.Header.Set("Content-Type", "application/json")

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("could create action on other user's bot: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_HandoffIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot
	tokenA := mtAuthToken(t, te.Server.URL, "mt_handoff_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "A's Handoff Bot")

	// User B tries to access A's handoff requests
	tokenB := mtAuthToken(t, te.Server.URL, "mt_handoff_userB@test.com")

	t.Run("User B cannot list User A's chatbot handoff requests", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botA+"/handoff-requests", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("handoff requests leak: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_ChatIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnvWithMocks()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot
	tokenA := mtAuthToken(t, te.Server.URL, "mt_chat_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "A's Chat Bot")

	// User B tries to chat on A's chatbot
	tokenB := mtAuthToken(t, te.Server.URL, "mt_chat_userB@test.com")

	t.Run("User B cannot chat on User A's chatbot", func(t *testing.T) {
		chatReq := map[string]any{
			"message":    "Hello",
			"session_id": "test-session",
		}
		chatJ, _ := json.Marshal(chatReq)
		req, _ := http.NewRequest("POST", te.Server.URL+"/api/v1/chatbots/"+botA+"/chat", bytes.NewReader(chatJ))
		req.Header.Set("Authorization", "Bearer "+tokenB)
		req.Header.Set("Content-Type", "application/json")

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("could chat on other user's bot: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_SuggestionsIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot
	tokenA := mtAuthToken(t, te.Server.URL, "mt_sugg_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "A's Suggestions Bot")

	// User B tries to access A's suggestions
	tokenB := mtAuthToken(t, te.Server.URL, "mt_sugg_userB@test.com")

	t.Run("User B cannot access User A's chatbot suggestions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botA+"/suggestions", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("suggestions leak: got %d", resp.StatusCode)
		}
	})

	t.Run("User B cannot update User A's chatbot suggestions", func(t *testing.T) {
		suggestions := map[string]any{
			"suggestions": []string{"malicious question 1", "malicious question 2"},
		}
		suggJ, _ := json.Marshal(suggestions)
		req, _ := http.NewRequest("PUT", te.Server.URL+"/api/v1/chatbots/"+botA+"/suggestions", bytes.NewReader(suggJ))
		req.Header.Set("Authorization", "Bearer "+tokenB)
		req.Header.Set("Content-Type", "application/json")

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("could update other user's suggestions: got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenant_PendingURLsIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// User A creates chatbot
	tokenA := mtAuthToken(t, te.Server.URL, "mt_urls_userA@test.com")
	botA := mtCreateChatbot(t, te.Server.URL, tokenA, "A's URL Bot")

	// User B tries to access A's pending URLs
	tokenB := mtAuthToken(t, te.Server.URL, "mt_urls_userB@test.com")

	t.Run("User B cannot access User A's chatbot pending URLs", func(t *testing.T) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots/"+botA+"/pending-urls", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer drainBody(resp)

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			t.Errorf("pending URLs leak: got %d", resp.StatusCode)
		}
	})
}
