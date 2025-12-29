package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

// TestMockActionRepo_InterfaceCompliance ensures MockActionRepo implements ActionRepository
func TestMockActionRepo_InterfaceCompliance(t *testing.T) {
	var _ ActionRepository = (*MockActionRepo)(nil)
}

// TestPostgresActionRepo_InterfaceCompliance ensures PostgresActionRepo implements ActionRepository
func TestPostgresActionRepo_InterfaceCompliance(t *testing.T) {
	var _ ActionRepository = (*PostgresActionRepo)(nil)
}

// TestNewMockActionRepo verifies that NewMockActionRepo creates a valid mock
func TestNewMockActionRepo(t *testing.T) {
	mock := NewMockActionRepo()
	if mock == nil {
		t.Fatal("NewMockActionRepo returned nil")
	}
}

// TestNewPostgresActionRepo verifies that NewPostgresActionRepo creates a valid repo
func TestNewPostgresActionRepo(t *testing.T) {
	repo := NewPostgresActionRepo(nil)
	if repo == nil {
		t.Fatal("NewPostgresActionRepo returned nil")
	}
}

// TestMockActionRepo_List tests the List mock functionality
func TestMockActionRepo_List(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		result, err := mock.List(context.Background(), "chatbot-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.List) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.List))
		}
		if mock.Calls.List[0].ChatbotID != "chatbot-1" {
			t.Errorf("expected call with ChatbotID 'chatbot-1', got: %s", mock.Calls.List[0].ChatbotID)
		}
	})

	t.Run("custom function returns actions", func(t *testing.T) {
		mock := NewMockActionRepo()
		expectedActions := []*models.ChatbotAction{
			{ID: "action-1", Name: "Action 1", ChatbotID: "chatbot-1"},
			{ID: "action-2", Name: "Action 2", ChatbotID: "chatbot-1"},
		}
		mock.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
			return expectedActions, nil
		}

		result, err := mock.List(context.Background(), "chatbot-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 actions, got: %d", len(result))
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockActionRepo()
		expectedErr := errors.New("database connection failed")
		mock.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
			return nil, expectedErr
		}

		_, err := mock.List(context.Background(), "any-chatbot")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockActionRepo_ListEnabled tests the ListEnabled mock functionality
func TestMockActionRepo_ListEnabled(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		result, err := mock.ListEnabled(context.Background(), "chatbot-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.ListEnabled) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.ListEnabled))
		}
	})

	t.Run("custom function returns only enabled actions", func(t *testing.T) {
		mock := NewMockActionRepo()
		mock.ListEnabledFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
			return []*models.ChatbotAction{
				{ID: "action-1", Name: "Enabled Action", Enabled: true},
			}, nil
		}

		result, err := mock.ListEnabled(context.Background(), "chatbot-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 action, got: %d", len(result))
		}
		if !result[0].Enabled {
			t.Error("expected action to be enabled")
		}
	})
}

// TestMockActionRepo_GetByID tests the GetByID mock functionality
func TestMockActionRepo_GetByID(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		result, err := mock.GetByID(context.Background(), "action-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByID))
		}
		if mock.Calls.GetByID[0].ID != "action-1" {
			t.Errorf("expected call with ID 'action-1', got: %s", mock.Calls.GetByID[0].ID)
		}
	})

	t.Run("custom function returns action", func(t *testing.T) {
		mock := NewMockActionRepo()
		expectedAction := &models.ChatbotAction{
			ID:         "action-123",
			Name:       "Test Action",
			ActionType: models.ActionTypeHTTP,
		}
		mock.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
			if id == "action-123" {
				return expectedAction, nil
			}
			return nil, nil
		}

		result, err := mock.GetByID(context.Background(), "action-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != expectedAction {
			t.Errorf("expected action, got: %v", result)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockActionRepo()
		expectedErr := errors.New("database connection failed")
		mock.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
			return nil, expectedErr
		}

		_, err := mock.GetByID(context.Background(), "any-id")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockActionRepo_GetByToolName tests the GetByToolName mock functionality
func TestMockActionRepo_GetByToolName(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		result, err := mock.GetByToolName(context.Background(), "chatbot-1", "get_weather")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByToolName) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByToolName))
		}
		call := mock.Calls.GetByToolName[0]
		if call.ChatbotID != "chatbot-1" || call.ToolName != "get_weather" {
			t.Errorf("expected call with ChatbotID 'chatbot-1' and ToolName 'get_weather', got: %v", call)
		}
	})

	t.Run("custom function returns action", func(t *testing.T) {
		mock := NewMockActionRepo()
		toolName := "submit_form"
		mock.GetByToolNameFunc = func(ctx context.Context, chatbotID, tn string) (*models.ChatbotAction, error) {
			if tn == toolName {
				return &models.ChatbotAction{
					ID:       "action-1",
					ToolName: &toolName,
					Enabled:  true,
				}, nil
			}
			return nil, nil
		}

		result, err := mock.GetByToolName(context.Background(), "chatbot-1", toolName)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil || *result.ToolName != toolName {
			t.Errorf("expected action with tool_name %s", toolName)
		}
	})
}

