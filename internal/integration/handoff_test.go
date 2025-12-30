package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/policy"
)

func TestHandoff_Flow(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Update plan config to allow handoff
	updateProPlanConfig(t, te)
	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)

	// 1. Setup User & Chatbot
	token := authTokenForHandoff(t, te.Server.URL, "handoff_user@example.com")
	// Update user to pro plan
	te.DB.Exec("UPDATE users SET plan_id=$1 WHERE email=$2", proPlanID, "handoff_user@example.com")

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
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_analytics@example.com")

	// Assign pro plan to user
	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	_, err = te.DB.Exec(`UPDATE users SET plan_id=$1 WHERE email=$2`, proPlanID, "handoff_analytics@example.com")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}

	// Update plan config to allow handoff
	updateProPlanConfig(t, te)

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
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_widget@example.com")

	// Assign pro plan to user
	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	_, err = te.DB.Exec(`UPDATE users SET plan_id=$1 WHERE email=$2`, proPlanID, "handoff_widget@example.com")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}

	// Update plan config to allow handoff
	updateProPlanConfig(t, te)

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
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_edge@example.com")

	// Assign pro plan to user
	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	_, err = te.DB.Exec(`UPDATE users SET plan_id=$1 WHERE email=$2`, proPlanID, "handoff_edge@example.com")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}

	// Update plan config to allow handoff
	updateProPlanConfig(t, te)

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
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	http.Post(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": "Test@123"}
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
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_status@example.com")

	// Assign pro plan to user
	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	_, err = te.DB.Exec(`UPDATE users SET plan_id=$1 WHERE email=$2`, proPlanID, "handoff_status@example.com")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}

	// Update plan config to allow handoff
	updateProPlanConfig(t, te)

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
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_dup@example.com")

	// Assign pro plan to user
	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	_, err = te.DB.Exec(`UPDATE users SET plan_id=$1 WHERE email=$2`, proPlanID, "handoff_dup@example.com")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}

	// Update plan config to allow handoff
	updateProPlanConfig(t, te)

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

	if resH2.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 for duplicate request, got %d", resH2.StatusCode)
	}
	var errResp struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	_ = json.NewDecoder(resH2.Body).Decode(&errResp)
	if errResp.Code != "handoff_exists" {
		t.Errorf("expected error code handoff_exists, got %q (error=%q)", errResp.Code, errResp.Error)
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

func TestHandoff_RequestDetail_NotFoundReturns404(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_detail_nf@example.com")

	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	_, err = te.DB.Exec(`UPDATE users SET plan_id=$1 WHERE email=$2`, proPlanID, "handoff_detail_nf@example.com")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}
	updateProPlanConfig(t, te)

	createBot := map[string]any{"name": "Detail NotFound Bot", "language": "en-US"}
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

	missingID := "00000000-0000-0000-0000-000000000000"
	reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests/"+missingID, nil)
	reqG.Header.Set("Authorization", "Bearer "+token)
	resG, _ := http.DefaultClient.Do(reqG)
	if resG.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resG.StatusCode)
	}
	var errResp struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	_ = json.NewDecoder(resG.Body).Decode(&errResp)
	resG.Body.Close()
	if errResp.Code != "handoff_not_found" {
		t.Fatalf("expected error code handoff_not_found, got %q (error=%q)", errResp.Code, errResp.Error)
	}
}

func TestHandoff_UpdateStatus_NotFoundReturns404(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")

	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "handoff_update_nf@example.com")

	var proPlanID string
	te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	_, err = te.DB.Exec(`UPDATE users SET plan_id=$1 WHERE email=$2`, proPlanID, "handoff_update_nf@example.com")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}
	updateProPlanConfig(t, te)

	createBot := map[string]any{"name": "Update NotFound Bot", "language": "en-US"}
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

	missingID := "00000000-0000-0000-0000-000000000000"
	updateReq := map[string]any{"status": "resolved"}
	ur, _ := json.Marshal(updateReq)
	reqUp, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/handoff-requests/"+missingID, bytes.NewReader(ur))
	reqUp.Header.Set("Authorization", "Bearer "+token)
	reqUp.Header.Set("Content-Type", "application/json")
	resUp, _ := http.DefaultClient.Do(reqUp)
	if resUp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resUp.StatusCode)
	}
	var errResp struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	_ = json.NewDecoder(resUp.Body).Decode(&errResp)
	resUp.Body.Close()
	if errResp.Code != "handoff_not_found" {
		t.Fatalf("expected error code handoff_not_found, got %q (error=%q)", errResp.Code, errResp.Error)
	}
}
