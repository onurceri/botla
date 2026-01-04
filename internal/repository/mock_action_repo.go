package repository

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

// MockActionRepo is a mock implementation of ActionRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockActionRepo struct {
	// ListFunc is called when List is invoked.
	ListFunc func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)

	// ListEnabledFunc is called when ListEnabled is invoked.
	ListEnabledFunc func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)

	// GetByIDFunc is called when GetByID is invoked.
	GetByIDFunc func(ctx context.Context, id string) (*models.ChatbotAction, error)

	// GetByToolNameFunc is called when GetByToolName is invoked.
	GetByToolNameFunc func(ctx context.Context, chatbotID, toolName string) (*models.ChatbotAction, error)

	// CreateFunc is called when Create is invoked.
	CreateFunc func(ctx context.Context, action *models.ChatbotAction) error

	// UpdateFunc is called when Update is invoked.
	UpdateFunc func(ctx context.Context, action *models.ChatbotAction) error

	// DeleteFunc is called when Delete is invoked.
	DeleteFunc func(ctx context.Context, id string) error

	// GetLogsFunc is called when GetLogs is invoked.
	GetLogsFunc func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error)

	// CreateLogFunc is called when CreateLog is invoked.
	CreateLogFunc func(ctx context.Context, log *models.ActionExecutionLog) error

	// Invocation tracking for test assertions
	Calls struct {
		List          []ActionListCall
		ListEnabled   []ActionListEnabledCall
		GetByID       []ActionGetByIDCall
		GetByToolName []ActionGetByToolNameCall
		Create        []ActionCreateCall
		Update        []ActionUpdateCall
		Delete        []ActionDeleteCall
		GetLogs       []ActionGetLogsCall
		CreateLog     []ActionCreateLogCall
	}
}

// Call recording types for test verification
type ActionListCall struct {
	ChatbotID string
}

type ActionListEnabledCall struct {
	ChatbotID string
}

type ActionGetByIDCall struct {
	ID string
}

type ActionGetByToolNameCall struct {
	ChatbotID string
	ToolName  string
}

type ActionCreateCall struct {
	Action *models.ChatbotAction
}

type ActionUpdateCall struct {
	Action *models.ChatbotAction
}

type ActionDeleteCall struct {
	ID string
}

type ActionGetLogsCall struct {
	ChatbotID string
	Limit     int
	Offset    int
}

type ActionCreateLogCall struct {
	Log *models.ActionExecutionLog
}

// Compile-time check that MockActionRepo implements ActionRepository.
var _ ActionRepository = (*MockActionRepo)(nil)

// NewMockActionRepo creates a new MockActionRepo with default no-op behavior.
func NewMockActionRepo() *MockActionRepo {
	return &MockActionRepo{}
}

// List returns all actions for a chatbot.
func (m *MockActionRepo) List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	m.Calls.List = append(m.Calls.List, ActionListCall{ChatbotID: chatbotID})
	if m.ListFunc != nil {
		return m.ListFunc(ctx, chatbotID)
	}
	return nil, nil
}

// ListEnabled returns only enabled actions for a chatbot.
func (m *MockActionRepo) ListEnabled(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	m.Calls.ListEnabled = append(m.Calls.ListEnabled, ActionListEnabledCall{ChatbotID: chatbotID})
	if m.ListEnabledFunc != nil {
		return m.ListEnabledFunc(ctx, chatbotID)
	}
	return nil, nil
}

// GetByID retrieves an action by its unique identifier.
func (m *MockActionRepo) GetByID(ctx context.Context, id string) (*models.ChatbotAction, error) {
	m.Calls.GetByID = append(m.Calls.GetByID, ActionGetByIDCall{ID: id})
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByToolName finds an action by its tool_name within a chatbot.
func (m *MockActionRepo) GetByToolName(ctx context.Context, chatbotID, toolName string) (*models.ChatbotAction, error) {
	m.Calls.GetByToolName = append(m.Calls.GetByToolName, ActionGetByToolNameCall{ChatbotID: chatbotID, ToolName: toolName})
	if m.GetByToolNameFunc != nil {
		return m.GetByToolNameFunc(ctx, chatbotID, toolName)
	}
	return nil, nil
}

// Create persists a new action.
func (m *MockActionRepo) Create(ctx context.Context, action *models.ChatbotAction) error {
	m.Calls.Create = append(m.Calls.Create, ActionCreateCall{Action: action})
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, action)
	}
	return nil
}

// Update modifies an existing action.
func (m *MockActionRepo) Update(ctx context.Context, action *models.ChatbotAction) error {
	m.Calls.Update = append(m.Calls.Update, ActionUpdateCall{Action: action})
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, action)
	}
	return nil
}

// Delete permanently removes an action by its ID.
func (m *MockActionRepo) Delete(ctx context.Context, id string) error {
	m.Calls.Delete = append(m.Calls.Delete, ActionDeleteCall{ID: id})
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// GetLogs retrieves action execution logs for a chatbot with pagination.
func (m *MockActionRepo) GetLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	m.Calls.GetLogs = append(m.Calls.GetLogs, ActionGetLogsCall{ChatbotID: chatbotID, Limit: limit, Offset: offset})
	if m.GetLogsFunc != nil {
		return m.GetLogsFunc(ctx, chatbotID, limit, offset)
	}
	return nil, nil
}

// CreateLog persists an action execution log entry.
func (m *MockActionRepo) CreateLog(ctx context.Context, log *models.ActionExecutionLog) error {
	m.Calls.CreateLog = append(m.Calls.CreateLog, ActionCreateLogCall{Log: log})
	if m.CreateLogFunc != nil {
		return m.CreateLogFunc(ctx, log)
	}
	return nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockActionRepo) Reset() {
	m.Calls.List = nil
	m.Calls.ListEnabled = nil
	m.Calls.GetByID = nil
	m.Calls.GetByToolName = nil
	m.Calls.Create = nil
	m.Calls.Update = nil
	m.Calls.Delete = nil
	m.Calls.GetLogs = nil
	m.Calls.CreateLog = nil
}