// TestMockActionRepo_Create tests the Create mock functionality
func TestMockActionRepo_Create(t *testing.T) {
	t.Run("default returns nil error", func(t *testing.T) {
		mock := NewMockActionRepo()
		action := &models.ChatbotAction{Name: "New Action", ChatbotID: "chatbot-1"}
		err := mock.Create(context.Background(), action)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.Create) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.Create))
		}
		if mock.Calls.Create[0].Action != action {
			t.Error("expected action to be recorded in call")
		}
	})

	t.Run("custom function populates fields", func(t *testing.T) {
		mock := NewMockActionRepo()
		mock.CreateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
			action.ID = "generated-uuid-123"
			action.Version = 1
			now := time.Now()
			action.CreatedAt = now
			action.UpdatedAt = &now
			return nil
		}

		action := &models.ChatbotAction{Name: "New Action"}
		err := mock.Create(context.Background(), action)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if action.ID != "generated-uuid-123" {
			t.Errorf("expected ID 'generated-uuid-123', got: %s", action.ID)
		}
		if action.Version != 1 {
			t.Errorf("expected Version 1, got: %d", action.Version)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockActionRepo()
		expectedErr := errors.New("validation failed")
		mock.CreateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
			return expectedErr
		}

		action := &models.ChatbotAction{Name: ""}
		err := mock.Create(context.Background(), action)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockActionRepo_Update tests the Update mock functionality
func TestMockActionRepo_Update(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		action := &models.ChatbotAction{ID: "action-1", Name: "Updated Action"}
		err := mock.Update(context.Background(), action)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.Update) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.Update))
		}
	})

	t.Run("custom function returns version conflict", func(t *testing.T) {
		mock := NewMockActionRepo()
		mock.UpdateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
			return ErrVersionConflict
		}

		action := &models.ChatbotAction{ID: "action-1", Version: 1}
		err := mock.Update(context.Background(), action)
		if err != ErrVersionConflict {
			t.Errorf("expected ErrVersionConflict, got: %v", err)
		}
	})

	t.Run("custom function increments version", func(t *testing.T) {
		mock := NewMockActionRepo()
		mock.UpdateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
			action.Version++
			now := time.Now()
			action.UpdatedAt = &now
			return nil
		}

		action := &models.ChatbotAction{ID: "action-1", Version: 1}
		err := mock.Update(context.Background(), action)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if action.Version != 2 {
			t.Errorf("expected Version 2, got: %d", action.Version)
		}
	})
}

