package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

// MockChatbotRepository is a test double for ChatbotRepository.
// It can be configured with custom behaviors for each method.
type MockChatbotRepository struct {
	GetByIDFunc                  func(ctx context.Context, id string) (*models.Chatbot, error)
	GetByUserIDFunc              func(ctx context.Context, userID string) ([]models.Chatbot, error)
	GetByWorkspaceFunc           func(ctx context.Context, workspaceID string) ([]models.Chatbot, error)
	CreateFunc                   func(ctx context.Context, bot *models.Chatbot) (string, error)
	UpdateFunc                   func(ctx context.Context, bot *models.Chatbot) error
	SoftDeleteFunc               func(ctx context.Context, id, userID string) ([]string, error)
	CountByUserIDFunc            func(ctx context.Context, userID string) (int, error)
	CountByWorkspaceFunc         func(ctx context.Context, workspaceID string) (int, error)
	UpdateSuggestedQuestionsFunc func(ctx context.Context, id string, suggestions []string) error
	GetDueForRefreshFunc         func(ctx context.Context, now time.Time) ([]models.Chatbot, error)
	UpdateRefreshTimesFunc       func(ctx context.Context, botID string, nextRefresh, lastRefresh time.Time) error

	// Invocation tracking
	GetByIDCalls                  []string
	GetByUserIDCalls              []string
	GetByWorkspaceCalls           []string
	CreateCalls                   []*models.Chatbot
	UpdateCalls                   []*models.Chatbot
	SoftDeleteCalls               []struct{ ID, UserID string }
	CountByUserIDCalls            []string
	CountByWorkspaceCalls         []string
	UpdateSuggestedQuestionsCalls []struct {
		ID          string
		Suggestions []string
	}
	GetDueForRefreshCalls   []time.Time
	UpdateRefreshTimesCalls []struct {
		BotID       string
		NextRefresh time.Time
		LastRefresh time.Time
	}
}

var _ repository.ChatbotRepository = (*MockChatbotRepository)(nil)

func (m *MockChatbotRepository) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
	m.GetByIDCalls = append(m.GetByIDCalls, id)
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockChatbotRepository) GetByUserID(ctx context.Context, userID string) ([]models.Chatbot, error) {
	m.GetByUserIDCalls = append(m.GetByUserIDCalls, userID)
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockChatbotRepository) GetByWorkspace(ctx context.Context, workspaceID string) ([]models.Chatbot, error) {
	m.GetByWorkspaceCalls = append(m.GetByWorkspaceCalls, workspaceID)
	if m.GetByWorkspaceFunc != nil {
		return m.GetByWorkspaceFunc(ctx, workspaceID)
	}
	return nil, nil
}

func (m *MockChatbotRepository) Create(ctx context.Context, bot *models.Chatbot) (string, error) {
	m.CreateCalls = append(m.CreateCalls, bot)
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, bot)
	}
	return "mock-id", nil
}

func (m *MockChatbotRepository) Update(ctx context.Context, bot *models.Chatbot) error {
	m.UpdateCalls = append(m.UpdateCalls, bot)
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, bot)
	}
	return nil
}

func (m *MockChatbotRepository) SoftDelete(ctx context.Context, id, userID string) ([]string, error) {
	m.SoftDeleteCalls = append(m.SoftDeleteCalls, struct{ ID, UserID string }{id, userID})
	if m.SoftDeleteFunc != nil {
		return m.SoftDeleteFunc(ctx, id, userID)
	}
	return nil, nil
}

func (m *MockChatbotRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	m.CountByUserIDCalls = append(m.CountByUserIDCalls, userID)
	if m.CountByUserIDFunc != nil {
		return m.CountByUserIDFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockChatbotRepository) CountByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	m.CountByWorkspaceCalls = append(m.CountByWorkspaceCalls, workspaceID)
	if m.CountByWorkspaceFunc != nil {
		return m.CountByWorkspaceFunc(ctx, workspaceID)
	}
	return 0, nil
}

