package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

func TestHandoff_Flow(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// 1. Setup User & Chatbot
	token := authTokenForHandoff(t, te.Server.URL, "handoff_user@example.com")

	// Create Chatbot
	createBot := map[string]any{
		"name":     "Handoff Bot",
		"language": "en-US",
		// Handoff not enabled initially by default usually
	}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create bot failed: %d", resC.StatusCode)
	}
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 2. Enable Handoff
	updateBot := map[string]any{
		"handoff_enabled": true,
		"handoff_type":    "email",
		"handoff_config": map[string]string{
			"email_to": "support@example.com",
		},
	}
	ub, _ := json.Marshal(updateBot)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update bot failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	// 3. Start a Conversation (Public)
	// We need a session ID
	sessionID := "handoff-session-1"
	chatReq := map[string]any{
		"message":    "I need help",
		"session_id": sessionID,
	}
	cr, _ := json.Marshal(chatReq)
	reqChat, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(cr))
	reqChat.Header.Set("Authorization", "Bearer "+token) // Can be authenticated or public, let's use auth for simplicity or public endpoint if exists
	reqChat.Header.Set("Content-Type", "application/json")
	resChat, _ := http.DefaultClient.Do(reqChat)
	if resChat.StatusCode != http.StatusOK {
		t.Fatalf("chat failed: %d", resChat.StatusCode)
	}
	resChat.Body.Close()

	// 4. Request Handoff (Public)
	handoffReq := map[string]any{
		"session_id": sessionID,
		"message":    "I want to talk to a human",
	}
	hr, _ := json.Marshal(handoffReq)
	// Note: public endpoint usually doesn't need auth, but it depends on implementation.
	// The handler calls `db.GetChatbotByID` and `db.GetOrCreateConversationBySessionID`.
	// Path: /api/public/:botId/handoff
	reqH, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH.Header.Set("Content-Type", "application/json")
	resH, _ := http.DefaultClient.Do(reqH)
	if resH.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resH.Body)
		t.Fatalf("handoff request failed: %d, body: %s", resH.StatusCode, buf.String())
	}
	var handoffRes struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	json.NewDecoder(resH.Body).Decode(&handoffRes)
	resH.Body.Close()

	if handoffRes.RequestID == "" || handoffRes.Status != "pending" {
		t.Fatalf("invalid handoff response: %+v", handoffRes)
	}

	// 5. List Handoff Requests (Owner)
	reqL, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests", nil)
	reqL.Header.Set("Authorization", "Bearer "+token)
	resL, _ := http.DefaultClient.Do(reqL)
	if resL.StatusCode != http.StatusOK {
		t.Fatalf("list requests failed: %d", resL.StatusCode)
	}
	var listRes struct {
		Requests []models.HandoffRequest `json:"requests"`
	}
	json.NewDecoder(resL.Body).Decode(&listRes)
	resL.Body.Close()

	if len(listRes.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(listRes.Requests))
	}
	if listRes.Requests[0].ID != handoffRes.RequestID {
		t.Fatalf("request ID mismatch")
	}

	// 6. Update Request Status (Owner)
	updateReq := map[string]any{
		"status": "assigned",
	}
	ur, _ := json.Marshal(updateReq)
	reqUp, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests/"+handoffRes.RequestID, bytes.NewReader(ur))
	reqUp.Header.Set("Authorization", "Bearer "+token)
	reqUp.Header.Set("Content-Type", "application/json")
	resUp, _ := http.DefaultClient.Do(reqUp)
	if resUp.StatusCode != http.StatusNoContent {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resUp.Body)
		t.Fatalf("update status failed: %d, body: %s", resUp.StatusCode, buf.String())
	}
	resUp.Body.Close()

	// 7. Verify Update
	resL2, _ := http.DefaultClient.Do(reqL) // Reuse list request
	json.NewDecoder(resL2.Body).Decode(&listRes)
	resL2.Body.Close()
	if listRes.Requests[0].Status != "assigned" {
		t.Fatalf("status not updated, got %s", listRes.Requests[0].Status)
	}
}

