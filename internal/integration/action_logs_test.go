package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/models"
)

func TestActionLogs_API(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	token := authTokenForAction(t, te.Server.URL, "action_logs_api@example.com")

	// Create chatbot
	createBot := map[string]any{"name": "Action Log API Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create bot: %d", resC.StatusCode)
	}
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Create Action
	// Use API to create action. Let's use API to be pure integration.
	createAction := map[string]any{
		"name":        "Test Action",
		"description": "A test action",
		"action_type": "http",
		"config":      map[string]string{"url": "https://example.com"},
		"parameters":  map[string]any{},
		"enabled":     true,
	}
	ca, _ := json.Marshal(createAction)
	reqA, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions", bytes.NewReader(ca))
	reqA.Header.Set("Authorization", "Bearer "+token)
	reqA.Header.Set("Content-Type", "application/json")
	resA, _ := testHTTPClient().Do(reqA)
	if resA.StatusCode != http.StatusCreated {
		t.Fatalf("create action: %d", resA.StatusCode)
	}
	var actionResp models.ChatbotAction
	json.NewDecoder(resA.Body).Decode(&actionResp)
	resA.Body.Close()

	// Create Action Logs directly in DB (to simulate execution)
	// We use te.DB (which is *sql.DB)
	reqRaw := json.RawMessage(`{"q": "hello"}`)
	resRaw := json.RawMessage(`{"a": "world"}`)
	log := &models.ActionExecutionLog{
		ChatbotID:       bot.ID,
		ActionID:        actionResp.ID,
		Status:          "success",
		RequestPayload:  &reqRaw,
		ResponsePayload: &resRaw,
		DurationMs:      123,
		CreatedAt:       time.Now(),
	}
	if err := db.CreateActionLog(context.Background(), te.DB, log); err != nil {
		t.Fatalf("failed to seed action log: %v", err)
	}

	// GET /api/v1/chatbots/{id}/actions/logs
	reqL, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions/logs", nil)
	reqL.Header.Set("Authorization", "Bearer "+token)
	resL, _ := testHTTPClient().Do(reqL)
	if resL.StatusCode != http.StatusOK {
		t.Fatalf("get logs: expected 200, got %d", resL.StatusCode)
	}

	var listResp struct {
		Logs  []models.ActionExecutionLog `json:"logs"`
		Page  int                         `json:"page"`
		Limit int                         `json:"limit"`
	}
	if err := json.NewDecoder(resL.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode logs: %v", err)
	}
	resL.Body.Close()

	if len(listResp.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(listResp.Logs))
	}
	if listResp.Logs[0].ID != log.ID {
		t.Errorf("log id mismatch")
	}
	if listResp.Logs[0].DurationMs != 123 {
		t.Errorf("duration mismatch")
	}

	// Test Pagination Params
	// Seed 25 more logs
	for i := 0; i < 25; i++ {
		l := &models.ActionExecutionLog{
			ChatbotID:  bot.ID,
			ActionID:   actionResp.ID,
			Status:     "success",
			DurationMs: i,
		}
		db.CreateActionLog(context.Background(), te.DB, l)
	}

	// Request page 1 with limit 10
	reqP, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions/logs?page=1&limit=10", nil)
	reqP.Header.Set("Authorization", "Bearer "+token)
	resP, _ := testHTTPClient().Do(reqP)
	if resP.StatusCode != http.StatusOK {
		t.Fatalf("get page 1: %d", resP.StatusCode)
	}
	json.NewDecoder(resP.Body).Decode(&listResp)
	resP.Body.Close()

	if len(listResp.Logs) != 10 {
		t.Errorf("expected 10 logs per page, got %d", len(listResp.Logs))
	}
	if listResp.Page != 1 {
		t.Errorf("expected page 1, got %d", listResp.Page)
	}
}

// Helper to get token (copied from action_test.go if not exported, but usually helper functions in same package are visible)
// authTokenForAction is in action_test.go, let's assume it's available since same package `integration`.