func (m *MockChatbotRepository) UpdateSuggestedQuestions(ctx context.Context, id string, suggestions []string) error {
	m.UpdateSuggestedQuestionsCalls = append(m.UpdateSuggestedQuestionsCalls, struct {
		ID          string
		Suggestions []string
	}{id, suggestions})
	if m.UpdateSuggestedQuestionsFunc != nil {
		return m.UpdateSuggestedQuestionsFunc(ctx, id, suggestions)
	}
	return nil
}

func (m *MockChatbotRepository) GetDueForRefresh(ctx context.Context, now time.Time) ([]models.Chatbot, error) {
	m.GetDueForRefreshCalls = append(m.GetDueForRefreshCalls, now)
	if m.GetDueForRefreshFunc != nil {
		return m.GetDueForRefreshFunc(ctx, now)
	}
	return nil, nil
}

func (m *MockChatbotRepository) UpdateRefreshTimes(ctx context.Context, botID string, nextRefresh, lastRefresh time.Time) error {
	m.UpdateRefreshTimesCalls = append(m.UpdateRefreshTimesCalls, struct {
		BotID       string
		NextRefresh time.Time
		LastRefresh time.Time
	}{botID, nextRefresh, lastRefresh})
	if m.UpdateRefreshTimesFunc != nil {
		return m.UpdateRefreshTimesFunc(ctx, botID, nextRefresh, lastRefresh)
	}
	return nil
}

// MockActionRepository is a test double for ActionRepository.
type MockActionRepository struct {
	ListFunc          func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)
	ListEnabledFunc   func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)
	GetByIDFunc       func(ctx context.Context, id string) (*models.ChatbotAction, error)
	GetByToolNameFunc func(ctx context.Context, chatbotID, toolName string) (*models.ChatbotAction, error)
	CreateFunc        func(ctx context.Context, action *models.ChatbotAction) error
	UpdateFunc        func(ctx context.Context, action *models.ChatbotAction) error
	DeleteFunc        func(ctx context.Context, id string) error
	GetLogsFunc       func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error)
	CreateLogFunc     func(ctx context.Context, log *models.ActionExecutionLog) error

	// Invocation tracking
	ListCalls          []string
	ListEnabledCalls   []string
	GetByIDCalls       []string
	GetByToolNameCalls []struct{ ChatbotID, ToolName string }
	CreateCalls        []*models.ChatbotAction
	UpdateCalls        []*models.ChatbotAction
	DeleteCalls        []string
	GetLogsCalls       []struct {
		ChatbotID     string
		Limit, Offset int
	}
	CreateLogCalls []*models.ActionExecutionLog
}

var _ repository.ActionRepository = (*MockActionRepository)(nil)

func (m *MockActionRepository) List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	m.ListCalls = append(m.ListCalls, chatbotID)
	if m.ListFunc != nil {
		return m.ListFunc(ctx, chatbotID)
	}
	return nil, nil
}

func (m *MockActionRepository) ListEnabled(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	m.ListEnabledCalls = append(m.ListEnabledCalls, chatbotID)
	if m.ListEnabledFunc != nil {
		return m.ListEnabledFunc(ctx, chatbotID)
	}
	return nil, nil
}

func (m *MockActionRepository) GetByID(ctx context.Context, id string) (*models.ChatbotAction, error) {
	m.GetByIDCalls = append(m.GetByIDCalls, id)
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockActionRepository) GetByToolName(ctx context.Context, chatbotID, toolName string) (*models.ChatbotAction, error) {
	m.GetByToolNameCalls = append(m.GetByToolNameCalls, struct{ ChatbotID, ToolName string }{chatbotID, toolName})
	if m.GetByToolNameFunc != nil {
		return m.GetByToolNameFunc(ctx, chatbotID, toolName)
	}
	return nil, nil
}

func (m *MockActionRepository) Create(ctx context.Context, action *models.ChatbotAction) error {
	m.CreateCalls = append(m.CreateCalls, action)
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, action)
	}
	return nil
}

func (m *MockActionRepository) Update(ctx context.Context, action *models.ChatbotAction) error {
	m.UpdateCalls = append(m.UpdateCalls, action)
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, action)
	}
	return nil
}

