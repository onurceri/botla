package repository

import (
	"context"
	"time"

	"github.com/onurceri/botla-app/internal/models"
)

// MockSourceRepo is a mock implementation of SourceRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockSourceRepo struct {
	// GetByIDFunc is called when GetByID is invoked.
	GetByIDFunc func(ctx context.Context, id string) (*models.DataSource, error)

	// GetByChatbotFunc is called when GetByChatbot is invoked.
	GetByChatbotFunc func(ctx context.Context, chatbotID string) ([]models.DataSource, error)

	// GetURLSourcesFunc is called when GetURLSources is invoked.
	GetURLSourcesFunc func(ctx context.Context, chatbotID string) ([]models.DataSource, error)

	// CreateFunc is called when Create is invoked.
	CreateFunc func(ctx context.Context, source *models.DataSource) (string, error)

	// SoftDeleteFunc is called when SoftDelete is invoked.
	SoftDeleteFunc func(ctx context.Context, id string) error

	// DeleteFunc is called when Delete is invoked.
	DeleteFunc func(ctx context.Context, id string) error

	// ExistsFunc is called when Exists is invoked.
	ExistsFunc func(ctx context.Context, chatbotID, url string) (bool, error)

	// ExistsByHashFunc is called when ExistsByHash is invoked.
	ExistsByHashFunc func(ctx context.Context, chatbotID, hash string) (bool, error)

	// GetByHashFunc is called when GetByHash is invoked.
	GetByHashFunc func(ctx context.Context, chatbotID, hash string) (*models.DataSource, error)

	// CountByTypeFunc is called when CountByType is invoked.
	CountByTypeFunc func(ctx context.Context, chatbotID, sourceType string) (int, error)

	// UpdateForRefreshFunc is called when UpdateForRefresh is invoked.
	UpdateForRefreshFunc func(ctx context.Context, id string) error

	// Invocation tracking for test assertions
	Calls struct {
		GetByID          []MockSourceGetByIDCall
		GetByChatbot     []MockSourceGetByChatbotCall
		GetURLSources    []MockSourceGetURLSourcesCall
		Create           []MockSourceCreateCall
		SoftDelete       []MockSourceSoftDeleteCall
		Delete           []MockSourceDeleteCall
		Exists           []MockSourceExistsCall
		ExistsByHash     []MockSourceExistsByHashCall
		GetByHash        []MockSourceGetByHashCall
		CountByType      []MockSourceCountByTypeCall
		UpdateForRefresh []MockSourceUpdateForRefreshCall
	}
}

// Call recording types for test verification.
type MockSourceGetByIDCall struct {
	ID string
}

type MockSourceGetByChatbotCall struct {
	ChatbotID string
}

type MockSourceGetURLSourcesCall struct {
	ChatbotID string
}

type MockSourceCreateCall struct {
	Source *models.DataSource
}

type MockSourceSoftDeleteCall struct {
	ID string
}

type MockSourceDeleteCall struct {
	ID string
}

type MockSourceExistsCall struct {
	ChatbotID string
	URL       string
}

type MockSourceExistsByHashCall struct {
	ChatbotID string
	Hash      string
}

type MockSourceGetByHashCall struct {
	ChatbotID string
	Hash      string
}

type MockSourceCountByTypeCall struct {
	ChatbotID  string
	SourceType string
}

type MockSourceUpdateForRefreshCall struct {
	ID string
}

// Compile-time check that MockSourceRepo implements SourceRepository.
var _ SourceRepository = (*MockSourceRepo)(nil)

// NewMockSourceRepo creates a new MockSourceRepo with default no-op behavior.
func NewMockSourceRepo() *MockSourceRepo {
	return &MockSourceRepo{}
}

// GetByID retrieves a data source by its unique identifier.
func (m *MockSourceRepo) GetByID(ctx context.Context, id string) (*models.DataSource, error) {
	m.Calls.GetByID = append(m.Calls.GetByID, MockSourceGetByIDCall{ID: id})
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByChatbot retrieves all non-deleted data sources for a chatbot.
func (m *MockSourceRepo) GetByChatbot(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
	m.Calls.GetByChatbot = append(m.Calls.GetByChatbot, MockSourceGetByChatbotCall{ChatbotID: chatbotID})
	if m.GetByChatbotFunc != nil {
		return m.GetByChatbotFunc(ctx, chatbotID)
	}
	return nil, nil
}

// GetURLSources retrieves all URL-type sources for a chatbot.
func (m *MockSourceRepo) GetURLSources(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
	m.Calls.GetURLSources = append(m.Calls.GetURLSources, MockSourceGetURLSourcesCall{ChatbotID: chatbotID})
	if m.GetURLSourcesFunc != nil {
		return m.GetURLSourcesFunc(ctx, chatbotID)
	}
	return nil, nil
}

// Create persists a new data source and returns its generated ID.
func (m *MockSourceRepo) Create(ctx context.Context, source *models.DataSource) (string, error) {
	m.Calls.Create = append(m.Calls.Create, MockSourceCreateCall{Source: source})
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, source)
	}
	return "", nil
}

