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
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

const (
	testChatbotID  = "00000000-0000-0000-0000-000000000001"
	testUserID     = "00000000-0000-0000-0000-000000000002"
	testActionID   = "00000000-0000-0000-0000-000000000003"
	otherChatbotID = "00000000-0000-0000-0000-999999999999"
)

type mockActionService struct {
	actions      []*models.ChatbotAction
	logs         []*models.ActionExecutionLog
	createErr    error
	updateErr    error
	deleteErr    error
	getErr       error
	listErr      error
	logsErr      error
	lastInput    services.CreateActionInput
	lastUpdateID string
	lastUpdateIn services.UpdateActionInput
}

func (m *mockActionService) CreateAction(ctx context.Context, chatbotID string, input services.CreateActionInput) (*models.ChatbotAction, error) {
	m.lastInput = input
	if m.createErr != nil {
		if errors.Is(m.createErr, services.ErrActionNameRequired) || errors.Is(m.createErr, services.ErrActionTypeRequired) {
			return nil, m.createErr
		}
		return nil, m.createErr
	}
	toolName := "generated_tool_name"
	return &models.ChatbotAction{
		ID:          testActionID,
		ChatbotID:   chatbotID,
		Name:        input.Name,
		Description: strPtr(input.Description),
		ActionType:  models.ActionType(input.ActionType),
		ToolName:    &toolName,
		Enabled:     input.Enabled,
	}, nil
}

func (m *mockActionService) UpdateAction(ctx context.Context, actionID string, input services.UpdateActionInput) (*models.ChatbotAction, error) {
	m.lastUpdateID = actionID
	m.lastUpdateIn = input
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	toolName := "updated_tool_name"
	return &models.ChatbotAction{
		ID:        actionID,
		ChatbotID: testChatbotID,
		Name:      "Updated Action",
		ToolName:  &toolName,
		Version:   2,
	}, nil
}

func (m *mockActionService) GetAction(ctx context.Context, actionID string) (*models.ChatbotAction, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, a := range m.actions {
		if a.ID == actionID {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockActionService) ListActions(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := []*models.ChatbotAction{}
	for _, a := range m.actions {
		if a.ChatbotID == chatbotID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockActionService) DeleteAction(ctx context.Context, actionID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}

func (m *mockActionService) GetActionLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	if m.logsErr != nil {
		return nil, m.logsErr
	}
	return m.logs, nil
}

func strPtr(s string) *string {
	return &s
}

func setupActionHandlersTest() (*ActionHandlers, *mockActionService, *mockChatbotRepo) {
	actionService := &mockActionService{
		actions: []*models.ChatbotAction{},
		logs:    []*models.ActionExecutionLog{},
	}
	chatbotRepo := &mockChatbotRepo{}

	h := &ActionHandlers{
		ActionService: actionService,
		ChatbotRepo:   chatbotRepo,
	}

	return h, actionService, chatbotRepo
}

type mockChatbotRepo struct {
	getByIDFunc func(ctx context.Context, id string) (*models.Chatbot, error)
}

func (r *mockChatbotRepo) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (r *mockChatbotRepo) GetByUserID(ctx context.Context, userID string) ([]models.Chatbot, error) {
	return nil, nil
}

func (r *mockChatbotRepo) GetByWorkspace(ctx context.Context, workspaceID string) ([]models.Chatbot, error) {
	return nil, nil
}

func (r *mockChatbotRepo) Create(ctx context.Context, bot *models.Chatbot) (string, error) {
	return "", nil
}

func (r *mockChatbotRepo) Update(ctx context.Context, bot *models.Chatbot) error {
	return nil
}

func (r *mockChatbotRepo) SoftDelete(ctx context.Context, id, userID string) ([]string, error) {
	return nil, nil
}

func (r *mockChatbotRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (r *mockChatbotRepo) CountByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	return 0, nil
}

func (r *mockChatbotRepo) UpdateSuggestedQuestions(ctx context.Context, id string, suggestions []string) error {
	return nil
}

func createAuthenticatedRequest(method, path string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.SetPathValue("id", testChatbotID)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, testUserID)
	return req.WithContext(ctx)
}

func TestActionHandlers_List_Success(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{
			ID:     id,
			UserID: testUserID,
			Name:   "Test Bot",
		}, nil
	}

	actionService.actions = []*models.ChatbotAction{
		{ID: "action-1", ChatbotID: testChatbotID, Name: "Action 1", Enabled: true},
		{ID: "action-2", ChatbotID: testChatbotID, Name: "Action 2", Enabled: false},
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

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
}

func TestActionHandlers_List_EmptyActions(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	actions, ok := resp["actions"].([]any)
	if !ok {
		t.Fatalf("expected actions array in response, got %T", resp["actions"])
	}
	if len(actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(actions))
	}
}

