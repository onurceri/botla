package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/models"
)

func TestUpdateAction_OptimisticLocking(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	token := authTokenForAction(t, te.Server.URL, "lock_user@example.com")

	// 1. Create Bot
	createBot := map[string]any{"name": "Lock Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 2. Create Action
	configMap := map[string]any{
		"url":    "https://example.com",
		"method": "GET",
	}
	configBytes, _ := json.Marshal(configMap)
	configRaw := json.RawMessage(configBytes)

	createAction := map[string]any{
		"name":        "Lock Action",
		"description": "desc",
		"action_type": "http",
		"config":      configRaw,
		"parameters":  json.RawMessage(`{}`),
		"enabled":     true,
	}
	ca, _ := json.Marshal(createAction)
	reqA, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions", bytes.NewReader(ca))
	reqA.Header.Set("Authorization", "Bearer "+token)
	reqA.Header.Set("Content-Type", "application/json")
	resA, _ := testHTTPClient().Do(reqA)
	if resA.StatusCode != http.StatusCreated && resA.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resA.Body)
		t.Fatalf("create action failed: %d body: %s", resA.StatusCode, string(body))
	}
	var action models.ChatbotAction
	json.NewDecoder(resA.Body).Decode(&action)
	resA.Body.Close()

	if action.ID == "" {
		t.Fatal("action ID is empty")
	}

	// 3. Update successful (Version 1 -> 2)
	update1 := map[string]any{
		"name":        "Updated Once",
		"action_type": "http",
		"config":      configRaw,
		"enabled":     true,
	}
	u1, _ := json.Marshal(update1)
	reqU1, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions/"+action.ID, bytes.NewReader(u1))
	reqU1.Header.Set("Authorization", "Bearer "+token)
	reqU1.Header.Set("Content-Type", "application/json")
	resU1, _ := testHTTPClient().Do(reqU1)
	if resU1.StatusCode != http.StatusOK {
		t.Fatalf("update 1 failed: %d", resU1.StatusCode)
	}
	var actionV2 models.ChatbotAction
	json.NewDecoder(resU1.Body).Decode(&actionV2)
	resU1.Body.Close()

	if actionV2.Version != 2 {
		t.Errorf("expected version 2, got %d", actionV2.Version)
	}
}

func TestUpdateAction_DB_OptimisticLocking(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForAction(t, te.Server.URL, "db_lock_user@example.com")

	// 1. Create Bot via API
	createBot := map[string]any{"name": "DB Lock Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 2. Create Action via API
	configMap := map[string]any{
		"url":    "https://example.com",
		"method": "GET",
	}
	configBytes, _ := json.Marshal(configMap)
	configRaw := json.RawMessage(configBytes)

	createAction := map[string]any{
		"name":        "Lock Action",
		"description": "desc",
		"action_type": "http",
		"config":      configRaw,
		"parameters":  json.RawMessage(`{}`),
		"enabled":     true,
	}
	ca, _ := json.Marshal(createAction)
	reqA, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions", bytes.NewReader(ca))
	reqA.Header.Set("Authorization", "Bearer "+token)
	reqA.Header.Set("Content-Type", "application/json")
	resA, _ := testHTTPClient().Do(reqA)

	if resA.StatusCode != http.StatusCreated && resA.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resA.Body)
		t.Fatalf("create action failed in DB test: %d body: %s", resA.StatusCode, string(body))
	}

	var action models.ChatbotAction
	json.NewDecoder(resA.Body).Decode(&action)
	resA.Body.Close()

	if action.ID == "" {
		t.Fatal("action ID is empty in DB test")
	}

	// 3. Direct DB Test for Optimistic Locking
	ctx := context.Background()

	// Get current version
	a1, err := db.GetActionByID(ctx, te.DB, action.ID)
	if err != nil {
		t.Fatalf("get action 1: %v", err)
	}

	// Simulate "Concurrent" Update 1 (increments version)
	a1_concurrent := *a1
	a1_concurrent.Name = "Updated Concurrent"
	if err := db.UpdateAction(ctx, te.DB, &a1_concurrent); err != nil {
		t.Fatalf("concurrent update failed: %v", err)
	}

	// Initial action "a1" is now stale (its version is old)
	a1.Name = "Updated Stale"
	err = db.UpdateAction(ctx, te.DB, a1)
	if err == nil {
		t.Fatal("expected error on stale update, got nil")
	}
	if err != db.ErrVersionConflict {
		t.Errorf("expected ErrVersionConflict, got %v", err)
	}
}

func TestGetOrCreateConversation_Concurrency(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForAction(t, te.Server.URL, "race_user@example.com")

	// 1. Create Bot via API
	createBot := map[string]any{"name": "Race Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	if bot.ID == "" {
		t.Fatal("failed to create bot for race test")
	}

	// 2. Race Condition Test
	sessionID := "race-session-integration"
	concurrency := 10
	errCh := make(chan error, concurrency)
	idCh := make(chan string, concurrency)
	ctx := context.Background()

	// Launch concurrent requests calling DB function directly (as we want to test DB logic primarily)
	// We use the integration DB connection.
	for i := 0; i < concurrency; i++ {
		go func() {
			c, err := db.GetOrCreateConversationBySessionID(ctx, te.DB, bot.ID, sessionID)
			if err != nil {
				errCh <- err
				return
			}
			idCh <- c.ID
			errCh <- nil
		}()
	}

	// Collect results
	ids := make([]string, 0)
	for i := 0; i < concurrency; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent request failed: %v", err)
		}
		if id := <-idCh; id != "" {
			ids = append(ids, id)
		}
	}

	// Verify all returned same ID
	if len(ids) == 0 {
		t.Fatal("no ids returned")
	}
	firstID := ids[0]
	for _, id := range ids {
		if id != firstID {
			t.Errorf("got different conversation IDs for same session: %s vs %s", firstID, id)
		}
	}
}
