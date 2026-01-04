package repository

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

// MockConversationRepo is a mock implementation of ConversationRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockConversationRepo struct {
	// GetOrCreateBySessionIDFunc is called when GetOrCreateBySessionID is invoked.
	GetOrCreateBySessionIDFunc func(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error)

	// GetByIDFunc is called when GetByID is invoked.
	GetByIDFunc func(ctx context.Context, id string) (*models.Conversation, error)

	// CreateMessageFunc is called when CreateMessage is invoked.
	CreateMessageFunc func(ctx context.Context, msg *models.Message) (string, error)

	// GetMessagesFunc is called when GetMessages is invoked.
	GetMessagesFunc func(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error)

	// IncrementMessageCountFunc is called when IncrementMessageCount is invoked.
	IncrementMessageCountFunc func(ctx context.Context, conversationID string) error

	// ListRecentMessagesFunc is called when ListRecentMessages is invoked.
	ListRecentMessagesFunc func(ctx context.Context, conversationID string, limit int) ([]models.Message, error)

	// SaveMessageSourcesFunc is called when SaveMessageSources is invoked.
	SaveMessageSourcesFunc func(ctx context.Context, messageID string, sources []models.ChunkMetadata) error

	// Invocation tracking for test assertions
	Calls struct {
		GetOrCreateBySessionID []ConversationGetOrCreateBySessionIDCall
		GetByID                []ConversationGetByIDCall
		CreateMessage          []ConversationCreateMessageCall
		GetMessages            []ConversationGetMessagesCall
		IncrementMessageCount  []ConversationIncrementMessageCountCall
		ListRecentMessages     []ConversationListRecentMessagesCall
		SaveMessageSources     []ConversationSaveMessageSourcesCall
	}
}

// Call recording types for test verification
type ConversationGetOrCreateBySessionIDCall struct {
	ChatbotID string
	SessionID string
}

type ConversationGetByIDCall struct {
	ID string
}

type ConversationCreateMessageCall struct {
	Message *models.Message
}

type ConversationGetMessagesCall struct {
	ConversationID string
	Limit          int
	Offset         int
}

type ConversationIncrementMessageCountCall struct {
	ConversationID string
}

type ConversationListRecentMessagesCall struct {
	ConversationID string
	Limit          int
}

type ConversationSaveMessageSourcesCall struct {
	MessageID string
	Sources   []models.ChunkMetadata
}

// Compile-time check that MockConversationRepo implements ConversationRepository.
var _ ConversationRepository = (*MockConversationRepo)(nil)

// NewMockConversationRepo creates a new MockConversationRepo with default no-op behavior.
func NewMockConversationRepo() *MockConversationRepo {
	return &MockConversationRepo{}
}

// GetOrCreateBySessionID finds an existing conversation or creates a new one.
func (m *MockConversationRepo) GetOrCreateBySessionID(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error) {
	m.Calls.GetOrCreateBySessionID = append(m.Calls.GetOrCreateBySessionID, ConversationGetOrCreateBySessionIDCall{
		ChatbotID: chatbotID,
		SessionID: sessionID,
	})
	if m.GetOrCreateBySessionIDFunc != nil {
		return m.GetOrCreateBySessionIDFunc(ctx, chatbotID, sessionID)
	}
	return nil, nil
}

// GetByID retrieves a conversation by its unique identifier.
func (m *MockConversationRepo) GetByID(ctx context.Context, id string) (*models.Conversation, error) {
	m.Calls.GetByID = append(m.Calls.GetByID, ConversationGetByIDCall{ID: id})
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// CreateMessage persists a new message in a conversation.
func (m *MockConversationRepo) CreateMessage(ctx context.Context, msg *models.Message) (string, error) {
	m.Calls.CreateMessage = append(m.Calls.CreateMessage, ConversationCreateMessageCall{Message: msg})
	if m.CreateMessageFunc != nil {
		return m.CreateMessageFunc(ctx, msg)
	}
	return "", nil
}

// GetMessages retrieves messages for a conversation with pagination.
func (m *MockConversationRepo) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error) {
	m.Calls.GetMessages = append(m.Calls.GetMessages, ConversationGetMessagesCall{
		ConversationID: conversationID,
		Limit:          limit,
		Offset:         offset,
	})
	if m.GetMessagesFunc != nil {
		return m.GetMessagesFunc(ctx, conversationID, limit, offset)
	}
	return nil, nil
}

// IncrementMessageCount atomically increments the message count for a conversation.
func (m *MockConversationRepo) IncrementMessageCount(ctx context.Context, conversationID string) error {
	m.Calls.IncrementMessageCount = append(m.Calls.IncrementMessageCount, ConversationIncrementMessageCountCall{ConversationID: conversationID})
	if m.IncrementMessageCountFunc != nil {
		return m.IncrementMessageCountFunc(ctx, conversationID)
	}
	return nil
}

// ListRecentMessages retrieves recent messages for a conversation.
func (m *MockConversationRepo) ListRecentMessages(ctx context.Context, conversationID string, limit int) ([]models.Message, error) {
	m.Calls.ListRecentMessages = append(m.Calls.ListRecentMessages, ConversationListRecentMessagesCall{ConversationID: conversationID, Limit: limit})
	if m.ListRecentMessagesFunc != nil {
		return m.ListRecentMessagesFunc(ctx, conversationID, limit)
	}
	return nil, nil
}

// SaveMessageSources persists source usage for a message.
func (m *MockConversationRepo) SaveMessageSources(ctx context.Context, messageID string, sources []models.ChunkMetadata) error {
	m.Calls.SaveMessageSources = append(m.Calls.SaveMessageSources, ConversationSaveMessageSourcesCall{MessageID: messageID, Sources: sources})
	if m.SaveMessageSourcesFunc != nil {
		return m.SaveMessageSourcesFunc(ctx, messageID, sources)
	}
	return nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockConversationRepo) Reset() {
	m.Calls.GetOrCreateBySessionID = nil
	m.Calls.GetByID = nil
	m.Calls.CreateMessage = nil
	m.Calls.GetMessages = nil
	m.Calls.IncrementMessageCount = nil
	m.Calls.ListRecentMessages = nil
	m.Calls.SaveMessageSources = nil
}
