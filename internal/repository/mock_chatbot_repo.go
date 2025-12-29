package repository

import (
	"context"

	"github.com/onurceri/botla-co/internal/models"
)

// MockChatbotRepo is a mock implementation of ChatbotRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockChatbotRepo struct {
	// GetByIDFunc is called when GetByID is invoked.
	GetByIDFunc func(ctx context.Context, id string) (*models.Chatbot, error)

	// GetByUserIDFunc is called when GetByUserID is invoked.
	GetByUserIDFunc func(ctx context.Context, userID string) ([]models.Chatbot, error)

	// GetByWorkspaceFunc is called when GetByWorkspace is invoked.
	GetByWorkspaceFunc func(ctx context.Context, workspaceID string) ([]models.Chatbot, error)

	// CreateFunc is called when Create is invoked.
	CreateFunc func(ctx context.Context, bot *models.Chatbot) (string, error)

	// UpdateFunc is called when Update is invoked.
	UpdateFunc func(ctx context.Context, bot *models.Chatbot) error

	// SoftDeleteFunc is called when SoftDelete is invoked.
	SoftDeleteFunc func(ctx context.Context, id, userID string) ([]string, error)

	// CountByUserIDFunc is called when CountByUserID is invoked.
	CountByUserIDFunc func(ctx context.Context, userID string) (int, error)

	// CountByWorkspaceFunc is called when CountByWorkspace is invoked.
	CountByWorkspaceFunc func(ctx context.Context, workspaceID string) (int, error)

	// UpdateSuggestedQuestionsFunc is called when UpdateSuggestedQuestions is invoked.
	UpdateSuggestedQuestionsFunc func(ctx context.Context, id string, suggestions []string) error

	// Invocation tracking for test assertions
	Calls struct {
		GetByID                 []GetByIDCall
		GetByUserID             []GetByUserIDCall
		GetByWorkspace          []GetByWorkspaceCall
		Create                  []CreateCall
		Update                  []UpdateCall
		SoftDelete              []SoftDeleteCall
		CountByUserID           []CountByUserIDCall
		CountByWorkspace        []CountByWorkspaceCall
		UpdateSuggestedQuestions []UpdateSuggestedQuestionsCall
	}
}

// Call recording types for test verification
type GetByIDCall struct {
	ID string
}

type GetByUserIDCall struct {
	UserID string
}

type GetByWorkspaceCall struct {
	WorkspaceID string
}

type CreateCall struct {
	Bot *models.Chatbot
}

type UpdateCall struct {
	Bot *models.Chatbot
}

type SoftDeleteCall struct {
	ID     string
	UserID string
}

type CountByUserIDCall struct {
	UserID string
}

type CountByWorkspaceCall struct {
	WorkspaceID string
}

type UpdateSuggestedQuestionsCall struct {
	ID          string
	Suggestions []string
}

// Compile-time check that MockChatbotRepo implements ChatbotRepository.
var _ ChatbotRepository = (*MockChatbotRepo)(nil)

// NewMockChatbotRepo creates a new MockChatbotRepo with default no-op behavior.
func NewMockChatbotRepo() *MockChatbotRepo {
	return &MockChatbotRepo{}
}

// GetByID retrieves a chatbot by its unique identifier.
func (m *MockChatbotRepo) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
	m.Calls.GetByID = append(m.Calls.GetByID, GetByIDCall{ID: id})
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByUserID retrieves all chatbots for a user.
func (m *MockChatbotRepo) GetByUserID(ctx context.Context, userID string) ([]models.Chatbot, error) {
	m.Calls.GetByUserID = append(m.Calls.GetByUserID, GetByUserIDCall{UserID: userID})
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

// GetByWorkspace retrieves all chatbots for a workspace.
func (m *MockChatbotRepo) GetByWorkspace(ctx context.Context, workspaceID string) ([]models.Chatbot, error) {
	m.Calls.GetByWorkspace = append(m.Calls.GetByWorkspace, GetByWorkspaceCall{WorkspaceID: workspaceID})
	if m.GetByWorkspaceFunc != nil {
		return m.GetByWorkspaceFunc(ctx, workspaceID)
	}
	return nil, nil
}

// Create persists a new chatbot.
func (m *MockChatbotRepo) Create(ctx context.Context, bot *models.Chatbot) (string, error) {
	m.Calls.Create = append(m.Calls.Create, CreateCall{Bot: bot})
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, bot)
	}
	return "", nil
}

// Update modifies an existing chatbot.
func (m *MockChatbotRepo) Update(ctx context.Context, bot *models.Chatbot) error {
	m.Calls.Update = append(m.Calls.Update, UpdateCall{Bot: bot})
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, bot)
	}
	return nil
}

// SoftDelete marks a chatbot as deleted.
func (m *MockChatbotRepo) SoftDelete(ctx context.Context, id, userID string) ([]string, error) {
	m.Calls.SoftDelete = append(m.Calls.SoftDelete, SoftDeleteCall{ID: id, UserID: userID})
	if m.SoftDeleteFunc != nil {
		return m.SoftDeleteFunc(ctx, id, userID)
	}
	return nil, nil
}

// CountByUserID returns the count of chatbots for a user.
func (m *MockChatbotRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	m.Calls.CountByUserID = append(m.Calls.CountByUserID, CountByUserIDCall{UserID: userID})
	if m.CountByUserIDFunc != nil {
		return m.CountByUserIDFunc(ctx, userID)
	}
	return 0, nil
}

// CountByWorkspace returns the count of chatbots for a workspace.
func (m *MockChatbotRepo) CountByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	m.Calls.CountByWorkspace = append(m.Calls.CountByWorkspace, CountByWorkspaceCall{WorkspaceID: workspaceID})
	if m.CountByWorkspaceFunc != nil {
		return m.CountByWorkspaceFunc(ctx, workspaceID)
	}
	return 0, nil
}

// UpdateSuggestedQuestions updates the AI-generated suggestions.
func (m *MockChatbotRepo) UpdateSuggestedQuestions(ctx context.Context, id string, suggestions []string) error {
	m.Calls.UpdateSuggestedQuestions = append(m.Calls.UpdateSuggestedQuestions, UpdateSuggestedQuestionsCall{ID: id, Suggestions: suggestions})
	if m.UpdateSuggestedQuestionsFunc != nil {
		return m.UpdateSuggestedQuestionsFunc(ctx, id, suggestions)
	}
	return nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockChatbotRepo) Reset() {
	m.Calls.GetByID = nil
	m.Calls.GetByUserID = nil
	m.Calls.GetByWorkspace = nil
	m.Calls.Create = nil
	m.Calls.Update = nil
	m.Calls.SoftDelete = nil
	m.Calls.CountByUserID = nil
	m.Calls.CountByWorkspace = nil
	m.Calls.UpdateSuggestedQuestions = nil
}