func (m *MockActionRepository) Delete(ctx context.Context, id string) error {
	m.DeleteCalls = append(m.DeleteCalls, id)
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockActionRepository) GetLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	m.GetLogsCalls = append(m.GetLogsCalls, struct {
		ChatbotID     string
		Limit, Offset int
	}{chatbotID, limit, offset})
	if m.GetLogsFunc != nil {
		return m.GetLogsFunc(ctx, chatbotID, limit, offset)
	}
	return nil, nil
}

func (m *MockActionRepository) CreateLog(ctx context.Context, log *models.ActionExecutionLog) error {
	m.CreateLogCalls = append(m.CreateLogCalls, log)
	if m.CreateLogFunc != nil {
		return m.CreateLogFunc(ctx, log)
	}
	return nil
}

// MockSourceRepository is a test double for SourceRepository.
type MockSourceRepository struct {
	GetByIDFunc          func(ctx context.Context, id string) (*models.DataSource, error)
	GetByChatbotFunc     func(ctx context.Context, chatbotID string) ([]models.DataSource, error)
	GetURLSourcesFunc    func(ctx context.Context, chatbotID string) ([]models.DataSource, error)
	CreateFunc           func(ctx context.Context, source *models.DataSource) (string, error)
	SoftDeleteFunc       func(ctx context.Context, id string) error
	DeleteFunc           func(ctx context.Context, id string) error
	ExistsFunc           func(ctx context.Context, chatbotID, url string) (bool, error)
	ExistsByHashFunc     func(ctx context.Context, chatbotID, hash string) (bool, error)
	GetByHashFunc        func(ctx context.Context, chatbotID, hash string) (*models.DataSource, error)
	CountByTypeFunc      func(ctx context.Context, chatbotID, sourceType string) (int, error)
	UpdateForRefreshFunc func(ctx context.Context, id string) error

	// Invocation tracking
	GetByIDCalls          []string
	GetByChatbotCalls     []string
	GetURLSourcesCalls    []string
	CreateCalls           []*models.DataSource
	SoftDeleteCalls       []string
	DeleteCalls           []string
	ExistsCalls           []struct{ ChatbotID, URL string }
	ExistsByHashCalls     []struct{ ChatbotID, Hash string }
	GetByHashCalls        []struct{ ChatbotID, Hash string }
	CountByTypeCalls      []struct{ ChatbotID, SourceType string }
	UpdateForRefreshCalls []string
}

var _ repository.SourceRepository = (*MockSourceRepository)(nil)

func (m *MockSourceRepository) GetByID(ctx context.Context, id string) (*models.DataSource, error) {
	m.GetByIDCalls = append(m.GetByIDCalls, id)
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSourceRepository) GetByChatbot(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
	m.GetByChatbotCalls = append(m.GetByChatbotCalls, chatbotID)
	if m.GetByChatbotFunc != nil {
		return m.GetByChatbotFunc(ctx, chatbotID)
	}
	return nil, nil
}

func (m *MockSourceRepository) GetURLSources(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
	m.GetURLSourcesCalls = append(m.GetURLSourcesCalls, chatbotID)
	if m.GetURLSourcesFunc != nil {
		return m.GetURLSourcesFunc(ctx, chatbotID)
	}
	return nil, nil
}

func (m *MockSourceRepository) Create(ctx context.Context, source *models.DataSource) (string, error) {
	m.CreateCalls = append(m.CreateCalls, source)
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, source)
	}
	return "mock-source-id", nil
}

func (m *MockSourceRepository) SoftDelete(ctx context.Context, id string) error {
	m.SoftDeleteCalls = append(m.SoftDeleteCalls, id)
	if m.SoftDeleteFunc != nil {
		return m.SoftDeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSourceRepository) Delete(ctx context.Context, id string) error {
	m.DeleteCalls = append(m.DeleteCalls, id)
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSourceRepository) Exists(ctx context.Context, chatbotID, url string) (bool, error) {
	m.ExistsCalls = append(m.ExistsCalls, struct{ ChatbotID, URL string }{chatbotID, url})
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, chatbotID, url)
	}
	return false, nil
}

func (m *MockSourceRepository) ExistsByHash(ctx context.Context, chatbotID, hash string) (bool, error) {
	m.ExistsByHashCalls = append(m.ExistsByHashCalls, struct{ ChatbotID, Hash string }{chatbotID, hash})
	if m.ExistsByHashFunc != nil {
		return m.ExistsByHashFunc(ctx, chatbotID, hash)
	}
	return false, nil
}