// TestMockActionRepo_Delete tests the Delete mock functionality
func TestMockActionRepo_Delete(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		err := mock.Delete(context.Background(), "action-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.Delete) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.Delete))
		}
		if mock.Calls.Delete[0].ID != "action-1" {
			t.Errorf("expected call with ID 'action-1', got: %s", mock.Calls.Delete[0].ID)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockActionRepo()
		expectedErr := errors.New("action not found")
		mock.DeleteFunc = func(ctx context.Context, id string) error {
			return expectedErr
		}

		err := mock.Delete(context.Background(), "non-existent")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockActionRepo_GetLogs tests the GetLogs mock functionality
func TestMockActionRepo_GetLogs(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		result, err := mock.GetLogs(context.Background(), "chatbot-1", 10, 0)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetLogs) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetLogs))
		}
		call := mock.Calls.GetLogs[0]
		if call.ChatbotID != "chatbot-1" || call.Limit != 10 || call.Offset != 0 {
			t.Errorf("expected call with ChatbotID 'chatbot-1', Limit 10, Offset 0, got: %v", call)
		}
	})

	t.Run("custom function returns logs", func(t *testing.T) {
		mock := NewMockActionRepo()
		mock.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
			return []*models.ActionExecutionLog{
				{ID: "log-1", ChatbotID: chatbotID, ActionID: "action-1", Status: "success"},
				{ID: "log-2", ChatbotID: chatbotID, ActionID: "action-1", Status: "failure"},
			}, nil
		}

		result, err := mock.GetLogs(context.Background(), "chatbot-1", 10, 0)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 logs, got: %d", len(result))
		}
	})

	t.Run("respects pagination", func(t *testing.T) {
		mock := NewMockActionRepo()
		mock.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
			allLogs := []*models.ActionExecutionLog{
				{ID: "log-1"}, {ID: "log-2"}, {ID: "log-3"}, {ID: "log-4"}, {ID: "log-5"},
			}
			end := offset + limit
			if end > len(allLogs) {
				end = len(allLogs)
			}
			if offset >= len(allLogs) {
				return nil, nil
			}
			return allLogs[offset:end], nil
		}

		// First page
		result1, _ := mock.GetLogs(context.Background(), "chatbot-1", 2, 0)
		if len(result1) != 2 {
			t.Errorf("expected 2 logs in first page, got: %d", len(result1))
		}

		// Second page
		result2, _ := mock.GetLogs(context.Background(), "chatbot-1", 2, 2)
		if len(result2) != 2 {
			t.Errorf("expected 2 logs in second page, got: %d", len(result2))
		}

		// Last page (partial)
		result3, _ := mock.GetLogs(context.Background(), "chatbot-1", 2, 4)
		if len(result3) != 1 {
			t.Errorf("expected 1 log in last page, got: %d", len(result3))
		}
	})
}

// TestMockActionRepo_CreateLog tests the CreateLog mock functionality
func TestMockActionRepo_CreateLog(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockActionRepo()
		log := &models.ActionExecutionLog{
			ChatbotID:  "chatbot-1",
			ActionID:   "action-1",
			Status:     "success",
			DurationMs: 150,
		}
		err := mock.CreateLog(context.Background(), log)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.CreateLog) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.CreateLog))
		}
		if mock.Calls.CreateLog[0].Log != log {
			t.Error("expected log to be recorded in call")
		}
	})

	t.Run("custom function populates ID and timestamp", func(t *testing.T) {
		mock := NewMockActionRepo()
		mock.CreateLogFunc = func(ctx context.Context, log *models.ActionExecutionLog) error {
			log.ID = "log-uuid-123"
			log.CreatedAt = time.Now()
			return nil
		}

		log := &models.ActionExecutionLog{ChatbotID: "chatbot-1", ActionID: "action-1", Status: "success"}
		err := mock.CreateLog(context.Background(), log)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if log.ID != "log-uuid-123" {
			t.Errorf("expected ID 'log-uuid-123', got: %s", log.ID)
		}
		if log.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})
}