// SoftDelete marks a source as deleted by setting deleted_at timestamp.
func (m *MockSourceRepo) SoftDelete(ctx context.Context, id string) error {
	m.Calls.SoftDelete = append(m.Calls.SoftDelete, MockSourceSoftDeleteCall{ID: id})
	if m.SoftDeleteFunc != nil {
		return m.SoftDeleteFunc(ctx, id)
	}
	return nil
}

// Delete permanently removes a source by its ID.
func (m *MockSourceRepo) Delete(ctx context.Context, id string) error {
	m.Calls.Delete = append(m.Calls.Delete, MockSourceDeleteCall{ID: id})
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// Exists checks if a source with the given URL already exists for a chatbot.
func (m *MockSourceRepo) Exists(ctx context.Context, chatbotID, url string) (bool, error) {
	m.Calls.Exists = append(m.Calls.Exists, MockSourceExistsCall{ChatbotID: chatbotID, URL: url})
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, chatbotID, url)
	}
	return false, nil
}

// ExistsByHash checks if a source with the same content hash exists for a chatbot.
func (m *MockSourceRepo) ExistsByHash(ctx context.Context, chatbotID, hash string) (bool, error) {
	m.Calls.ExistsByHash = append(m.Calls.ExistsByHash, MockSourceExistsByHashCall{ChatbotID: chatbotID, Hash: hash})
	if m.ExistsByHashFunc != nil {
		return m.ExistsByHashFunc(ctx, chatbotID, hash)
	}
	return false, nil
}

// GetByHash retrieves a source by its content hash within a chatbot.
func (m *MockSourceRepo) GetByHash(ctx context.Context, chatbotID, hash string) (*models.DataSource, error) {
	m.Calls.GetByHash = append(m.Calls.GetByHash, MockSourceGetByHashCall{ChatbotID: chatbotID, Hash: hash})
	if m.GetByHashFunc != nil {
		return m.GetByHashFunc(ctx, chatbotID, hash)
	}
	return nil, nil
}

// CountByType counts non-deleted, non-failed sources of a specific type.
func (m *MockSourceRepo) CountByType(ctx context.Context, chatbotID, sourceType string) (int, error) {
	m.Calls.CountByType = append(m.Calls.CountByType, MockSourceCountByTypeCall{ChatbotID: chatbotID, SourceType: sourceType})
	if m.CountByTypeFunc != nil {
		return m.CountByTypeFunc(ctx, chatbotID, sourceType)
	}
	return 0, nil
}

// UpdateForRefresh sets status to pending and clears error_message for a source refresh.
func (m *MockSourceRepo) UpdateForRefresh(ctx context.Context, id string) error {
	m.Calls.UpdateForRefresh = append(m.Calls.UpdateForRefresh, MockSourceUpdateForRefreshCall{ID: id})
	if m.UpdateForRefreshFunc != nil {
		return m.UpdateForRefreshFunc(ctx, id)
	}
	return nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockSourceRepo) Reset() {
	m.Calls.GetByID = nil
	m.Calls.GetByChatbot = nil
	m.Calls.GetURLSources = nil
	m.Calls.Create = nil
	m.Calls.SoftDelete = nil
	m.Calls.Delete = nil
	m.Calls.Exists = nil
	m.Calls.ExistsByHash = nil
	m.Calls.GetByHash = nil
	m.Calls.CountByType = nil
	m.Calls.UpdateForRefresh = nil
}

// UpdateSourceHash updates the content hash for a source.
func (m *MockSourceRepo) UpdateSourceHash(ctx context.Context, id string, hash string) error {
	return nil
}

// UpdateSourceProcessing updates processing status, error, chunk count, and processed_at.
func (m *MockSourceRepo) UpdateSourceProcessing(ctx context.Context, id string, status string, errorMessage *string, chunkCount int, processedAt *time.Time) error {
	return nil
}

// UpdateSourceCapability updates the capability summary for a source.
func (m *MockSourceRepo) UpdateSourceCapability(ctx context.Context, id string, summary string) error {
	return nil
}

// UpdateSourceSuggestions updates the suggested questions for a source.
func (m *MockSourceRepo) UpdateSourceSuggestions(ctx context.Context, id string, suggestions []string) error {
	return nil
}

// GetLastDeletedAtForURL returns the most recent deleted_at timestamp for a given URL in a chatbot.
func (m *MockSourceRepo) GetLastDeletedAtForURL(ctx context.Context, chatbotID, url string) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

// GetSourceSuggestions retrieves source suggested questions with chunk counts for aggregation.
func (m *MockSourceRepo) GetSourceSuggestions(ctx context.Context, chatbotID string) ([]SourceSuggestion, error) {
	return nil, nil
}
