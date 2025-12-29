package repository

import (
	"context"
)

// MockAdminChatbotRepo is a mock implementation of AdminChatbotRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockAdminChatbotRepo struct {
	// ListChatbotsFunc is called when ListChatbots is invoked.
	ListChatbotsFunc func(ctx context.Context, filter AdminChatbotFilter, limit, offset int) ([]AdminChatbot, int, error)

	// GetByIDFunc is called when GetByID is invoked.
	GetByIDFunc func(ctx context.Context, id string) (*AdminChatbot, error)

	// ResetSourcesFunc is called when ResetSources is invoked.
	ResetSourcesFunc func(ctx context.Context, chatbotID string) (int64, error)

	// GetSourceIDsFunc is called when GetSourceIDs is invoked.
	GetSourceIDsFunc func(ctx context.Context, chatbotID string) ([]string, error)

	// DeleteVectorsFunc is called when DeleteVectors is invoked.
	DeleteVectorsFunc func(ctx context.Context, chatbotID string) error

	// Invocation tracking for test assertions
	Calls struct {
		ListChatbots  []ListChatbotsCall
		GetByID       []AdminGetByIDCall
		ResetSources  []ResetSourcesCall
		GetSourceIDs  []GetSourceIDsCall
		DeleteVectors []DeleteVectorsCall
	}
}

// Call recording types for test verification
type ListChatbotsCall struct {
	Filter AdminChatbotFilter
	Limit  int
	Offset int
}

type AdminGetByIDCall struct {
	ID string
}

type ResetSourcesCall struct {
	ChatbotID string
}

type GetSourceIDsCall struct {
	ChatbotID string
}

type DeleteVectorsCall struct {
	ChatbotID string
}

// Compile-time check that MockAdminChatbotRepo implements AdminChatbotRepository.
var _ AdminChatbotRepository = (*MockAdminChatbotRepo)(nil)

// NewMockAdminChatbotRepo creates a new MockAdminChatbotRepo with default no-op behavior.
func NewMockAdminChatbotRepo() *MockAdminChatbotRepo {
	return &MockAdminChatbotRepo{}
}

// ListChatbots returns a paginated list of all chatbots with their metadata.
func (m *MockAdminChatbotRepo) ListChatbots(ctx context.Context, filter AdminChatbotFilter, limit, offset int) ([]AdminChatbot, int, error) {
	m.Calls.ListChatbots = append(m.Calls.ListChatbots, ListChatbotsCall{Filter: filter, Limit: limit, Offset: offset})
	if m.ListChatbotsFunc != nil {
		return m.ListChatbotsFunc(ctx, filter, limit, offset)
	}
	return nil, 0, nil
}

// GetByID retrieves a single chatbot by ID with all admin-visible details.
func (m *MockAdminChatbotRepo) GetByID(ctx context.Context, id string) (*AdminChatbot, error) {
	m.Calls.GetByID = append(m.Calls.GetByID, AdminGetByIDCall{ID: id})
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// ResetSources resets all sources for a chatbot to pending status for reprocessing.
func (m *MockAdminChatbotRepo) ResetSources(ctx context.Context, chatbotID string) (int64, error) {
	m.Calls.ResetSources = append(m.Calls.ResetSources, ResetSourcesCall{ChatbotID: chatbotID})
	if m.ResetSourcesFunc != nil {
		return m.ResetSourcesFunc(ctx, chatbotID)
	}
	return 0, nil
}

// GetSourceIDs returns all pending source IDs for a chatbot for queue processing.
func (m *MockAdminChatbotRepo) GetSourceIDs(ctx context.Context, chatbotID string) ([]string, error) {
	m.Calls.GetSourceIDs = append(m.Calls.GetSourceIDs, GetSourceIDsCall{ChatbotID: chatbotID})
	if m.GetSourceIDsFunc != nil {
		return m.GetSourceIDsFunc(ctx, chatbotID)
	}
	return nil, nil
}

// DeleteVectors resets chunk counts to 0 for all sources (for reindexing).
func (m *MockAdminChatbotRepo) DeleteVectors(ctx context.Context, chatbotID string) error {
	m.Calls.DeleteVectors = append(m.Calls.DeleteVectors, DeleteVectorsCall{ChatbotID: chatbotID})
	if m.DeleteVectorsFunc != nil {
		return m.DeleteVectorsFunc(ctx, chatbotID)
	}
	return nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockAdminChatbotRepo) Reset() {
	m.Calls.ListChatbots = nil
	m.Calls.GetByID = nil
	m.Calls.ResetSources = nil
	m.Calls.GetSourceIDs = nil
	m.Calls.DeleteVectors = nil
}
