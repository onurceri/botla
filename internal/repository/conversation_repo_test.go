package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
)

// TestMockConversationRepo_InterfaceCompliance ensures MockConversationRepo implements ConversationRepository
func TestMockConversationRepo_InterfaceCompliance(t *testing.T) {
	var _ ConversationRepository = (*MockConversationRepo)(nil)
}

// TestPostgresConversationRepo_InterfaceCompliance ensures PostgresConversationRepo implements ConversationRepository
func TestPostgresConversationRepo_InterfaceCompliance(t *testing.T) {
	var _ ConversationRepository = (*PostgresConversationRepo)(nil)
}

// TestNewMockConversationRepo verifies that NewMockConversationRepo creates a valid mock
func TestNewMockConversationRepo(t *testing.T) {
	mock := NewMockConversationRepo()
	if mock == nil {
		t.Fatal("NewMockConversationRepo returned nil")
	}
}

// TestNewPostgresConversationRepo verifies that NewPostgresConversationRepo creates a valid repo
func TestNewPostgresConversationRepo(t *testing.T) {
	repo := NewPostgresConversationRepo(nil)
	if repo == nil {
		t.Fatal("NewPostgresConversationRepo returned nil")
	}
}

// TestMockConversationRepo_GetOrCreateBySessionID tests the GetOrCreateBySessionID mock functionality
func TestMockConversationRepo_GetOrCreateBySessionID(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockConversationRepo()
		result, err := mock.GetOrCreateBySessionID(context.Background(), "bot-1", "session-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetOrCreateBySessionID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetOrCreateBySessionID))
		}
		if mock.Calls.GetOrCreateBySessionID[0].ChatbotID != "bot-1" {
			t.Errorf("expected ChatbotID 'bot-1', got: %s", mock.Calls.GetOrCreateBySessionID[0].ChatbotID)
		}
		if mock.Calls.GetOrCreateBySessionID[0].SessionID != "session-1" {
			t.Errorf("expected SessionID 'session-1', got: %s", mock.Calls.GetOrCreateBySessionID[0].SessionID)
		}
	})

	t.Run("custom function returns conversation", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedConv := &models.Conversation{
			ID:        "conv-1",
			ChatbotID: "bot-1",
			SessionID: strPtr("session-1"),
		}
		mock.GetOrCreateBySessionIDFunc = func(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error) {
			return expectedConv, nil
		}

		result, err := mock.GetOrCreateBySessionID(context.Background(), "bot-1", "session-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != expectedConv {
			t.Errorf("expected conversation, got: %v", result)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedErr := errors.New("database connection failed")
		mock.GetOrCreateBySessionIDFunc = func(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error) {
			return nil, expectedErr
		}

		_, err := mock.GetOrCreateBySessionID(context.Background(), "any-bot", "any-session")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockConversationRepo_GetByID tests the GetByID mock functionality
func TestMockConversationRepo_GetByID(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockConversationRepo()
		result, err := mock.GetByID(context.Background(), "conv-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByID))
		}
		if mock.Calls.GetByID[0].ID != "conv-1" {
			t.Errorf("expected ID 'conv-1', got: %s", mock.Calls.GetByID[0].ID)
		}
	})

	t.Run("custom function returns conversation", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedConv := &models.Conversation{
			ID:        "conv-1",
			ChatbotID: "bot-1",
			SessionID: strPtr("session-1"),
		}
		mock.GetByIDFunc = func(ctx context.Context, id string) (*models.Conversation, error) {
			return expectedConv, nil
		}

		result, err := mock.GetByID(context.Background(), "conv-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != expectedConv {
			t.Errorf("expected conversation, got: %v", result)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedErr := errors.New("conversation not found")
		mock.GetByIDFunc = func(ctx context.Context, id string) (*models.Conversation, error) {
			return nil, expectedErr
		}

		_, err := mock.GetByID(context.Background(), "non-existent")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockConversationRepo_CreateMessage tests the CreateMessage mock functionality
func TestMockConversationRepo_CreateMessage(t *testing.T) {
	t.Run("default returns empty string", func(t *testing.T) {
		mock := NewMockConversationRepo()
		msg := &models.Message{
			ConversationID: "conv-1",
			Role:           "user",
			Content:        "Hello",
		}
		id, err := mock.CreateMessage(context.Background(), msg)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if id != "" {
			t.Errorf("expected empty string, got: %s", id)
		}
		if len(mock.Calls.CreateMessage) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.CreateMessage))
		}
		if mock.Calls.CreateMessage[0].Message != msg {
			t.Errorf("expected message to be passed through")
		}
	})

	t.Run("custom function returns id", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedID := "msg-123"
		mock.CreateMessageFunc = func(ctx context.Context, msg *models.Message) (string, error) {
			return expectedID, nil
		}

		msg := &models.Message{ConversationID: "conv-1", Role: "user", Content: "Hello"}
		id, err := mock.CreateMessage(context.Background(), msg)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if id != expectedID {
			t.Errorf("expected id %s, got: %s", expectedID, id)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedErr := errors.New("insert failed")
		mock.CreateMessageFunc = func(ctx context.Context, msg *models.Message) (string, error) {
			return "", expectedErr
		}

		msg := &models.Message{ConversationID: "conv-1", Role: "user", Content: "Hello"}
		_, err := mock.CreateMessage(context.Background(), msg)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockConversationRepo_GetMessages tests the GetMessages mock functionality
func TestMockConversationRepo_GetMessages(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockConversationRepo()
		result, err := mock.GetMessages(context.Background(), "conv-1", 10, 0)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetMessages) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetMessages))
		}
		if mock.Calls.GetMessages[0].ConversationID != "conv-1" {
			t.Errorf("expected ConversationID 'conv-1', got: %s", mock.Calls.GetMessages[0].ConversationID)
		}
	})

	t.Run("custom function returns messages", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedMsgs := []models.Message{
			{ID: "msg-1", ConversationID: "conv-1", Role: "user", Content: "Hello"},
			{ID: "msg-2", ConversationID: "conv-1", Role: "assistant", Content: "Hi there!"},
		}
		mock.GetMessagesFunc = func(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error) {
			return expectedMsgs, nil
		}

		result, err := mock.GetMessages(context.Background(), "conv-1", 10, 0)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 messages, got: %d", len(result))
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockConversationRepo()
		expectedErr := errors.New("query failed")
		mock.GetMessagesFunc = func(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error) {
			return nil, expectedErr
		}

		_, err := mock.GetMessages(context.Background(), "conv-1", 10, 0)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockConversationRepo_Reset tests that Reset clears recorded calls
func TestMockConversationRepo_Reset(t *testing.T) {
	mock := NewMockConversationRepo()
	_, _ = mock.GetOrCreateBySessionID(context.Background(), "bot-1", "session-1")
	_, _ = mock.GetByID(context.Background(), "conv-1")
	_, _ = mock.CreateMessage(context.Background(), &models.Message{})
	_, _ = mock.GetMessages(context.Background(), "conv-1", 10, 0)

	mock.Reset()

	if len(mock.Calls.GetOrCreateBySessionID) != 0 {
		t.Errorf("expected GetOrCreateBySessionID calls to be reset, got: %d", len(mock.Calls.GetOrCreateBySessionID))
	}
	if len(mock.Calls.GetByID) != 0 {
		t.Errorf("expected GetByID calls to be reset, got: %d", len(mock.Calls.GetByID))
	}
	if len(mock.Calls.CreateMessage) != 0 {
		t.Errorf("expected CreateMessage calls to be reset, got: %d", len(mock.Calls.CreateMessage))
	}
	if len(mock.Calls.GetMessages) != 0 {
		t.Errorf("expected GetMessages calls to be reset, got: %d", len(mock.Calls.GetMessages))
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
