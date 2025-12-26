package db_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/require"
)

func TestActionLog_CRUD(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)

	// 1. Create User
	uid := createUser(t, dbConn)

	// 2. Create Chatbot
	b := &models.Chatbot{
		UserID: uid, Name: "Action Log Bot",
		LanguageCode: "en-US", Model: "gpt-4",
	}
	bid, err := db.CreateChatbot(context.Background(), dbConn, b)
	require.NoError(t, err)

	// 3. Create Action
	rawConfig := json.RawMessage(`{"url": "http://example.com"}`)
	rawParams := json.RawMessage(`{}`)
	desc := "test description"
	toolName := "test_tool"
	action := &models.ChatbotAction{
		ChatbotID:   bid,
		Name:        "Test Action",
		Description: &desc,
		ActionType:  "http",
		Config:      &rawConfig,
		Parameters:  &rawParams,
		ToolName:    &toolName,
		Enabled:     true,
	}
	err = db.CreateAction(context.Background(), dbConn, action)
	require.NoError(t, err)
	aid := action.ID

	// 4. Create Action Log
	reqJSON, _ := json.Marshal(map[string]any{"key": "value"})
	resJSON, _ := json.Marshal(map[string]any{"result": "ok"})
	reqRaw := json.RawMessage(reqJSON)
	resRaw := json.RawMessage(resJSON)

	log := &models.ActionExecutionLog{
		ChatbotID:       bid,
		ActionID:        aid,
		Status:          "success",
		RequestPayload:  &reqRaw,
		ResponsePayload: &resRaw,
		DurationMs:      100,
	}

	err = db.CreateActionLog(context.Background(), dbConn, log)
	require.NoError(t, err)
	require.NotEmpty(t, log.ID)
	require.False(t, log.CreatedAt.IsZero())

	// 5. Get Action Logs
	logs, err := db.GetActionLogs(context.Background(), dbConn, bid, 10, 0)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, log.ID, logs[0].ID)
	require.Equal(t, "success", logs[0].Status)
	require.Equal(t, 100, logs[0].DurationMs)

	// 6. Test Pagination
	// Create another log
	errMsg := "timeout"
	log2 := &models.ActionExecutionLog{
		ChatbotID:    bid,
		ActionID:     aid,
		Status:       "failed",
		ErrorMessage: &errMsg,
		DurationMs:   5000,
	}
	err = db.CreateActionLog(context.Background(), dbConn, log2)
	require.NoError(t, err)

	// Get page 1 (latest first)
	logs, err = db.GetActionLogs(context.Background(), dbConn, bid, 1, 0)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, log2.ID, logs[0].ID) // Latest one first

	// Get page 2
	logs, err = db.GetActionLogs(context.Background(), dbConn, bid, 1, 1)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, log.ID, logs[0].ID)
}