func TestActionHandlers_List_ServiceError(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.listErr = errors.New("database error")

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_Get_Success(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	toolName := "test_tool"
	actionService.actions = []*models.ChatbotAction{
		{
			ID:        testActionID,
			ChatbotID: testChatbotID,
			Name:      "Test Action",
			ToolName:  &toolName,
			Enabled:   true,
		},
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var action models.ChatbotAction
	if err := json.Unmarshal(w.Body.Bytes(), &action); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if action.Name != "Test Action" {
		t.Errorf("expected name 'Test Action', got %s", action.Name)
	}
}

func TestActionHandlers_Get_NotFound(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
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
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.actions = []*models.ChatbotAction{
		{
			ID:        testActionID,
			ChatbotID: otherChatbotID,
			Name:      "Test Action",
		},
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
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.actions = []*models.ChatbotAction{
		{
			ID:        testActionID,
			ChatbotID: testChatbotID,
			Name:      "Test Action",
		},
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestActionHandlers_Delete_NotFound(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/v1/chatbots/"+testChatbotID+"/actions/nonexistent", nil)
	req.SetPathValue("actionId", "nonexistent")
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestActionHandlers_Update_VersionConflict(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	toolName := "original_tool"
	actionService.actions = []*models.ChatbotAction{
		{
			ID:        testActionID,
			ChatbotID: testChatbotID,
			Name:      "Test Action",
			ToolName:  &toolName,
			Version:   1,
		},
	}

	actionService.updateErr = services.ErrVersionConflict

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
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.logs = []*models.ActionExecutionLog{
		{ID: "log-1", ChatbotID: testChatbotID, ActionID: "action-1"},
		{ID: "log-2", ChatbotID: testChatbotID, ActionID: "action-2"},
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

	logs, ok := resp["logs"].([]any)
	if !ok {
		t.Fatalf("expected logs array in response")
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}

	if page, ok := resp["page"].(float64); !ok || int(page) != 1 {
		t.Errorf("expected page 1 in response")
	}
	if limit, ok := resp["limit"].(float64); !ok || int(limit) != 20 {
		t.Errorf("expected limit 20 in response")
	}
}

func TestActionHandlers_GetLogs_Pagination(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.logs = []*models.ActionExecutionLog{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/logs?page=3&limit=10", nil)
	w := httptest.NewRecorder()

	h.GetLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestActionHandlers_AuthorizationCheck_ChatbotNotFound(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
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

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return nil, errors.New("database connection error")
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_Delete_ServiceError(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.actions = []*models.ChatbotAction{
		{
			ID:        testActionID,
			ChatbotID: testChatbotID,
			Name:      "Test Action",
		},
	}

	actionService.deleteErr = errors.New("database error")

	req := createAuthenticatedRequest(http.MethodDelete, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_Get_ServiceError(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.getErr = errors.New("database error")

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/"+testActionID, nil)
	req.SetPathValue("actionId", testActionID)
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_GetLogs_ServiceError(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.logsErr = errors.New("database error")

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/chatbots/"+testChatbotID+"/actions/logs", nil)
	w := httptest.NewRecorder()

	h.GetLogs(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestActionHandlers_GetLogs_EmptyLogs(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
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

	logs, ok := resp["logs"].([]any)
	if !ok {
		t.Fatalf("expected logs array in response")
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}

func TestActionHandlers_Create_MissingName(t *testing.T) {
	h, actionService, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	actionService.createErr = services.ErrActionNameRequired

	body, _ := json.Marshal(map[string]any{
		"action_type": "http",
	})

	req := createAuthenticatedRequest(http.MethodPost, "/api/v1/chatbots/"+testChatbotID+"/actions", body)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestActionHandlers_Update_NotFound(t *testing.T) {
	h, _, chatbotRepo := setupActionHandlersTest()

	chatbotRepo.getByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testUserID}, nil
	}

	body, _ := json.Marshal(map[string]any{
		"name": "Updated Name",
	})

	req := createAuthenticatedRequest(http.MethodPut, "/api/v1/chatbots/"+testChatbotID+"/actions/nonexistent", body)
	req.SetPathValue("actionId", "nonexistent")
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
