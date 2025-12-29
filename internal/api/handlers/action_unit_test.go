package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/repository"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// Test UUIDs - using valid UUID format
const (
	testChatbotID    = "00000000-0000-0000-0000-000000000001"
	testUserID       = "00000000-0000-0000-0000-000000000002"
	testActionID     = "00000000-0000-0000-0000-000000000003"
	otherChatbotID   = "00000000-0000-0000-0000-999999999999"
)

// mockToolsLLMClient implements rag.ToolsLLMClient for testing
type mockToolsLLMClient struct{}

func (m *mockToolsLLMClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	return &models.CompletionResult{}, nil
}

func (m *mockToolsLLMClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{Name: "mock"}
}

func (m *mockToolsLLMClient) CreateCompletionWithTools(ctx context.Context, messages []rag.ChatMessage, tools []rag.Tool, model string, temperature float32, maxTokens int) (*rag.ChatResponseWithTools, error) {
	// Use JSON unmarshal to construct the response since Choices uses anonymous struct
	jsonData := `{"choices":[{"message":{"role":"assistant","content":"mock_tool_name"},"finish_reason":"stop"}],"usage":{"total_tokens":10}}`
	var resp rag.ChatResponseWithTools
	_ = json.Unmarshal([]byte(jsonData), &resp)
	return &resp, nil
}

// setupActionHandlersTest creates an ActionHandlers with mocks for testing
func setupActionHandlersTest() (*ActionHandlers, *repository.MockActionRepo, *repository.MockChatbotRepo) {
	actionRepo := repository.NewMockActionRepo()
	chatbotRepo := repository.NewMockChatbotRepo()
	mockClient := &mockToolsLLMClient{}
	toolNameGenerator := rag.NewToolNameGenerator(mockClient)

	h := &ActionHandlers{
		ActionRepo:        actionRepo,
		ChatbotRepo:       chatbotRepo,
		ToolNameGenerator: toolNameGenerator,
	}

	return h, actionRepo, chatbotRepo
}

// createAuthenticatedRequest creates a request with a valid user context
func createAuthenticatedRequest(method, path string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	// Use valid UUIDs to pass httputil.IsValidUUID check
	req.SetPathValue("id", testChatbotID)

	// Add user ID to context
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, testUserID)
	return req.WithContext(ctx)
}

func TestActionHandlers_List_Success(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	// Setup chatbot repo to return a valid chatbot owned by the user
	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{
			ID:     id,
			UserID: testUserID, // Must match the user in context
			Name:   "Test Bot",
		}, nil
	}

	// Setup action repo to return actions
	actionRepo.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		return []*models.ChatbotAction{
			{ID: "action-1", ChatbotID: chatbotID, Name: "Action 1", Enabled: true},
			{ID: "action-2", ChatbotID: chatbotID, Name: "Action 2", Enabled: false},
		}, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify response contains actions
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	actions, ok := resp["actions"].([]any)
	if !ok {
		t.Fatalf("expected actions array in response")
	}
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}

	// Verify action repo was called with correct chatbot ID
	if len(actionRepo.Calls.List) != 1 {
		t.Errorf("expected 1 List call, got %d", len(actionRepo.Calls.List))
	}
	if actionRepo.Calls.List[0].ChatbotID != testChatbotID {
		t.Errorf("expected chatbotID %q, got %s", testChatbotID, actionRepo.Calls.List[0].ChatbotID)
	}
}

func TestActionHandlers_List_EmptyActions(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	// Return nil (no actions)
	actionRepo.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		return nil, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Should return empty array, not null
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	actions, ok := resp["actions"].([]any)
	if !ok {
		t.Fatalf("expected actions array in response")
	}
	if len(actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(actions))
	}
}