// TestMockActionRepo_Reset tests that Reset clears all recorded calls
func TestMockActionRepo_Reset(t *testing.T) {
	mock := NewMockActionRepo()
	ctx := context.Background()

	// Make various calls
	_, _ = mock.List(ctx, "chatbot-1")
	_, _ = mock.ListEnabled(ctx, "chatbot-1")
	_, _ = mock.GetByID(ctx, "action-1")
	_, _ = mock.GetByToolName(ctx, "chatbot-1", "tool")
	_ = mock.Create(ctx, &models.ChatbotAction{})
	_ = mock.Update(ctx, &models.ChatbotAction{})
	_ = mock.Delete(ctx, "action-1")
	_, _ = mock.GetLogs(ctx, "chatbot-1", 10, 0)
	_ = mock.CreateLog(ctx, &models.ActionExecutionLog{})

	// Verify calls were recorded
	if len(mock.Calls.List) != 1 {
		t.Error("expected List call to be recorded")
	}

	// Reset
	mock.Reset()

	// Verify all calls are cleared
	if len(mock.Calls.List) != 0 || len(mock.Calls.ListEnabled) != 0 ||
		len(mock.Calls.GetByID) != 0 || len(mock.Calls.GetByToolName) != 0 ||
		len(mock.Calls.Create) != 0 || len(mock.Calls.Update) != 0 ||
		len(mock.Calls.Delete) != 0 || len(mock.Calls.GetLogs) != 0 ||
		len(mock.Calls.CreateLog) != 0 {
		t.Error("expected all calls to be cleared after Reset")
	}
}

// TestMockActionRepo_MultipleCalls tests recording multiple calls
func TestMockActionRepo_MultipleCalls(t *testing.T) {
	mock := NewMockActionRepo()
	ctx := context.Background()

	// Make multiple GetByID calls
	_, _ = mock.GetByID(ctx, "id-1")
	_, _ = mock.GetByID(ctx, "id-2")
	_, _ = mock.GetByID(ctx, "id-3")

	if len(mock.Calls.GetByID) != 3 {
		t.Errorf("expected 3 calls, got: %d", len(mock.Calls.GetByID))
	}

	// Verify IDs are correctly recorded
	expectedIDs := []string{"id-1", "id-2", "id-3"}
	for i, call := range mock.Calls.GetByID {
		if call.ID != expectedIDs[i] {
			t.Errorf("call %d: expected ID %s, got %s", i, expectedIDs[i], call.ID)
		}
	}
}

// TestMockActionRepo_ContextPropagation verifies context is passed to custom functions
func TestMockActionRepo_ContextPropagation(t *testing.T) {
	mock := NewMockActionRepo()
	type ctxKey string
	key := ctxKey("testKey")
	ctx := context.WithValue(context.Background(), key, "testValue")

	var receivedCtx context.Context
	mock.GetByIDFunc = func(c context.Context, id string) (*models.ChatbotAction, error) {
		receivedCtx = c
		return nil, nil
	}

	_, _ = mock.GetByID(ctx, "test-id")

	if receivedCtx.Value(key) != "testValue" {
		t.Error("context was not properly propagated")
	}
}

// TestMockActionRepo_ComplexScenario demonstrates a realistic usage pattern
func TestMockActionRepo_ComplexScenario(t *testing.T) {
	mock := NewMockActionRepo()
	ctx := context.Background()

	// Setup: Simulate an action service that creates an action and logs execution
	actionStore := make(map[string]*models.ChatbotAction)
	actionIDCounter := 0

	mock.CreateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
		actionIDCounter++
		action.ID = "action-" + string(rune('0'+actionIDCounter))
		action.Version = 1
		now := time.Now()
		action.CreatedAt = now
		action.UpdatedAt = &now
		actionStore[action.ID] = action
		return nil
	}

	mock.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		if action, ok := actionStore[id]; ok {
			return action, nil
		}
		return nil, nil
	}

	mock.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		var actions []*models.ChatbotAction
		for _, action := range actionStore {
			if action.ChatbotID == chatbotID {
				actions = append(actions, action)
			}
		}
		return actions, nil
	}

	// Test: Create an action
	configData := json.RawMessage(`{"url": "https://api.example.com/webhook", "method": "POST"}`)
	newAction := &models.ChatbotAction{
		ChatbotID:  "chatbot-1",
		Name:       "My HTTP Action",
		ActionType: models.ActionTypeHTTP,
		Config:     &configData,
		Enabled:    true,
	}

	err := mock.Create(ctx, newAction)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if newAction.ID == "" {
		t.Error("expected ID to be populated after Create")
	}

	// Verify the action was stored
	got, err := mock.GetByID(ctx, newAction.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected action to be found")
	}
	if got.Name != "My HTTP Action" {
		t.Errorf("expected 'My HTTP Action', got: %s", got.Name)
	}

	// Create another action for the same chatbot
	anotherAction := &models.ChatbotAction{
		ChatbotID:  "chatbot-1",
		Name:       "Second Action",
		ActionType: models.ActionTypeZapier,
		Enabled:    false,
	}
	_ = mock.Create(ctx, anotherAction)

	// List all actions for chatbot
	actions, err := mock.List(ctx, "chatbot-1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got: %d", len(actions))
	}

	// Verify the sequence of operations
	if len(mock.Calls.Create) != 2 {
		t.Errorf("expected 2 Create calls, got: %d", len(mock.Calls.Create))
	}
	if len(mock.Calls.GetByID) != 1 {
		t.Errorf("expected 1 GetByID call, got: %d", len(mock.Calls.GetByID))
	}
	if len(mock.Calls.List) != 1 {
		t.Errorf("expected 1 List call, got: %d", len(mock.Calls.List))
	}
}

