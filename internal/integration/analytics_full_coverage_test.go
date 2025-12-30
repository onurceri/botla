package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/models"
)

func TestAnalytics_FullCoverage(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := fixtures.StartQdrantStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	updateProPlanConfig(t, te)

	token, err := te.AuthToken("analytics_full@example.com")
	if err != nil {
		t.Fatalf("failed to get auth token: %v", err)
	}

	_, err = te.DB.Exec("UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = 'pro') WHERE email = 'analytics_full@example.com'")
	if err != nil {
		t.Fatalf("failed to assign pro plan: %v", err)
	}

	create := map[string]any{"name": "Analytics Full Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	sourceID := "00000000-0000-0000-0000-00000000000c"
	_, err = te.DB.Exec("INSERT INTO data_sources (id, chatbot_id, source_type, status) VALUES ($1, $2, $3, $4)",
		sourceID, bot.ID, "text", "completed")
	if err != nil {
		t.Fatalf("failed to create dummy source: %v", err)
	}

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

	// 3. Perform Chat (New Conversation)
	sess := "session-full-1"
	cr := chatReq{Message: "hello analytics", SessionID: sess}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := http.DefaultClient.Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("chat failed: %d", resCh.StatusCode)
	}
	resCh.Body.Close()

	// 4. Send Feedback (Thumbs Up)
	// Get message ID first
	var convID string
	err = te.DB.QueryRow("SELECT id FROM conversations WHERE chatbot_id=$1 AND session_id=$2", bot.ID, sess).Scan(&convID)
	if err != nil {
		t.Fatalf("conv query error: %v", err)
	}
	var msgID string
	err = te.DB.QueryRow("SELECT id FROM messages WHERE conversation_id=$1 AND role='assistant' ORDER BY created_at DESC LIMIT 1", convID).Scan(&msgID)
	if err != nil {
		t.Fatalf("msg query error: %v", err)
	}
	fb := map[string]any{"thumbs_up": true}
	fbj, _ := json.Marshal(fb)
	reqF, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/messages/"+msgID+"/feedback", bytes.NewReader(fbj))
	reqF.Header.Set("Authorization", "Bearer "+token)
	reqF.Header.Set("Content-Type", "application/json")
	resF, _ := http.DefaultClient.Do(reqF)
	if resF.StatusCode != http.StatusOK {
		t.Fatalf("feedback failed: %d", resF.StatusCode)
	}
	resF.Body.Close()

	// 5. Request Handoff
	handoffReq := map[string]any{
		"session_id": sess,
		"message":    "human please",
	}
	hr, _ := json.Marshal(handoffReq)
	reqH, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/handoff", bytes.NewReader(hr))
	reqH.Header.Set("Content-Type", "application/json")
	resH, _ := http.DefaultClient.Do(reqH)
	if resH.StatusCode != http.StatusOK {
		t.Fatalf("handoff request failed: %d", resH.StatusCode)
	}
	resH.Body.Close()

	// Wait for async analytics updates (especially handoff)
	time.Sleep(500 * time.Millisecond)

	// Verify DB directly
	var dbTotalM, dbTotalC, dbTotalTokens, dbThumbsUp, dbHandoff int
	err = te.DB.QueryRow(`
		SELECT total_messages, total_conversations, total_tokens_used, thumbs_up_count, handoff_count 
		FROM analytics 
		WHERE chatbot_id=$1 AND analytics_date=$2`,
		bot.ID, time.Now().Format("2006-01-02")).Scan(&dbTotalM, &dbTotalC, &dbTotalTokens, &dbThumbsUp, &dbHandoff)

	if err != nil {
		t.Fatalf("DB Query failed: %v", err)
	}
	t.Logf("DB State: Msg=%d, Conv=%d, Tok=%d, Up=%d, Handoff=%d", dbTotalM, dbTotalC, dbTotalTokens, dbThumbsUp, dbHandoff)

	// Check Chatbot state
	var cUserID string
	var cOrgID, cWsID *string
	err = te.DB.QueryRow("SELECT user_id, organization_id, workspace_id FROM chatbots WHERE id=$1", bot.ID).Scan(&cUserID, &cOrgID, &cWsID)
	if err != nil {
		t.Fatalf("Chatbot Query failed: %v", err)
	}
	t.Logf("Chatbot State: UserID=%s, OrgID=%v, WsID=%v", cUserID, cOrgID, cWsID)

	// 6. Verify Analytics Data via API
	reqA, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/analytics", nil)
	reqA.Header.Set("Authorization", "Bearer "+token)
	resA, err := http.DefaultClient.Do(reqA)
	if err != nil {
		t.Fatalf("Get analytics failed: %v", err)
	}
	defer resA.Body.Close()

	if resA.StatusCode != http.StatusOK {
		t.Fatalf("Get analytics returned status %d", resA.StatusCode)
	}

	// Define struct matching the actual JSON response of /api/v1/analytics
	type AnalyticsResponse struct {
		Date          string `json:"date"`
		Messages      int    `json:"messages"`
		Conversations int    `json:"conversations"`
		Tokens        int    `json:"tokens"`
		ThumbsUp      int    `json:"thumbs_up"`
		ThumbsDown    int    `json:"thumbs_down"`
		Handoffs      int    `json:"handoffs"`
	}

	var series []AnalyticsResponse
	if err := json.NewDecoder(resA.Body).Decode(&series); err != nil {
		t.Fatalf("decode analytics failed: %v", err)
	}
	t.Logf("Series: %+v", series)

	// Verify we have data for today
	today := time.Now().Format("2006-01-02")
	var todayStats AnalyticsResponse
	found := false
	for _, s := range series {
		if s.Date == today {
			todayStats = s
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("No analytics found for today %s", today)
	}

	// ANL-001: Message count increment
	if todayStats.Messages < 2 {
		t.Errorf("ANL-001 Failed: expected >= 2 messages, got %d", todayStats.Messages)
	}

	// ANL-002: Conversation tracking
	if todayStats.Conversations < 1 {
		t.Errorf("ANL-002 Failed: expected >= 1 conversation, got %d", todayStats.Conversations)
	}

	// ANL-003: Token usage tracking
	if todayStats.Tokens == 0 {
		t.Errorf("ANL-003 Failed: expected > 0 tokens, got %d", todayStats.Tokens)
	}

	// ANL-004: Feedback counts
	if todayStats.ThumbsUp < 1 {
		t.Errorf("ANL-004 Failed: expected >= 1 thumbs up, got %d", todayStats.ThumbsUp)
	}

	// ANL-005: Handoff tracking
	if todayStats.Handoffs < 1 {
		t.Errorf("ANL-005 Failed: expected >= 1 handoff, got %d", todayStats.Handoffs)
	}

	// 7. Verify Chatbot Specific Analytics (GET /api/v1/chatbots/:id/analytics/trends)
	reqCA, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/analytics/trends", nil)
	reqCA.Header.Set("Authorization", "Bearer "+token)
	resCA, errCA := http.DefaultClient.Do(reqCA)
	if errCA != nil {
		t.Fatalf("Get chatbot analytics failed: %v", errCA)
	}
	defer resCA.Body.Close()

	if resCA.StatusCode != http.StatusOK {
		t.Fatalf("Get chatbot analytics returned status %d", resCA.StatusCode)
	}
	var resp models.TrendData
	if err := json.NewDecoder(resCA.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode chatbot analytics response: %v", err)
	}

	// Verify today's data in the list
	foundBot := false
	for _, day := range resp.Daily {
		if day.Date == time.Now().Format("2006-01-02") {
			foundBot = true
			if day.TotalMessages != 4 {
				t.Errorf("Chatbot analytics TotalMessages mismatch. Got %d, want 4", day.TotalMessages)
			}
			if day.TotalConversations != 1 {
				t.Errorf("Chatbot analytics TotalConversations mismatch. Got %d, want 1", day.TotalConversations)
			}
			if day.TotalTokensUsed != 20 {
				t.Errorf("Chatbot analytics TotalTokensUsed mismatch. Got %d, want 20", day.TotalTokensUsed)
			}
			if day.ThumbsUpCount != 1 {
				t.Errorf("Chatbot analytics ThumbsUpCount mismatch. Got %d, want 1", day.ThumbsUpCount)
			}
			if day.HandoffCount != 1 {
				t.Errorf("Chatbot analytics HandoffCount mismatch. Got %d, want 1", day.HandoffCount)
			}
		}
	}
	if !foundBot {
		t.Error("Today's date not found in chatbot analytics trends")
	}

	// 8. Verify Chatbot Analytics Overview (GET /api/v1/chatbots/:id/analytics/overview)
	reqOverview, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/analytics/overview", nil)
	reqOverview.Header.Set("Authorization", "Bearer "+token)
	resOverview, errOverview := http.DefaultClient.Do(reqOverview)
	if errOverview != nil {
		t.Fatalf("Get chatbot overview failed: %v", errOverview)
	}
	defer resOverview.Body.Close()

	if resOverview.StatusCode != http.StatusOK {
		t.Fatalf("Get chatbot overview returned status %d", resOverview.StatusCode)
	}
	var overview models.AnalyticsOverview
	if err := json.NewDecoder(resOverview.Body).Decode(&overview); err != nil {
		t.Fatalf("Failed to decode chatbot overview response: %v", err)
	}
	if overview.TotalMessages != 4 {
		t.Errorf("Chatbot overview TotalMessages mismatch. Got %d, want 4", overview.TotalMessages)
	}
	if overview.TotalConversations != 1 {
		t.Errorf("Chatbot overview TotalConversations mismatch. Got %d, want 1", overview.TotalConversations)
	}
}