func (m *MockSourceRepository) GetByHash(ctx context.Context, chatbotID, hash string) (*models.DataSource, error) {
	m.GetByHashCalls = append(m.GetByHashCalls, struct{ ChatbotID, Hash string }{chatbotID, hash})
	if m.GetByHashFunc != nil {
		return m.GetByHashFunc(ctx, chatbotID, hash)
	}
	return nil, nil
}

func (m *MockSourceRepository) CountByType(ctx context.Context, chatbotID, sourceType string) (int, error) {
	m.CountByTypeCalls = append(m.CountByTypeCalls, struct{ ChatbotID, SourceType string }{chatbotID, sourceType})
	if m.CountByTypeFunc != nil {
		return m.CountByTypeFunc(ctx, chatbotID, sourceType)
	}
	return 0, nil
}

func (m *MockSourceRepository) UpdateForRefresh(ctx context.Context, id string) error {
	m.UpdateForRefreshCalls = append(m.UpdateForRefreshCalls, id)
	if m.UpdateForRefreshFunc != nil {
		return m.UpdateForRefreshFunc(ctx, id)
	}
	return nil
}

func (m *MockSourceRepository) UpdateSourceHash(ctx context.Context, id string, hash string) error {
	return nil
}

func (m *MockSourceRepository) UpdateSourceProcessing(ctx context.Context, id string, status string, errorMessage *string, chunkCount int, processedAt *time.Time) error {
	return nil
}

func (m *MockSourceRepository) UpdateSourceCapability(ctx context.Context, id string, summary string) error {
	return nil
}

func (m *MockSourceRepository) UpdateSourceSuggestions(ctx context.Context, id string, suggestions []string) error {
	return nil
}

func (m *MockSourceRepository) GetLastDeletedAtForURL(ctx context.Context, chatbotID, url string) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

func (m *MockSourceRepository) GetSourceSuggestions(ctx context.Context, chatbotID string) ([]repository.SourceSuggestion, error) {
	return nil, nil
}

func (m *MockSourceRepository) GetCapabilitySummaries(ctx context.Context, chatbotID string) ([]string, error) {
	return nil, nil
}

// =============================================================================
// ChatbotRepository Interface Tests
// =============================================================================

func TestChatbotRepository_InterfaceCompliance(t *testing.T) {
	// Verify that our mock implements the interface
	var _ repository.ChatbotRepository = (*MockChatbotRepository)(nil)
}

func TestChatbotRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(m *MockChatbotRepository)
		id      string
		want    *models.Chatbot
		wantErr bool
	}{
		{
			name: "found",
			setupFn: func(m *MockChatbotRepository) {
				m.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
					return &models.Chatbot{ID: id, Name: "Test Bot"}, nil
				}
			},
			id:   "test-id",
			want: &models.Chatbot{ID: "test-id", Name: "Test Bot"},
		},
		{
			name: "not found returns nil",
			setupFn: func(m *MockChatbotRepository) {
				m.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
					return nil, nil
				}
			},
			id:   "missing-id",
			want: nil,
		},
		{
			name: "error propagates",
			setupFn: func(m *MockChatbotRepository) {
				m.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
					return nil, errors.New("database error")
				}
			},
			id:      "error-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockChatbotRepository{}
			tt.setupFn(mock)

			got, err := mock.GetByID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil && tt.want != nil {
					t.Errorf("GetByID() got nil, want %v", tt.want)
				}
				if got != nil && tt.want != nil && got.ID != tt.want.ID {
					t.Errorf("GetByID() ID = %v, want %v", got.ID, tt.want.ID)
				}
			}
			// Verify call was tracked
			if len(mock.GetByIDCalls) != 1 || mock.GetByIDCalls[0] != tt.id {
				t.Errorf("GetByID() call not tracked correctly")
			}
		})
	}
}

func TestChatbotRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(m *MockChatbotRepository)
		userID  string
		want    int
		wantErr bool
	}{
		{
			name: "returns chatbots",
			setupFn: func(m *MockChatbotRepository) {
				m.GetByUserIDFunc = func(ctx context.Context, userID string) ([]models.Chatbot, error) {
					return []models.Chatbot{
						{ID: "bot-1", UserID: userID},
						{ID: "bot-2", UserID: userID},
					}, nil
				}
			},
			userID: "user-1",
			want:   2,
		},
		{
			name: "empty result",
			setupFn: func(m *MockChatbotRepository) {
				m.GetByUserIDFunc = func(ctx context.Context, userID string) ([]models.Chatbot, error) {
					return []models.Chatbot{}, nil
				}
			},
			userID: "new-user",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockChatbotRepository{}
			tt.setupFn(mock)

			got, err := mock.GetByUserID(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("GetByUserID() len = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestChatbotRepository_Create(t *testing.T) {
	mock := &MockChatbotRepository{}
	mock.CreateFunc = func(ctx context.Context, bot *models.Chatbot) (string, error) {
		return "new-bot-id", nil
	}

	bot := &models.Chatbot{Name: "New Bot", UserID: "user-1"}
	id, err := mock.Create(context.Background(), bot)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if id != "new-bot-id" {
		t.Errorf("Create() id = %v, want new-bot-id", id)
	}
	if len(mock.CreateCalls) != 1 {
		t.Errorf("Create() should have been called once")
	}
}

func TestChatbotRepository_SoftDelete(t *testing.T) {
	mock := &MockChatbotRepository{}
	mock.SoftDeleteFunc = func(ctx context.Context, id, userID string) ([]string, error) {
		return []string{"source-1", "source-2"}, nil
	}

	sourceIDs, err := mock.SoftDelete(context.Background(), "bot-id", "user-id")
	if err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	if len(sourceIDs) != 2 {
		t.Errorf("SoftDelete() returned %d source IDs, want 2", len(sourceIDs))
	}
}

func TestChatbotRepository_CountByUserID(t *testing.T) {
	mock := &MockChatbotRepository{}
	mock.CountByUserIDFunc = func(ctx context.Context, userID string) (int, error) {
		return 5, nil
	}

	count, err := mock.CountByUserID(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("CountByUserID() error = %v", err)
	}
	if count != 5 {
		t.Errorf("CountByUserID() = %v, want 5", count)
	}
}

// =============================================================================
// ActionRepository Interface Tests
// =============================================================================

func TestActionRepository_InterfaceCompliance(t *testing.T) {
	var _ repository.ActionRepository = (*MockActionRepository)(nil)
}

func TestActionRepository_List(t *testing.T) {
	mock := &MockActionRepository{}
	mock.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		return []*models.ChatbotAction{
			{ID: "action-1", ChatbotID: chatbotID, Name: "Action 1"},
			{ID: "action-2", ChatbotID: chatbotID, Name: "Action 2"},
		}, nil
	}

	actions, err := mock.List(context.Background(), "bot-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(actions) != 2 {
		t.Errorf("List() returned %d actions, want 2", len(actions))
	}
	if len(mock.ListCalls) != 1 {
		t.Errorf("List() should have been called once")
	}
}

func TestActionRepository_ListEnabled(t *testing.T) {
	mock := &MockActionRepository{}
	mock.ListEnabledFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		return []*models.ChatbotAction{
			{ID: "action-1", ChatbotID: chatbotID, Enabled: true},
		}, nil
	}

	actions, err := mock.ListEnabled(context.Background(), "bot-1")
	if err != nil {
		t.Fatalf("ListEnabled() error = %v", err)
	}
	if len(actions) != 1 {
		t.Errorf("ListEnabled() returned %d actions, want 1", len(actions))
	}
}

func TestActionRepository_UpdateWithVersionConflict(t *testing.T) {
	mock := &MockActionRepository{}
	mock.UpdateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
		return repository.ErrVersionConflict
	}

	action := &models.ChatbotAction{ID: "action-1", Version: 1}
	err := mock.Update(context.Background(), action)
	if !errors.Is(err, repository.ErrVersionConflict) {
		t.Errorf("Update() error = %v, want ErrVersionConflict", err)
	}
}