// TestMockActionRepo_OptimisticLocking tests version conflict handling
func TestMockActionRepo_OptimisticLocking(t *testing.T) {
	mock := NewMockActionRepo()
	ctx := context.Background()

	// Simulate optimistic locking
	currentVersion := 1
	mock.UpdateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
		if action.Version != currentVersion {
			return ErrVersionConflict
		}
		currentVersion++
		action.Version = currentVersion
		return nil
	}

	// First update should succeed
	action := &models.ChatbotAction{ID: "action-1", Version: 1}
	err := mock.Update(ctx, action)
	if err != nil {
		t.Errorf("expected first update to succeed, got: %v", err)
	}
	if action.Version != 2 {
		t.Errorf("expected version 2, got: %d", action.Version)
	}

	// Concurrent update with stale version should fail
	staleAction := &models.ChatbotAction{ID: "action-1", Version: 1}
	err = mock.Update(ctx, staleAction)
	if err != ErrVersionConflict {
		t.Errorf("expected ErrVersionConflict for stale update, got: %v", err)
	}

	// Update with current version should succeed
	freshAction := &models.ChatbotAction{ID: "action-1", Version: 2}
	err = mock.Update(ctx, freshAction)
	if err != nil {
		t.Errorf("expected update with fresh version to succeed, got: %v", err)
	}
}

// TestMockActionRepo_ErrorScenarios tests various error handling scenarios
func TestMockActionRepo_ErrorScenarios(t *testing.T) {
	testCases := []struct {
		name       string
		setupMock  func(*MockActionRepo)
		execute    func(context.Context, *MockActionRepo) error
		wantErrMsg string
	}{
		{
			name: "List database error",
			setupMock: func(m *MockActionRepo) {
				m.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
					return nil, errors.New("connection refused")
				}
			},
			execute: func(ctx context.Context, m *MockActionRepo) error {
				_, err := m.List(ctx, "chatbot-1")
				return err
			},
			wantErrMsg: "connection refused",
		},
		{
			name: "Create validation error",
			setupMock: func(m *MockActionRepo) {
				m.CreateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
					if action.Name == "" {
						return errors.New("name is required")
					}
					return nil
				}
			},
			execute: func(ctx context.Context, m *MockActionRepo) error {
				return m.Create(ctx, &models.ChatbotAction{Name: ""})
			},
			wantErrMsg: "name is required",
		},
		{
			name: "Update version conflict",
			setupMock: func(m *MockActionRepo) {
				m.UpdateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
					return ErrVersionConflict
				}
			},
			execute: func(ctx context.Context, m *MockActionRepo) error {
				return m.Update(ctx, &models.ChatbotAction{ID: "action-1"})
			},
			wantErrMsg: "version conflict: entity was modified by another request",
		},
		{
			name: "Delete not found",
			setupMock: func(m *MockActionRepo) {
				m.DeleteFunc = func(ctx context.Context, id string) error {
					return errors.New("action not found")
				}
			},
			execute: func(ctx context.Context, m *MockActionRepo) error {
				return m.Delete(ctx, "non-existent")
			},
			wantErrMsg: "action not found",
		},
		{
			name: "GetLogs query error",
			setupMock: func(m *MockActionRepo) {
				m.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
					return nil, errors.New("query timeout")
				}
			},
			execute: func(ctx context.Context, m *MockActionRepo) error {
				_, err := m.GetLogs(ctx, "chatbot-1", 10, 0)
				return err
			},
			wantErrMsg: "query timeout",
		},
		{
			name: "CreateLog constraint violation",
			setupMock: func(m *MockActionRepo) {
				m.CreateLogFunc = func(ctx context.Context, log *models.ActionExecutionLog) error {
					return errors.New("foreign key constraint violation")
				}
			},
			execute: func(ctx context.Context, m *MockActionRepo) error {
				return m.CreateLog(ctx, &models.ActionExecutionLog{ActionID: "non-existent"})
			},
			wantErrMsg: "foreign key constraint violation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockActionRepo()
			tc.setupMock(mock)
			err := tc.execute(context.Background(), mock)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.wantErrMsg {
				t.Errorf("expected error message %q, got %q", tc.wantErrMsg, err.Error())
			}
		})
	}
}

