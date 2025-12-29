package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

// TestMockChatbotRepo_InterfaceCompliance ensures MockChatbotRepo implements ChatbotRepository
func TestMockChatbotRepo_InterfaceCompliance(t *testing.T) {
	var _ ChatbotRepository = (*MockChatbotRepo)(nil)
}

// TestPostgresChatbotRepo_InterfaceCompliance ensures PostgresChatbotRepo implements ChatbotRepository
func TestPostgresChatbotRepo_InterfaceCompliance(t *testing.T) {
	var _ ChatbotRepository = (*PostgresChatbotRepo)(nil)
}

// TestNewMockChatbotRepo verifies that NewMockChatbotRepo creates a valid mock
func TestNewMockChatbotRepo(t *testing.T) {
	mock := NewMockChatbotRepo()
	if mock == nil {
		t.Fatal("NewMockChatbotRepo returned nil")
	}
}

// TestNewPostgresChatbotRepo verifies that NewPostgresChatbotRepo creates a valid repo
func TestNewPostgresChatbotRepo(t *testing.T) {
	repo := NewPostgresChatbotRepo(nil)
	if repo == nil {
		t.Fatal("NewPostgresChatbotRepo returned nil")
	}
}

// TestMockChatbotRepo_GetByID tests the GetByID mock functionality
func TestMockChatbotRepo_GetByID(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		result, err := mock.GetByID(context.Background(), "test-id")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByID))
		}
		if mock.Calls.GetByID[0].ID != "test-id" {
			t.Errorf("expected call with ID 'test-id', got: %s", mock.Calls.GetByID[0].ID)
		}
	})

	t.Run("custom function returns chatbot", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedBot := &models.Chatbot{
			ID:   "chatbot-123",
			Name: "Test Bot",
		}
		mock.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
			if id == "chatbot-123" {
				return expectedBot, nil
			}
			return nil, nil
		}

		result, err := mock.GetByID(context.Background(), "chatbot-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != expectedBot {
			t.Errorf("expected chatbot, got: %v", result)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedErr := errors.New("database connection failed")
		mock.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
			return nil, expectedErr
		}

		_, err := mock.GetByID(context.Background(), "any-id")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockChatbotRepo_GetByUserID tests the GetByUserID mock functionality
func TestMockChatbotRepo_GetByUserID(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		result, err := mock.GetByUserID(context.Background(), "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByUserID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByUserID))
		}
	})

	t.Run("custom function returns chatbots", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedBots := []models.Chatbot{
			{ID: "bot-1", Name: "Bot 1"},
			{ID: "bot-2", Name: "Bot 2"},
		}
		mock.GetByUserIDFunc = func(ctx context.Context, userID string) ([]models.Chatbot, error) {
			return expectedBots, nil
		}

		result, err := mock.GetByUserID(context.Background(), "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 chatbots, got: %d", len(result))
		}
	})
}

// TestMockChatbotRepo_GetByWorkspace tests the GetByWorkspace mock functionality
func TestMockChatbotRepo_GetByWorkspace(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		result, err := mock.GetByWorkspace(context.Background(), "workspace-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByWorkspace) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByWorkspace))
		}
	})

	t.Run("custom function returns chatbots", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedBots := []models.Chatbot{
			{ID: "bot-1", Name: "Workspace Bot"},
		}
		mock.GetByWorkspaceFunc = func(ctx context.Context, workspaceID string) ([]models.Chatbot, error) {
			return expectedBots, nil
		}

		result, err := mock.GetByWorkspace(context.Background(), "workspace-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 chatbot, got: %d", len(result))
		}
	})
}