func TestActionHandlers_List_RepoError(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		return nil, errors.New("database error")
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_Get_Success(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	toolName := "test_tool"
	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return &models.ChatbotAction{
			ID:        id,
			ChatbotID: testChatbotID, // Must match the chatbot in context
			Name:      "Test Action",
			ToolName:  &toolName,
			Enabled:   true,
		}, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify response
	var action models.ChatbotAction
	if err := json.Unmarshal(w.Body.Bytes(), &action); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if action.Name != "Test Action" {
		t.Errorf("expected name 'Test Action', got %s", action.Name)
	}
}

func TestActionHandlers_Get_NotFound(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	// Return nil (action not found)
	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return nil, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/nonexistent", nil)
	req.SetPathValue("actionId", "nonexistent")
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestActionHandlers_Get_WrongChatbot(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	// Return action belonging to a different chatbot
	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return &models.ChatbotAction{
			ID:        id,
			ChatbotID: otherChatbotID, // Different chatbot
			Name:      "Test Action",
		}, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d (access denied treated as not found), got %d", http.StatusNotFound, w.Code)
	}
}

func TestActionHandlers_Delete_Success(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return &models.ChatbotAction{
			ID:        id,
			ChatbotID: testChatbotID,
			Name:      "Test Action",
		}, nil
	}

	actionRepo.DeleteFunc = func(ctx context.Context, id string) error {
		return nil
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify delete was called
	if len(actionRepo.Calls.Delete) != 1 {
		t.Fatalf("expected 1 Delete call, got %d", len(actionRepo.Calls.Delete))
	}
	if actionRepo.Calls.Delete[0].ID != testActionID {
		t.Errorf("expected action ID %q, got %s", testActionID, actionRepo.Calls.Delete[0].ID)
	}
}

func TestActionHandlers_Delete_NotFound(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	// Action not found
	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return nil, nil
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/v1/chatbots/"+testChatbotID+"/actions/nonexistent", nil)
	req.SetPathValue("actionId", "nonexistent")
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	// Verify delete was NOT called
	if len(actionRepo.Calls.Delete) != 0 {
		t.Errorf("expected 0 Delete calls, got %d", len(actionRepo.Calls.Delete))
	}
}

func TestActionHandlers_Update_VersionConflict(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	toolName := "original_tool"
	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return &models.ChatbotAction{
			ID:        id,
			ChatbotID: testChatbotID,
			Name:      "Test Action",
			ToolName:  &toolName,
			Version:   1,
		}, nil
	}

	// Simulate version conflict
	actionRepo.UpdateFunc = func(ctx context.Context, action *models.ChatbotAction) error {
		return repository.ErrVersionConflict
	}

	body, _ := json.Marshal(map[string]any{
		"name":    "Updated Name",
		"enabled": true,
	})

	req := createAuthenticatedRequest(http.MethodPut, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, body)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestActionHandlers_GetLogs_Success(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
		return []*models.ActionExecutionLog{
			{ID: "log-1", ChatbotID: chatbotID, ActionID: "action-1"},
			{ID: "log-2", ChatbotID: chatbotID, ActionID: "action-2"},
		}, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/logs?page=1&limit=20", nil)
	w := httptest.NewRecorder()

	h.GetLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify logs are in response
	logs, ok := resp["logs"].([]any)
	if !ok {
		t.Fatalf("expected logs array in response")
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}

	// Verify pagination info
	if page, ok := resp["page"].(float64); !ok || int(page) != 1 {
		t.Errorf("expected page 1 in response")
	}
	if limit, ok := resp["limit"].(float64); !ok || int(limit) != 20 {
		t.Errorf("expected limit 20 in response")
	}
}

func TestActionHandlers_GetLogs_Pagination(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	var capturedOffset, capturedLimit int
	actionRepo.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
		capturedLimit = limit
		capturedOffset = offset
		return []*models.ActionExecutionLog{}, nil
	}

	// Page 3 with limit 10 means offset = (3-1) * 10 = 20
	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/logs?page=3&limit=10", nil)
	w := httptest.NewRecorder()

	h.GetLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if capturedOffset != 20 {
		t.Errorf("expected offset 20, got %d", capturedOffset)
	}
	if capturedLimit != 10 {
		t.Errorf("expected limit 10, got %d", capturedLimit)
	}
}

func TestActionHandlers_AuthorizationCheck_ChatbotNotFound(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	// Chatbot not found
	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return nil, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for nonexistent chatbot, got %d", http.StatusNotFound, w.Code)
	}
}

func TestActionHandlers_ChatbotRepoError(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	// Simulate database error
	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return nil, errors.New("database connection error")
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_CallTracking(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.ListFunc = func(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
		return []*models.ChatbotAction{}, nil
	}

	// Make multiple calls
	for i := 0; i < 3; i++ {
		req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
		w := httptest.NewRecorder()
		h.List(w, req)
	}

	// Verify call count
	if len(actionRepo.Calls.List) != 3 {
		t.Errorf("expected 3 List calls, got %d", len(actionRepo.Calls.List))
	}

	// Reset and verify
	actionRepo.Reset()
	if len(actionRepo.Calls.List) != 0 {
		t.Errorf("expected 0 calls after Reset, got %d", len(actionRepo.Calls.List))
	}
}

func TestActionHandlers_Delete_RepoError(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return &models.ChatbotAction{
			ID:        id,
			ChatbotID: testChatbotID,
			Name:      "Test Action",
		}, nil
	}

	actionRepo.DeleteFunc = func(ctx context.Context, id string) error {
		return errors.New("database error")
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_Get_RepoError(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.ChatbotAction, error) {
		return nil, errors.New("database error")
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_GetLogs_RepoError(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
		return nil, errors.New("database error")
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/logs", nil)
	w := httptest.NewRecorder()

	h.GetLogs(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_GetLogs_EmptyLogs(t *testing.T) {
	h, actionRepo, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionRepo.GetLogsFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
		return nil, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/logs", nil)
	w := httptest.NewRecorder()

	h.GetLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should return empty array, not null
	logs, ok := resp["logs"].([]any)
	if !ok {
		t.Fatalf("expected logs array in response")
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}