func TestHandoff_Analytics(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_analytics@example.com")

	// Create Bot
	createBot := map[string]any{"name": "Handoff Analytics Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Enable Handoff
	updateBot := map[string]any{
		"handoff_enabled": true,
		"handoff_type":    "email",
		"handoff_config": map[string]string{
			"email_to": "support@example.com",
		},
	}
	ub, _ := json.Marshal(updateBot)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqU)

	// Trigger Handoff
	handoffReq := map[string]any{
		"session_id": "analytics-session",
		"message":    "Human please",
	}
	hr, _ := json.Marshal(handoffReq)
	reqH, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH.Header.Set("Content-Type", "application/json")
	resH, _ := http.DefaultClient.Do(reqH)
	if resH.StatusCode != http.StatusOK {
		t.Fatalf("handoff failed: %d", resH.StatusCode)
	}
	resH.Body.Close()

	// Check Analytics with retry (async update)
	found := false
	for i := 0; i < 5; i++ {
		reqA, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/analytics", nil)
		reqA.Header.Set("Authorization", "Bearer "+token)
		resA, _ := http.DefaultClient.Do(reqA)
		if resA.StatusCode != http.StatusOK {
			t.Fatalf("analytics failed: %d", resA.StatusCode)
		}

		var analytics []map[string]any
		json.NewDecoder(resA.Body).Decode(&analytics)
		resA.Body.Close()

		for _, entry := range analytics {
			// Check for handoffs (it might be float64 from JSON)
			if count, ok := entry["handoffs"].(float64); ok && count > 0 {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !found {
		t.Errorf("expected handoffs > 0 in analytics after retries")
	}
}

func TestHandoff_Widget(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_widget@example.com")

	// Create Bot (Handoff disabled by default)
	createBot := map[string]any{"name": "Widget Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Check Public Config
	reqP, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID, nil)
	resP, _ := http.DefaultClient.Do(reqP)
	var pubConfig struct {
		HandoffEnabled bool `json:"handoff_enabled"`
	}
	json.NewDecoder(resP.Body).Decode(&pubConfig)
	resP.Body.Close()

	if pubConfig.HandoffEnabled {
		t.Errorf("expected handoff_enabled to be false initially")
	}

	// Enable Handoff
	updateBot := map[string]any{
		"handoff_enabled": true,
		"handoff_type":    "email",
		"handoff_config": map[string]string{
			"email_to": "support@example.com",
		},
	}
	ub, _ := json.Marshal(updateBot)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqU)

	// Check Public Config Again
	reqP2, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID, nil)
	resP2, _ := http.DefaultClient.Do(reqP2)
	json.NewDecoder(resP2.Body).Decode(&pubConfig)
	resP2.Body.Close()

	if !pubConfig.HandoffEnabled {
		t.Errorf("expected handoff_enabled to be true after update")
	}
}

func TestHandoff_EdgeCases(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_edge@example.com")

	// 1. Create Bot (Handoff disabled by default)
	createBot := map[string]any{"name": "Edge Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 2. HND-002: Request Handoff when disabled
	handoffReq := map[string]any{
		"session_id": "edge-session",
		"message":    "Human please",
	}
	hr, _ := json.Marshal(handoffReq)
	reqH, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH.Header.Set("Content-Type", "application/json")
	resH, _ := http.DefaultClient.Do(reqH)

	if resH.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for disabled handoff, got %d", resH.StatusCode)
	}
	var errRes map[string]string
	json.NewDecoder(resH.Body).Decode(&errRes)
	resH.Body.Close()
	if errRes["error"] != "handoff is not enabled for this chatbot" {
		t.Errorf("unexpected error message: %s", errRes["error"])
	}

	// 3. Enable Handoff but missing email config (HND-003)
	// We enable handoff, set type to email, but leave email_to empty/nil
	updateBot := map[string]any{
		"handoff_enabled": true,
		"handoff_type":    "email",
		"handoff_config":  map[string]string{
			// "email_to" missing
		},
	}
	ub, _ := json.Marshal(updateBot)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update bot failed: %d", resU.StatusCode)
	}
	resU.Body.Close()

	// Request Handoff again
	reqH2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH2.Header.Set("Content-Type", "application/json")
	resH2, _ := http.DefaultClient.Do(reqH2)

	// Expect 200 OK but with ErrorMessage in body
	if resH2.StatusCode != http.StatusOK {
		t.Errorf("expected 200 (partial success) for misconfigured email, got %d", resH2.StatusCode)
	}
	var resBody struct {
		ErrorMessage string `json:"error_message"`
		EmailSent    bool   `json:"email_sent"`
	}
	json.NewDecoder(resH2.Body).Decode(&resBody)
	resH2.Body.Close()

	if resBody.EmailSent {
		t.Errorf("expected email_sent to be false")
	}
	// "Configuration missing for handoff email" or similar key "HANDOFF_EMAIL_NOT_CONFIGURED"
	// The default error message for "HANDOFF_EMAIL_NOT_CONFIGURED" might be "Handoff email not configured"
	if resBody.ErrorMessage == "" {
		t.Errorf("expected error message for missing config")
	}
}

