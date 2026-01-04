package repository

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

// MockPendingURLRepo is a mock implementation of PendingURLRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockPendingURLRepo struct {
	// InsertPendingURLFunc is called when InsertPendingURL is invoked.
	InsertPendingURLFunc func(ctx context.Context, chatbotID string, sourceID *string, url string) error

	// ListPendingURLsFunc is called when ListPendingURLs is invoked.
	ListPendingURLsFunc func(ctx context.Context, chatbotID string, limit, offset int) ([]models.PendingURL, error)

	// CountPendingURLsFunc is called when CountPendingURLs is invoked.
	CountPendingURLsFunc func(ctx context.Context, chatbotID string) (int, error)

	// UpdatePendingURLStatusFunc is called when UpdatePendingURLStatus is invoked.
	UpdatePendingURLStatusFunc func(ctx context.Context, chatbotID string, urlIDs []string, status string) (int, error)

	// GetPendingURLsByIDsFunc is called when GetPendingURLsByIDs is invoked.
	GetPendingURLsByIDsFunc func(ctx context.Context, chatbotID string, urlIDs []string) ([]models.PendingURL, error)

	// DeletePendingURLsByChatbotFunc is called when DeletePendingURLsByChatbot is invoked.
	DeletePendingURLsByChatbotFunc func(ctx context.Context, chatbotID string) (int, error)

	// Invocation tracking for test assertions
	Calls struct {
		InsertPendingURL         []MockInsertPendingURLCall
		ListPendingURLs          []MockListPendingURLsCall
		CountPendingURLs         []MockCountPendingURLsCall
		UpdatePendingURLStatus   []MockUpdatePendingURLStatusCall
		GetPendingURLsByIDs      []MockGetPendingURLsByIDsCall
		DeletePendingURLsByChatbot []MockDeletePendingURLsByChatbotCall
	}
}

// Call recording types for test verification.
type MockInsertPendingURLCall struct {
	ChatbotID string
	SourceID  *string
	URL       string
}

type MockListPendingURLsCall struct {
	ChatbotID string
	Limit     int
	Offset    int
}

type MockCountPendingURLsCall struct {
	ChatbotID string
}

type MockUpdatePendingURLStatusCall struct {
	ChatbotID string
	URLIDs    []string
	Status    string
}

type MockGetPendingURLsByIDsCall struct {
	ChatbotID string
	URLIDs    []string
}

type MockDeletePendingURLsByChatbotCall struct {
	ChatbotID string
}

// Compile-time check that MockPendingURLRepo implements PendingURLRepository.
var _ PendingURLRepository = (*MockPendingURLRepo)(nil)

// NewMockPendingURLRepo creates a new MockPendingURLRepo with default no-op behavior.
func NewMockPendingURLRepo() *MockPendingURLRepo {
	return &MockPendingURLRepo{}
}

// InsertPendingURL adds a URL to the pending list for approval.
func (m *MockPendingURLRepo) InsertPendingURL(ctx context.Context, chatbotID string, sourceID *string, url string) error {
	m.Calls.InsertPendingURL = append(m.Calls.InsertPendingURL, MockInsertPendingURLCall{
		ChatbotID: chatbotID,
		SourceID:  sourceID,
		URL:       url,
	})
	if m.InsertPendingURLFunc != nil {
		return m.InsertPendingURLFunc(ctx, chatbotID, sourceID, url)
	}
	return nil
}

// ListPendingURLs returns pending URLs for a chatbot with pagination.
func (m *MockPendingURLRepo) ListPendingURLs(ctx context.Context, chatbotID string, limit, offset int) ([]models.PendingURL, error) {
	m.Calls.ListPendingURLs = append(m.Calls.ListPendingURLs, MockListPendingURLsCall{
		ChatbotID: chatbotID,
		Limit:     limit,
		Offset:    offset,
	})
	if m.ListPendingURLsFunc != nil {
		return m.ListPendingURLsFunc(ctx, chatbotID, limit, offset)
	}
	return nil, nil
}

// CountPendingURLs returns the total count of pending URLs for a chatbot.
func (m *MockPendingURLRepo) CountPendingURLs(ctx context.Context, chatbotID string) (int, error) {
	m.Calls.CountPendingURLs = append(m.Calls.CountPendingURLs, MockCountPendingURLsCall{ChatbotID: chatbotID})
	if m.CountPendingURLsFunc != nil {
		return m.CountPendingURLsFunc(ctx, chatbotID)
	}
	return 0, nil
}

// UpdatePendingURLStatus updates the status of multiple pending URLs.
func (m *MockPendingURLRepo) UpdatePendingURLStatus(ctx context.Context, chatbotID string, urlIDs []string, status string) (int, error) {
	m.Calls.UpdatePendingURLStatus = append(m.Calls.UpdatePendingURLStatus, MockUpdatePendingURLStatusCall{
		ChatbotID: chatbotID,
		URLIDs:    urlIDs,
		Status:    status,
	})
	if m.UpdatePendingURLStatusFunc != nil {
		return m.UpdatePendingURLStatusFunc(ctx, chatbotID, urlIDs, status)
	}
	return len(urlIDs), nil
}

// GetPendingURLsByIDs returns pending URLs by their IDs.
func (m *MockPendingURLRepo) GetPendingURLsByIDs(ctx context.Context, chatbotID string, urlIDs []string) ([]models.PendingURL, error) {
	m.Calls.GetPendingURLsByIDs = append(m.Calls.GetPendingURLsByIDs, MockGetPendingURLsByIDsCall{
		ChatbotID: chatbotID,
		URLIDs:    urlIDs,
	})
	if m.GetPendingURLsByIDsFunc != nil {
		return m.GetPendingURLsByIDsFunc(ctx, chatbotID, urlIDs)
	}
	return nil, nil
}

// DeletePendingURLsByChatbot clears all pending URLs for a chatbot.
func (m *MockPendingURLRepo) DeletePendingURLsByChatbot(ctx context.Context, chatbotID string) (int, error) {
	m.Calls.DeletePendingURLsByChatbot = append(m.Calls.DeletePendingURLsByChatbot, MockDeletePendingURLsByChatbotCall{ChatbotID: chatbotID})
	if m.DeletePendingURLsByChatbotFunc != nil {
		return m.DeletePendingURLsByChatbotFunc(ctx, chatbotID)
	}
	return 0, nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockPendingURLRepo) Reset() {
	m.Calls.InsertPendingURL = nil
	m.Calls.ListPendingURLs = nil
	m.Calls.CountPendingURLs = nil
	m.Calls.UpdatePendingURLStatus = nil
	m.Calls.GetPendingURLsByIDs = nil
	m.Calls.DeletePendingURLsByChatbot = nil
}