// TestMockChatbotRepo_Create tests the Create mock functionality
func TestMockChatbotRepo_Create(t *testing.T) {
	t.Run("default returns empty string", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		bot := &models.Chatbot{Name: "New Bot"}
		id, err := mock.Create(context.Background(), bot)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if id != "" {
			t.Errorf("expected empty string, got: %s", id)
		}
		if len(mock.Calls.Create) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.Create))
		}
		if mock.Calls.Create[0].Bot != bot {
			t.Errorf("expected bot to be recorded in call")
		}
	})

	t.Run("custom function returns generated ID", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		mock.CreateFunc = func(ctx context.Context, bot *models.Chatbot) (string, error) {
			return "generated-uuid-123", nil
		}

		bot := &models.Chatbot{Name: "New Bot"}
		id, err := mock.Create(context.Background(), bot)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if id != "generated-uuid-123" {
			t.Errorf("expected 'generated-uuid-123', got: %s", id)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedErr := errors.New("validation failed")
		mock.CreateFunc = func(ctx context.Context, bot *models.Chatbot) (string, error) {
			return "", expectedErr
		}

		bot := &models.Chatbot{Name: ""}
		_, err := mock.Create(context.Background(), bot)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockChatbotRepo_Update tests the Update mock functionality
func TestMockChatbotRepo_Update(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		bot := &models.Chatbot{ID: "bot-1", Name: "Updated Bot"}
		err := mock.Update(context.Background(), bot)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.Update) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.Update))
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedErr := errors.New("not found")
		mock.UpdateFunc = func(ctx context.Context, bot *models.Chatbot) error {
			return expectedErr
		}

		bot := &models.Chatbot{ID: "non-existent"}
		err := mock.Update(context.Background(), bot)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockChatbotRepo_SoftDelete tests the SoftDelete mock functionality
func TestMockChatbotRepo_SoftDelete(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		sourceIDs, err := mock.SoftDelete(context.Background(), "bot-1", "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if sourceIDs != nil {
			t.Errorf("expected nil source IDs, got: %v", sourceIDs)
		}
		if len(mock.Calls.SoftDelete) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.SoftDelete))
		}
		if mock.Calls.SoftDelete[0].ID != "bot-1" || mock.Calls.SoftDelete[0].UserID != "user-1" {
			t.Errorf("expected call with ID 'bot-1' and UserID 'user-1'")
		}
	})

	t.Run("custom function returns source IDs", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedSourceIDs := []string{"source-1", "source-2"}
		mock.SoftDeleteFunc = func(ctx context.Context, id, userID string) ([]string, error) {
			return expectedSourceIDs, nil
		}

		sourceIDs, err := mock.SoftDelete(context.Background(), "bot-1", "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(sourceIDs) != 2 {
			t.Errorf("expected 2 source IDs, got: %d", len(sourceIDs))
		}
	})
}

// TestMockChatbotRepo_CountByUserID tests the CountByUserID mock functionality
func TestMockChatbotRepo_CountByUserID(t *testing.T) {
	t.Run("default returns 0", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		count, err := mock.CountByUserID(context.Background(), "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got: %d", count)
		}
		if len(mock.Calls.CountByUserID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.CountByUserID))
		}
	})

	t.Run("custom function returns count", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		mock.CountByUserIDFunc = func(ctx context.Context, userID string) (int, error) {
			return 5, nil
		}

		count, err := mock.CountByUserID(context.Background(), "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if count != 5 {
			t.Errorf("expected 5, got: %d", count)
		}
	})
}

// TestMockChatbotRepo_CountByWorkspace tests the CountByWorkspace mock functionality
func TestMockChatbotRepo_CountByWorkspace(t *testing.T) {
	t.Run("default returns 0", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		count, err := mock.CountByWorkspace(context.Background(), "workspace-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got: %d", count)
		}
		if len(mock.Calls.CountByWorkspace) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.CountByWorkspace))
		}
	})

	t.Run("custom function returns count", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		mock.CountByWorkspaceFunc = func(ctx context.Context, workspaceID string) (int, error) {
			return 10, nil
		}

		count, err := mock.CountByWorkspace(context.Background(), "workspace-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if count != 10 {
			t.Errorf("expected 10, got: %d", count)
		}
	})
}