func TestActionRepository_GetByToolName(t *testing.T) {
	mock := &MockActionRepository{}
	toolName := "send_email"
	mock.GetByToolNameFunc = func(ctx context.Context, chatbotID, tn string) (*models.ChatbotAction, error) {
		if tn == toolName {
			return &models.ChatbotAction{ID: "action-1", ToolName: &toolName}, nil
		}
		return nil, nil
	}

	action, err := mock.GetByToolName(context.Background(), "bot-1", toolName)
	if err != nil {
		t.Fatalf("GetByToolName() error = %v", err)
	}
	if action == nil || *action.ToolName != toolName {
		t.Errorf("GetByToolName() did not find the expected action")
	}
}

func TestActionRepository_CreateLog(t *testing.T) {
	mock := &MockActionRepository{}
	mock.CreateLogFunc = func(ctx context.Context, log *models.ActionExecutionLog) error {
		log.ID = "log-123"
		log.CreatedAt = time.Now()
		return nil
	}

	log := &models.ActionExecutionLog{
		ChatbotID: "bot-1",
		ActionID:  "action-1",
		Status:    "success",
	}
	err := mock.CreateLog(context.Background(), log)
	if err != nil {
		t.Fatalf("CreateLog() error = %v", err)
	}
	if log.ID != "log-123" {
		t.Errorf("CreateLog() did not populate ID")
	}
}

func TestActionRepository_GetLogs(t *testing.T) {
	mock := &MockActionRepository{}
	mock.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
		return []*models.ActionExecutionLog{
			{ID: "log-1", ChatbotID: chatbotID},
			{ID: "log-2", ChatbotID: chatbotID},
		}, nil
	}

	logs, err := mock.GetLogs(context.Background(), "bot-1", 10, 0)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("GetLogs() returned %d logs, want 2", len(logs))
	}
	if len(mock.GetLogsCalls) != 1 {
		t.Errorf("GetLogs() should have been called once")
	}
	if mock.GetLogsCalls[0].Limit != 10 || mock.GetLogsCalls[0].Offset != 0 {
		t.Errorf("GetLogs() pagination params not tracked correctly")
	}
}

// =============================================================================
// SourceRepository Interface Tests
// =============================================================================

func TestSourceRepository_InterfaceCompliance(t *testing.T) {
	var _ repository.SourceRepository = (*MockSourceRepository)(nil)
}

func TestSourceRepository_GetByID(t *testing.T) {
	mock := &MockSourceRepository{}
	mock.GetByIDFunc = func(ctx context.Context, id string) (*models.DataSource, error) {
		return &models.DataSource{ID: id, SourceType: "url"}, nil
	}

	source, err := mock.GetByID(context.Background(), "source-1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if source == nil || source.ID != "source-1" {
		t.Errorf("GetByID() returned unexpected source")
	}
}

func TestSourceRepository_GetByChatbot(t *testing.T) {
	mock := &MockSourceRepository{}
	mock.GetByChatbotFunc = func(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
		return []models.DataSource{
			{ID: "s1", ChatbotID: chatbotID, SourceType: "url"},
			{ID: "s2", ChatbotID: chatbotID, SourceType: "file"},
		}, nil
	}

	sources, err := mock.GetByChatbot(context.Background(), "bot-1")
	if err != nil {
		t.Fatalf("GetByChatbot() error = %v", err)
	}
	if len(sources) != 2 {
		t.Errorf("GetByChatbot() returned %d sources, want 2", len(sources))
	}
}

func TestSourceRepository_Exists(t *testing.T) {
	mock := &MockSourceRepository{}
	mock.ExistsFunc = func(ctx context.Context, chatbotID, url string) (bool, error) {
		return url == "https://existing.com", nil
	}

	tests := []struct {
		url  string
		want bool
	}{
		{"https://existing.com", true},
		{"https://new.com", false},
	}

	for _, tt := range tests {
		exists, err := mock.Exists(context.Background(), "bot-1", tt.url)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if exists != tt.want {
			t.Errorf("Exists(%s) = %v, want %v", tt.url, exists, tt.want)
		}
	}
}

