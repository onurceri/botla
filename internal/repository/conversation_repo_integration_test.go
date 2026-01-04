package repository_test

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
)

// TestPostgresConversationRepo_GetOrCreateBySessionID tests the actual database implementation
func TestPostgresConversationRepo_GetOrCreateBySessionID(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	repo := repository.NewPostgresConversationRepo(db)
	ctx := context.Background()

	// Create a chatbot first (needed for foreign key constraint)
	chatbotResult := testdb.CreateChatbot(t, repo.Pool())

	t.Run("creates new conversation", func(t *testing.T) {
		conv, err := repo.GetOrCreateBySessionID(ctx, chatbotResult.Chatbot.ID, "session-1")
		if err != nil {
			t.Fatalf("GetOrCreateBySessionID failed: %v", err)
		}
		if conv == nil {
			t.Fatal("expected conversation, got nil")
		}
		if conv.ChatbotID != chatbotResult.Chatbot.ID {
			t.Errorf("expected ChatbotID %s, got: %s", chatbotResult.Chatbot.ID, conv.ChatbotID)
		}
		if conv.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("returns existing conversation", func(t *testing.T) {
		// First call creates
		conv1, err := repo.GetOrCreateBySessionID(ctx, chatbotResult.Chatbot.ID, "session-2")
		if err != nil {
			t.Fatalf("first GetOrCreateBySessionID failed: %v", err)
		}

		// Second call should return same conversation
		conv2, err := repo.GetOrCreateBySessionID(ctx, chatbotResult.Chatbot.ID, "session-2")
		if err != nil {
			t.Fatalf("second GetOrCreateBySessionID failed: %v", err)
		}
		if conv1.ID != conv2.ID {
			t.Errorf("expected same conversation ID, got: %s and %s", conv1.ID, conv2.ID)
		}
	})
}

// TestPostgresConversationRepo_GetByID tests fetching conversation by ID
func TestPostgresConversationRepo_GetByID(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	repo := repository.NewPostgresConversationRepo(db)
	ctx := context.Background()

	t.Run("returns nil for non-existent", func(t *testing.T) {
		conv, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if conv != nil {
			t.Errorf("expected nil for non-existent, got: %v", conv)
		}
	})

	t.Run("returns conversation", func(t *testing.T) {
		// Create a conversation first
		chatbotResult := testdb.CreateChatbot(t, repo.Pool())
		created, err := repo.GetOrCreateBySessionID(ctx, chatbotResult.Chatbot.ID, "session-getbyid")
		if err != nil {
			t.Fatalf("GetOrCreateBySessionID failed: %v", err)
		}

		// Fetch by ID
		fetched, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if fetched == nil {
			t.Fatal("expected conversation, got nil")
		}
		if fetched.ID != created.ID {
			t.Errorf("expected ID %s, got: %s", created.ID, fetched.ID)
		}
	})
}

// TestPostgresConversationRepo_CreateMessage tests creating messages
func TestPostgresConversationRepo_CreateMessage(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	repo := repository.NewPostgresConversationRepo(db)
	ctx := context.Background()

	// Create a conversation first
	chatbotResult := testdb.CreateChatbot(t, repo.Pool())
	conv, err := repo.GetOrCreateBySessionID(ctx, chatbotResult.Chatbot.ID, "session-createmessage")
	if err != nil {
		t.Fatalf("GetOrCreateBySessionID failed: %v", err)
	}

	t.Run("creates user message", func(t *testing.T) {
		msg := &models.Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "Hello, chatbot!",
			TokensUsed:     10,
			Type:           "normal",
		}

		id, err := repo.CreateMessage(ctx, msg)
		if err != nil {
			t.Fatalf("CreateMessage failed: %v", err)
		}
		if id == "" {
			t.Error("expected non-empty message ID")
		}
	})

	t.Run("creates assistant message", func(t *testing.T) {
		msg := &models.Message{
			ConversationID: conv.ID,
			Role:           "assistant",
			Content:        "Hello! How can I help you?",
			TokensUsed:     15,
			Type:           "normal",
		}

		id, err := repo.CreateMessage(ctx, msg)
		if err != nil {
			t.Fatalf("CreateMessage failed: %v", err)
		}
		if id == "" {
			t.Error("expected non-empty message ID")
		}
	})
}

// TestPostgresConversationRepo_GetMessages tests retrieving messages with pagination
func TestPostgresConversationRepo_GetMessages(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	repo := repository.NewPostgresConversationRepo(db)
	ctx := context.Background()

	// Create a conversation with multiple messages
	chatbotResult := testdb.CreateChatbot(t, repo.Pool())
	conv, err := repo.GetOrCreateBySessionID(ctx, chatbotResult.Chatbot.ID, "session-getmessages")
	if err != nil {
		t.Fatalf("GetOrCreateBySessionID failed: %v", err)
	}

	// Create 5 messages
	for i := 0; i < 5; i++ {
		msg := &models.Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "Message",
			TokensUsed:     5,
			Type:           "normal",
		}
		_, err := repo.CreateMessage(ctx, msg)
		if err != nil {
			t.Fatalf("CreateMessage failed: %v", err)
		}
	}

	t.Run("retrieves all messages", func(t *testing.T) {
		messages, err := repo.GetMessages(ctx, conv.ID, 100, 0)
		if err != nil {
			t.Fatalf("GetMessages failed: %v", err)
		}
		if len(messages) != 5 {
			t.Errorf("expected 5 messages, got: %d", len(messages))
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		messages, err := repo.GetMessages(ctx, conv.ID, 2, 0)
		if err != nil {
			t.Fatalf("GetMessages failed: %v", err)
		}
		if len(messages) != 2 {
			t.Errorf("expected 2 messages, got: %d", len(messages))
		}
	})

	t.Run("respects offset", func(t *testing.T) {
		messages, err := repo.GetMessages(ctx, conv.ID, 100, 3)
		if err != nil {
			t.Fatalf("GetMessages failed: %v", err)
		}
		if len(messages) != 2 {
			t.Errorf("expected 2 messages with offset, got: %d", len(messages))
		}
	})

	t.Run("returns messages in chronological order", func(t *testing.T) {
		messages, err := repo.GetMessages(ctx, conv.ID, 100, 0)
		if err != nil {
			t.Fatalf("GetMessages failed: %v", err)
		}
		for i := 1; i < len(messages); i++ {
			if messages[i].CreatedAt.Before(messages[i-1].CreatedAt) {
				t.Errorf("messages not in chronological order: %v before %v",
					messages[i].CreatedAt, messages[i-1].CreatedAt)
			}
		}
	})

	t.Run("returns empty for non-existent conversation", func(t *testing.T) {
		messages, err := repo.GetMessages(ctx, "00000000-0000-0000-0000-000000000000", 10, 0)
		if err != nil {
			t.Fatalf("GetMessages failed: %v", err)
		}
		if len(messages) != 0 {
			t.Errorf("expected 0 messages for non-existent conversation, got: %d", len(messages))
		}
	})
}
