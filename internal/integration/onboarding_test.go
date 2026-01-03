package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

// authToken helper for onboarding tests
func authTokenOnboarding(t *testing.T, base string, email string) string {
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	testHTTPPost(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": "Test@123"}
	lbj, _ := json.Marshal(lb)
	res, _ := testHTTPPost(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	var tr tokenResp
	json.NewDecoder(res.Body).Decode(&tr)
	drainBody(res)
	return tr.Token
}

// TestOnboardingFlow tests the complete onboarding flow
func TestOnboardingFlow(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// 1. Register a new user
	email := fmt.Sprintf("onboarding-flow+%d@example.com", time.Now().UnixNano())
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "Onboarding User"}
	regBytes, _ := json.Marshal(regBody)
	regRes, err := testHTTPPost(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBytes))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if regRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", regRes.StatusCode)
	}
	regRes.Body.Close()

	// 2. Login to get token
	token := authTokenOnboarding(t, te.Server.URL, email)

	// 3. Get initial onboarding state (should be step 0, not completed)
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/onboarding", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("get onboarding state failed: %d", res.StatusCode)
	}
	var state map[string]interface{}
	json.NewDecoder(res.Body).Decode(&state)
	drainBody(res)

	if state["completed"] != false {
		t.Errorf("expected completed=false, got %v", state["completed"])
	}
	if state["skipped"] != false {
		t.Errorf("expected skipped=false, got %v", state["skipped"])
	}
	if state["step"] != float64(0) {
		t.Errorf("expected step=0, got %v", state["step"])
	}

	// 4. Update onboarding state to step 1
	updateBody := map[string]interface{}{
		"step": 1,
		"data": map[string]interface{}{
			"bot_name": "My Test Bot",
		},
	}
	updateBytes, _ := json.Marshal(updateBody)
	reqUpdate, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/me/onboarding", bytes.NewReader(updateBytes))
	reqUpdate.Header.Set("Authorization", "Bearer "+token)
	reqUpdate.Header.Set("Content-Type", "application/json")
	resUpdate, _ := testHTTPClient().Do(reqUpdate)
	if resUpdate.StatusCode != http.StatusOK {
		t.Fatalf("update onboarding failed: %d", resUpdate.StatusCode)
	}
	resUpdate.Body.Close()

	// 5. Verify state was updated
	req2, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/onboarding", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	res2, _ := testHTTPClient().Do(req2)
	var state2 map[string]interface{}
	json.NewDecoder(res2.Body).Decode(&state2)
	res2.Body.Close()

	if state2["step"] != float64(1) {
		t.Errorf("expected step=1, got %v", state2["step"])
	}
	data := state2["data"].(map[string]interface{})
	if data["bot_name"] != "My Test Bot" {
		t.Errorf("expected bot_name='My Test Bot', got %v", data["bot_name"])
	}

	// 6. Create a bot
	botBody := map[string]string{"name": "My Test Bot"}
	botBytes, _ := json.Marshal(botBody)
	reqBot, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(botBytes))
	reqBot.Header.Set("Authorization", "Bearer "+token)
	reqBot.Header.Set("Content-Type", "application/json")
	resBot, _ := testHTTPClient().Do(reqBot)
	if resBot.StatusCode != http.StatusCreated {
		t.Fatalf("create bot failed: %d", resBot.StatusCode)
	}
	var bot map[string]interface{}
	json.NewDecoder(resBot.Body).Decode(&bot)
	resBot.Body.Close()
	botID := bot["id"].(string)

	// 7. Complete onboarding
	completeBody := map[string]string{"bot_id": botID}
	completeBytes, _ := json.Marshal(completeBody)
	reqComplete, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/onboarding/complete", bytes.NewReader(completeBytes))
	reqComplete.Header.Set("Authorization", "Bearer "+token)
	reqComplete.Header.Set("Content-Type", "application/json")
	resComplete, _ := testHTTPClient().Do(reqComplete)
	if resComplete.StatusCode != http.StatusOK {
		t.Fatalf("complete onboarding failed: %d", resComplete.StatusCode)
	}
	resComplete.Body.Close()

	// 8. Verify onboarding is completed
	req3, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/onboarding", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	res3, _ := testHTTPClient().Do(req3)
	var state3 map[string]interface{}
	json.NewDecoder(res3.Body).Decode(&state3)
	res3.Body.Close()

	if state3["completed"] != true {
		t.Errorf("expected completed=true, got %v", state3["completed"])
	}
	if state3["step"] != float64(4) {
		t.Errorf("expected step=4, got %v", state3["step"])
	}
}