func authTokenForHandoff(t *testing.T, base string, email string) string {
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	http.Post(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": "pass1234"}
	lbj, _ := json.Marshal(lb)
	res, _ := http.Post(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	var tr struct {
		Token string `json:"token"`
	}
	json.NewDecoder(res.Body).Decode(&tr)
	res.Body.Close()
	return tr.Token
}

func TestHandoff_Status_Lifecycle(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_status@example.com")

	// 1. Create Bot & Enable Handoff
	createBot := map[string]any{"name": "Status Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	updateBot := map[string]any{
		"handoff_enabled": true,
		"handoff_type":    "email",
		"handoff_config": map[string]string{
			"email_to": "support@example.com",
		},
	}
	ub, _ := json.Marshal(updateBot)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqU)

	// 2. Create Request
	handoffReq := map[string]any{
		"session_id": "status-session",
		"message":    "Status check",
	}
	hr, _ := json.Marshal(handoffReq)
	reqH, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH.Header.Set("Content-Type", "application/json")
	resH, _ := http.DefaultClient.Do(reqH)
	var hRes struct {
		RequestID string `json:"request_id"`
	}
	json.NewDecoder(resH.Body).Decode(&hRes)
	resH.Body.Close()
	reqID := hRes.RequestID

	// 3. Cycle: Pending -> Assigned -> Resolved

	// Check Initial (Pending)
	reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests", nil)
	reqG.Header.Set("Authorization", "Bearer "+token)
	resG, _ := http.DefaultClient.Do(reqG)
	var listRes struct {
		Requests []models.HandoffRequest `json:"requests"`
	}
	json.NewDecoder(resG.Body).Decode(&listRes)
	resG.Body.Close()
	if len(listRes.Requests) == 0 {
		t.Fatalf("expected requests, got 0")
	}
	if listRes.Requests[0].Status != "pending" {
		t.Fatalf("expected pending, got %s", listRes.Requests[0].Status)
	}

	// Update to Assigned
	update1 := map[string]string{"status": "assigned"}
	u1, _ := json.Marshal(update1)
	reqUp1, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests/"+reqID, bytes.NewReader(u1))
	reqUp1.Header.Set("Authorization", "Bearer "+token)
	reqUp1.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqUp1)

	// Verify Assigned
	resG2, _ := http.DefaultClient.Do(reqG)
	json.NewDecoder(resG2.Body).Decode(&listRes)
	resG2.Body.Close()
	if listRes.Requests[0].Status != "assigned" {
		t.Fatalf("expected assigned, got %s", listRes.Requests[0].Status)
	}

	// Update to Resolved
	update2 := map[string]string{"status": "resolved"}
	u2, _ := json.Marshal(update2)
	reqUp2, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests/"+reqID, bytes.NewReader(u2))
	reqUp2.Header.Set("Authorization", "Bearer "+token)
	reqUp2.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqUp2)

	// Verify Resolved
	resG3, _ := http.DefaultClient.Do(reqG)
	json.NewDecoder(resG3.Body).Decode(&listRes)
	resG3.Body.Close()
	if listRes.Requests[0].Status != "resolved" {
		t.Fatalf("expected resolved, got %s", listRes.Requests[0].Status)
	}
}

func TestHandoff_DuplicateRequest(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_dup@example.com")

	// 1. Create Bot & Enable Handoff
	createBot := map[string]any{"name": "Dup Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	updateBot := map[string]any{
		"handoff_enabled": true,
		"handoff_type":    "email",
		"handoff_config": map[string]string{
			"email_to": "support@example.com",
		},
	}
	ub, _ := json.Marshal(updateBot)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqU)

	// 2. Make First Request
	handoffReq := map[string]any{
		"session_id": "dup-session",
		"message":    "First request",
	}
	hr, _ := json.Marshal(handoffReq)
	reqH, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH.Header.Set("Content-Type", "application/json")
	resH, _ := http.DefaultClient.Do(reqH)
	if resH.StatusCode != http.StatusOK {
		t.Fatalf("first handoff failed: %d", resH.StatusCode)
	}
	var resBody1 struct {
		RequestID string `json:"request_id"`
	}
	json.NewDecoder(resH.Body).Decode(&resBody1)
	resH.Body.Close()

	// 3. Make Second Request (Should Fail)
	reqH2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH2.Header.Set("Content-Type", "application/json")
	resH2, _ := http.DefaultClient.Do(reqH2)
	
	// Expect failure because pending request exists
	if resH2.StatusCode != http.StatusInternalServerError {
		// Note: The service returns generic error, which handler maps to 500 currently. 
		// Ideally it should be 400 or 409, but based on current implementation:
		// Service returns error -> Handler logs it -> Handler returns 500 "failed to create handoff request"
		t.Logf("Got status code %d for duplicate request", resH2.StatusCode)
	}
	// We can check the body to be sure it's the expected error if we propagated it,
	// but currently handler masks errors unless we update it too.
	// For now, let's just ensure we can't create another one easily or check response.
	// Actually, wait, let's verify if the error returned is indeed what we expect.
	// The service returns "HANDOFF_ALREADY_EXISTS" (from config).
	// The public handler does: 
	// if err != nil { w.WriteHeader(http.StatusInternalServerError); ... "error": "failed to create handoff request" }
	// So we expect 500.

	if resH2.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 for duplicate request, got %d", resH2.StatusCode)
	}
	resH2.Body.Close()

	// 4. Resolve First Request
	updateReq := map[string]string{"status": "resolved"}
	ur, _ := json.Marshal(updateReq)
	reqUp, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests/"+resBody1.RequestID, bytes.NewReader(ur))
	reqUp.Header.Set("Authorization", "Bearer "+token)
	reqUp.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqUp)

	// 5. Make Third Request (Should Succeed)
	reqH3, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH3.Header.Set("Content-Type", "application/json")
	resH3, _ := http.DefaultClient.Do(reqH3)
	if resH3.StatusCode != http.StatusOK {
		t.Errorf("expected success for new request after resolution, got %d", resH3.StatusCode)
	}
	resH3.Body.Close()
}