// TestMockActionRepo_ActionWithConfig tests handling of JSON config fields
func TestMockActionRepo_ActionWithConfig(t *testing.T) {
	mock := NewMockActionRepo()
	ctx := context.Background()

	// Test with HTTP action config
	configJSON := json.RawMessage(`{"url":"https://api.example.com","method":"POST","headers":{"Content-Type":"application/json"}}`)
	parametersJSON := json.RawMessage(`{"type":"object","properties":{"email":{"type":"string"}}}`)
	toolName := "submit_email"

	action := &models.ChatbotAction{
		ChatbotID:   "chatbot-1",
		Name:        "Submit Email",
		Description: ptrString("Submits an email to external API"),
		ActionType:  models.ActionTypeHTTP,
		Config:      &configJSON,
		Parameters:  &parametersJSON,
		ToolName:    &toolName,
		Enabled:     true,
	}

	createdID := "action-created-123"
	mock.CreateFunc = func(ctx context.Context, a *models.ChatbotAction) error {
		a.ID = createdID
		a.Version = 1
		return nil
	}

	mock.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		if id == createdID {
			return action, nil
		}
		return nil, nil
	}

	// Create
	err := mock.Create(ctx, action)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Retrieve and verify config is preserved
	got, err := mock.GetByID(ctx, createdID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Config == nil {
		t.Error("expected Config to be set")
	}
	if got.Parameters == nil {
		t.Error("expected Parameters to be set")
	}
	if got.ToolName == nil || *got.ToolName != toolName {
		t.Errorf("expected ToolName to be %s", toolName)
	}
}

// Helper function to create a pointer to a string
func ptrString(s string) *string {
	return &s
}