// TestOnboardingSkip tests skipping the onboarding
func TestOnboardingSkip(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Register and login
	email := fmt.Sprintf("onboarding-skip+%d@example.com", time.Now().UnixNano())
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "Skip User"}
	regBytes, _ := json.Marshal(regBody)
	regRes, _ := testHTTPPost(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBytes))
	regRes.Body.Close()

	token := authTokenOnboarding(t, te.Server.URL, email)

	// Skip onboarding
	reqSkip, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/onboarding/skip", nil)
	reqSkip.Header.Set("Authorization", "Bearer "+token)
	resSkip, _ := testHTTPClient().Do(reqSkip)
	if resSkip.StatusCode != http.StatusOK {
		t.Fatalf("skip onboarding failed: %d", resSkip.StatusCode)
	}
	resSkip.Body.Close()

	// Verify skipped
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/onboarding", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := testHTTPClient().Do(req)
	var state map[string]interface{}
	json.NewDecoder(res.Body).Decode(&state)
	drainBody(res)

	if state["skipped"] != true {
		t.Errorf("expected skipped=true, got %v", state["skipped"])
	}
}

// TestOnboardingStatePersistedAcrossSessions tests that onboarding state persists
func TestOnboardingStatePersistedAcrossSessions(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Register and login
	email := fmt.Sprintf("onboarding-persist+%d@example.com", time.Now().UnixNano())
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "Persist User"}
	regBytes, _ := json.Marshal(regBody)
	testHTTPPost(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBytes))

	token1 := authTokenOnboarding(t, te.Server.URL, email)

	// Set onboarding data
	updateBody := map[string]interface{}{
		"step": 2,
		"data": map[string]interface{}{
			"bot_name":      "Persistent Bot",
			"source_type":   "url",
			"url_content":   "https://example.com",
			"system_prompt": "Custom prompt",
		},
	}
	updateBytes, _ := json.Marshal(updateBody)
	reqUpdate, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/me/onboarding", bytes.NewReader(updateBytes))
	reqUpdate.Header.Set("Authorization", "Bearer "+token1)
	reqUpdate.Header.Set("Content-Type", "application/json")
	resUpdate, _ := testHTTPClient().Do(reqUpdate)
	resUpdate.Body.Close()

	// Simulate new session - login again with new token
	token2 := authTokenOnboarding(t, te.Server.URL, email)

	// Verify data persists
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/onboarding", nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	res, _ := testHTTPClient().Do(req)
	var state map[string]interface{}
	json.NewDecoder(res.Body).Decode(&state)
	drainBody(res)

	if state["step"] != float64(2) {
		t.Errorf("expected step=2, got %v", state["step"])
	}

	data := state["data"].(map[string]interface{})
	if data["bot_name"] != "Persistent Bot" {
		t.Errorf("expected bot_name='Persistent Bot', got %v", data["bot_name"])
	}
	if data["url_content"] != "https://example.com" {
		t.Errorf("expected url_content='https://example.com', got %v", data["url_content"])
	}
	if data["system_prompt"] != "Custom prompt" {
		t.Errorf("expected system_prompt='Custom prompt', got %v", data["system_prompt"])
	}
}