func TestSourceRepository_ExistsByHash(t *testing.T) {
	mock := &MockSourceRepository{}
	existingHash := "abc123"
	mock.ExistsByHashFunc = func(ctx context.Context, chatbotID, hash string) (bool, error) {
		return hash == existingHash, nil
	}

	exists, err := mock.ExistsByHash(context.Background(), "bot-1", existingHash)
	if err != nil {
		t.Fatalf("ExistsByHash() error = %v", err)
	}
	if !exists {
		t.Error("ExistsByHash() should have found existing hash")
	}

	exists, err = mock.ExistsByHash(context.Background(), "bot-1", "different-hash")
	if err != nil {
		t.Fatalf("ExistsByHash() error = %v", err)
	}
	if exists {
		t.Error("ExistsByHash() should not find different hash")
	}
}

func TestSourceRepository_CountByType(t *testing.T) {
	mock := &MockSourceRepository{}
	mock.CountByTypeFunc = func(ctx context.Context, chatbotID, sourceType string) (int, error) {
		if sourceType == "url" {
			return 5, nil
		}
		return 2, nil
	}

	count, err := mock.CountByType(context.Background(), "bot-1", "url")
	if err != nil {
		t.Fatalf("CountByType() error = %v", err)
	}
	if count != 5 {
		t.Errorf("CountByType(url) = %v, want 5", count)
	}

	count, _ = mock.CountByType(context.Background(), "bot-1", "file")
	if count != 2 {
		t.Errorf("CountByType(file) = %v, want 2", count)
	}
}

func TestSourceRepository_SoftDelete(t *testing.T) {
	mock := &MockSourceRepository{}
	var deletedID string
	mock.SoftDeleteFunc = func(ctx context.Context, id string) error {
		deletedID = id
		return nil
	}

	err := mock.SoftDelete(context.Background(), "source-to-delete")
	if err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	if deletedID != "source-to-delete" {
		t.Errorf("SoftDelete() did not receive correct ID")
	}
	if len(mock.SoftDeleteCalls) != 1 {
		t.Errorf("SoftDelete() should have been called once")
	}
}

func TestSourceRepository_GetByHash(t *testing.T) {
	mock := &MockSourceRepository{}
	mock.GetByHashFunc = func(ctx context.Context, chatbotID, hash string) (*models.DataSource, error) {
		if hash == "known-hash" {
			return &models.DataSource{ID: "existing-source", ChatbotID: chatbotID}, nil
		}
		return nil, nil
	}

	source, err := mock.GetByHash(context.Background(), "bot-1", "known-hash")
	if err != nil {
		t.Fatalf("GetByHash() error = %v", err)
	}
	if source == nil || source.ID != "existing-source" {
		t.Error("GetByHash() did not find existing source")
	}

	source, err = mock.GetByHash(context.Background(), "bot-1", "unknown-hash")
	if err != nil {
		t.Fatalf("GetByHash() error = %v", err)
	}
	if source != nil {
		t.Error("GetByHash() should return nil for unknown hash")
	}
}

// =============================================================================
// ErrVersionConflict Tests
// =============================================================================

func TestErrVersionConflict(t *testing.T) {
	err := repository.ErrVersionConflict
	if err.Error() != "version conflict: entity was modified by another request" {
		t.Errorf("ErrVersionConflict message = %q", err.Error())
	}

	// Test that it can be used with errors.Is
	wrappedErr := errors.New("wrapped: " + err.Error())
	if errors.Is(wrappedErr, repository.ErrVersionConflict) {
		t.Error("Non-wrapped error should not match with errors.Is")
	}
}

// =============================================================================
// Context Cancellation Tests
// =============================================================================

func TestChatbotRepository_ContextCancellation(t *testing.T) {
	mock := &MockChatbotRepository{}
	mock.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return &models.Chatbot{ID: id}, nil
		}
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := mock.GetByID(ctx, "test-id")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestActionRepository_ContextCancellation(t *testing.T) {
	mock := &MockActionRepository{}
	mock.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return []*models.ChatbotAction{}, nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := mock.List(ctx, "bot-1")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestSourceRepository_ContextCancellation(t *testing.T) {
	mock := &MockSourceRepository{}
	mock.GetByChatbotFunc = func(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return []models.DataSource{}, nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := mock.GetByChatbot(ctx, "bot-1")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}