// TestMockChatbotRepo_UpdateSuggestedQuestions tests the UpdateSuggestedQuestions mock functionality
func TestMockChatbotRepo_UpdateSuggestedQuestions(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		suggestions := []string{"question 1", "question 2"}
		err := mock.UpdateSuggestedQuestions(context.Background(), "bot-1", suggestions)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.UpdateSuggestedQuestions) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.UpdateSuggestedQuestions))
		}
		call := mock.Calls.UpdateSuggestedQuestions[0]
		if call.ID != "bot-1" {
			t.Errorf("expected ID 'bot-1', got: %s", call.ID)
		}
		if len(call.Suggestions) != 2 {
			t.Errorf("expected 2 suggestions, got: %d", len(call.Suggestions))
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockChatbotRepo()
		expectedErr := errors.New("chatbot not found")
		mock.UpdateSuggestedQuestionsFunc = func(ctx context.Context, id string, suggestions []string) error {
			return expectedErr
		}

		err := mock.UpdateSuggestedQuestions(context.Background(), "non-existent", nil)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockChatbotRepo_Reset tests that Reset clears all recorded calls
func TestMockChatbotRepo_Reset(t *testing.T) {
	mock := NewMockChatbotRepo()
	ctx := context.Background()

	// Make various calls
	_, _ = mock.GetByID(ctx, "id")
	_, _ = mock.GetByUserID(ctx, "user")
	_, _ = mock.GetByWorkspace(ctx, "workspace")
	_, _ = mock.Create(ctx, &models.Chatbot{})
	_ = mock.Update(ctx, &models.Chatbot{})
	_, _ = mock.SoftDelete(ctx, "id", "user")
	_, _ = mock.CountByUserID(ctx, "user")
	_, _ = mock.CountByWorkspace(ctx, "workspace")
	_ = mock.UpdateSuggestedQuestions(ctx, "id", []string{})

	// Verify calls were recorded
	if len(mock.Calls.GetByID) != 1 {
		t.Error("expected calls to be recorded")
	}

	// Reset
	mock.Reset()

	// Verify all calls are cleared
	if len(mock.Calls.GetByID) != 0 || len(mock.Calls.GetByUserID) != 0 ||
		len(mock.Calls.GetByWorkspace) != 0 || len(mock.Calls.Create) != 0 ||
		len(mock.Calls.Update) != 0 || len(mock.Calls.SoftDelete) != 0 ||
		len(mock.Calls.CountByUserID) != 0 || len(mock.Calls.CountByWorkspace) != 0 ||
		len(mock.Calls.UpdateSuggestedQuestions) != 0 {
		t.Error("expected all calls to be cleared after Reset")
	}
}

// TestMockChatbotRepo_MultipleCalls tests recording multiple calls
func TestMockChatbotRepo_MultipleCalls(t *testing.T) {
	mock := NewMockChatbotRepo()
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

// TestMockChatbotRepo_ContextPropagation verifies context is passed to custom functions
func TestMockChatbotRepo_ContextPropagation(t *testing.T) {
	mock := NewMockChatbotRepo()
	type ctxKey string
	key := ctxKey("testKey")
	ctx := context.WithValue(context.Background(), key, "testValue")

	var receivedCtx context.Context
	mock.GetByIDFunc = func(c context.Context, id string) (*models.Chatbot, error) {
		receivedCtx = c
		return nil, nil
	}

	_, _ = mock.GetByID(ctx, "test-id")

	if receivedCtx.Value(key) != "testValue" {
		t.Error("context was not properly propagated")
	}
}

// TestMockChatbotRepo_ComplexScenario demonstrates a realistic usage pattern
func TestMockChatbotRepo_ComplexScenario(t *testing.T) {
	mock := NewMockChatbotRepo()
	ctx := context.Background()

	// Setup for a "create chatbot" service test
	mock.CountByUserIDFunc = func(ctx context.Context, userID string) (int, error) {
		if userID == "user-1" {
			return 2, nil // User has 2 chatbots
		}
		return 0, nil
	}

	mock.CreateFunc = func(ctx context.Context, bot *models.Chatbot) (string, error) {
		return "new-bot-id", nil
	}

	// Simulate a service that checks count before creating
	userID := "user-1"
	count, err := mock.CountByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count >= 10 {
		t.Fatal("user has reached chatbot limit")
	}

	newBot := &models.Chatbot{
		UserID: userID,
		Name:   "New Chatbot",
	}
	id, err := mock.Create(ctx, newBot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "new-bot-id" {
		t.Errorf("expected 'new-bot-id', got: %s", id)
	}

	// Verify the sequence of operations
	if len(mock.Calls.CountByUserID) != 1 {
		t.Error("expected CountByUserID to be called once")
	}
	if len(mock.Calls.Create) != 1 {
		t.Error("expected Create to be called once")
	}
	if mock.Calls.Create[0].Bot.Name != "New Chatbot" {
		t.Error("expected bot name to be 'New Chatbot'")
	}
}

// TestMockChatbotRepo_ErrorScenarios tests various error handling scenarios
func TestMockChatbotRepo_ErrorScenarios(t *testing.T) {
	testCases := []struct {
		name       string
		setupMock  func(*MockChatbotRepo)
		execute    func(context.Context, *MockChatbotRepo) error
		wantErrMsg string
	}{
		{
			name: "GetByID database error",
			setupMock: func(m *MockChatbotRepo) {
				m.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
					return nil, errors.New("connection refused")
				}
			},
			execute: func(ctx context.Context, m *MockChatbotRepo) error {
				_, err := m.GetByID(ctx, "id")
				return err
			},
			wantErrMsg: "connection refused",
		},
		{
			name: "Create validation error",
			setupMock: func(m *MockChatbotRepo) {
				m.CreateFunc = func(ctx context.Context, bot *models.Chatbot) (string, error) {
					if bot.Name == "" {
						return "", errors.New("name is required")
					}
					return "id", nil
				}
			},
			execute: func(ctx context.Context, m *MockChatbotRepo) error {
				_, err := m.Create(ctx, &models.Chatbot{Name: ""})
				return err
			},
			wantErrMsg: "name is required",
		},
		{
			name: "SoftDelete not found",
			setupMock: func(m *MockChatbotRepo) {
				m.SoftDeleteFunc = func(ctx context.Context, id, userID string) ([]string, error) {
					return nil, errors.New("chatbot not found")
				}
			},
			execute: func(ctx context.Context, m *MockChatbotRepo) error {
				_, err := m.SoftDelete(ctx, "non-existent", "user")
				return err
			},
			wantErrMsg: "chatbot not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockChatbotRepo()
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

// TestMockChatbotRepo_Stateful demonstrates using the mock for stateful testing
func TestMockChatbotRepo_Stateful(t *testing.T) {
	mock := NewMockChatbotRepo()
	ctx := context.Background()

	// Create an in-memory store for testing
	store := make(map[string]*models.Chatbot)
	idCounter := 0

	mock.CreateFunc = func(ctx context.Context, bot *models.Chatbot) (string, error) {
		idCounter++
		id := "chatbot-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+idCounter))
		newBot := *bot // Copy
		newBot.ID = id
		store[id] = &newBot
		return id, nil
	}

	mock.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		if bot, ok := store[id]; ok {
			return bot, nil
		}
		return nil, nil // Not found
	}

	mock.UpdateFunc = func(ctx context.Context, bot *models.Chatbot) error {
		if _, ok := store[bot.ID]; !ok {
			return errors.New("not found")
		}
		store[bot.ID] = bot
		return nil
	}

	// Test: Create -> Get -> Update -> Get
	newBot := &models.Chatbot{Name: "Original Name", UserID: "user-1"}
	id, err := mock.Create(ctx, newBot)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := mock.GetByID(ctx, id)
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

	got2, err := mock.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if got2.Name != "Updated Name" {
		t.Errorf("expected 'Updated Name', got: %s", got2.Name)
	}
}