// TestMockActionRepo_Stateful demonstrates using the mock for stateful CRUD testing
func TestMockActionRepo_Stateful(t *testing.T) {
	mock := NewMockActionRepo()
	ctx := context.Background()

	// Create an in-memory store
	store := make(map[string]*models.ChatbotAction)
	idCounter := 0

	mock.CreateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
		idCounter++
		id := "action-" + time.Now().Format("150405") + "-" + string(rune('0'+idCounter))
		newAction := *action
		newAction.ID = id
		newAction.Version = 1
		now := time.Now()
		newAction.CreatedAt = now
		newAction.UpdatedAt = &now
		store[id] = &newAction
		action.ID = id
		action.Version = 1
		action.CreatedAt = now
		action.UpdatedAt = &now
		return nil
	}

	mock.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		if action, ok := store[id]; ok {
			return action, nil
		}
		return nil, nil
	}

	mock.UpdateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
		existing, ok := store[action.ID]
		if !ok {
			return errors.New("not found")
		}
		if existing.Version != action.Version {
			return ErrVersionConflict
		}
		action.Version++
		now := time.Now()
		action.UpdatedAt = &now
		store[action.ID] = action
		return nil
	}

	mock.DeleteFunc = func(ctx context.Context, id string) error {
		if _, ok := store[id]; !ok {
			return errors.New("not found")
		}
		delete(store, id)
		return nil
	}

	// Test: Create -> Get -> Update -> Get -> Delete -> Get
	newAction := &models.ChatbotAction{ChatbotID: "chatbot-1", Name: "Original Name", Enabled: true}
	err := mock.Create(ctx, newAction)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	actionID := newAction.ID

	got, err := mock.GetByID(ctx, actionID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "Original Name" {
		t.Errorf("expected 'Original Name', got: %s", got.Name)
	}

	got.Name = "Updated Name"
	err = mock.Update(ctx, got)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got2, err := mock.GetByID(ctx, actionID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if got2.Name != "Updated Name" {
		t.Errorf("expected 'Updated Name', got: %s", got2.Name)
	}
	if got2.Version != 2 {
		t.Errorf("expected version 2, got: %d", got2.Version)
	}

	err = mock.Delete(ctx, actionID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	got3, err := mock.GetByID(ctx, actionID)
	if err != nil {
		t.Fatalf("GetByID after delete failed: %v", err)
	}
	if got3 != nil {
		t.Error("expected nil after deletion")
	}
}

// TestMockActionRepo_ActionExecutionLogging tests the action execution logging workflow
func TestMockActionRepo_ActionExecutionLogging(t *testing.T) {
	mock := NewMockActionRepo()
	ctx := context.Background()

	// Simulate log storage
	logs := make([]*models.ActionExecutionLog, 0)

	mock.CreateLogFunc = func(ctx context.Context, log *models.ActionExecutionLog) error {
		log.ID = "log-" + time.Now().Format("150405.000")
		log.CreatedAt = time.Now()
		logs = append(logs, log)
		return nil
	}

	mock.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
		var result []*models.ActionExecutionLog
		for _, log := range logs {
			if log.ChatbotID == chatbotID {
				result = append(result, log)
			}
		}
		// Apply pagination
		if offset >= len(result) {
			return nil, nil
		}
		end := offset + limit
		if end > len(result) {
			end = len(result)
		}
		return result[offset:end], nil
	}

	// Simulate action execution
	requestPayload := json.RawMessage(`{"email":"test@example.com"}`)
	responsePayload := json.RawMessage(`{"success":true}`)
	convID := "conv-123"
	msgID := "msg-456"

	log1 := &models.ActionExecutionLog{
		ChatbotID:       "chatbot-1",
		ActionID:        "action-1",
		ConversationID:  &convID,
		MessageID:       &msgID,
		Status:          "success",
		RequestPayload:  &requestPayload,
		ResponsePayload: &responsePayload,
		DurationMs:      150,
	}

	err := mock.CreateLog(ctx, log1)
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}
	if log1.ID == "" {
		t.Error("expected log ID to be populated")
	}

	// Add another log (failure)
	errorMsg := "connection timeout"
	log2 := &models.ActionExecutionLog{
		ChatbotID:    "chatbot-1",
		ActionID:     "action-1",
		Status:       "failure",
		ErrorMessage: &errorMsg,
		DurationMs:   5000,
	}
	_ = mock.CreateLog(ctx, log2)

	// Retrieve logs
	retrievedLogs, err := mock.GetLogs(ctx, "chatbot-1", 10, 0)
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
	if len(retrievedLogs) != 2 {
		t.Errorf("expected 2 logs, got: %d", len(retrievedLogs))
	}

	// Verify log content
	successLog := retrievedLogs[0]
	if successLog.Status != "success" {
		t.Errorf("expected status 'success', got: %s", successLog.Status)
	}
	if successLog.DurationMs != 150 {
		t.Errorf("expected duration 150ms, got: %d", successLog.DurationMs)
	}
}

// TestErrVersionConflict tests the ErrVersionConflict error behavior
func TestErrVersionConflict(t *testing.T) {
	err := ErrVersionConflict

	// Test that it implements error interface
	var e error = err
	_ = e

	// Test error message
	expectedMsg := "version conflict: entity was modified by another request"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}

	// Test that it's not equal to a different error
	otherErr := errors.New("some other error")
	if errors.Is(otherErr, ErrVersionConflict) {
		t.Error("expected different errors to not be equal")
	}
}
