package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

type mockActionRepository struct {
	actions   map[string]*models.ChatbotAction
	logs      []*models.ActionExecutionLog
	createErr error
	updateErr error
	deleteErr error
	getErr    error
	listErr   error
	logsErr   error
}

func newMockActionRepository() *mockActionRepository {
	return &mockActionRepository{
		actions: make(map[string]*models.ChatbotAction),
		logs:    []*models.ActionExecutionLog{},
	}
}

func (m *mockActionRepository) List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*models.ChatbotAction
	for _, action := range m.actions {
		if action.ChatbotID == chatbotID {
			result = append(result, action)
		}
	}
	return result, nil
}

func (m *mockActionRepository) GetByID(ctx context.Context, id string) (*models.ChatbotAction, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.actions[id], nil
}

func (m *mockActionRepository) Create(ctx context.Context, action *models.ChatbotAction) error {
	if m.createErr != nil {
		return m.createErr
	}
	if action.ID == "" {
		action.ID = "test-action-id"
	}
	if action.Version == 0 {
		action.Version = 1
	}
	now := time.Now()
	action.CreatedAt = now
	action.UpdatedAt = &now
	m.actions[action.ID] = action
	return nil
}

func (m *mockActionRepository) Update(ctx context.Context, action *models.ChatbotAction) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.actions[action.ID]; !ok {
		return errors.New("action not found")
	}
	now := time.Now()
	action.Version++
	action.UpdatedAt = &now
	m.actions[action.ID] = action
	return nil
}

func (m *mockActionRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.actions, id)
	return nil
}

func (m *mockActionRepository) GetLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	if m.logsErr != nil {
		return nil, m.logsErr
	}
	return m.logs, nil
}

type mockToolNameGenerator struct {
	generateFunc func(ctx context.Context, name, description string) (string, error)
}

func (m *mockToolNameGenerator) Generate(ctx context.Context, name, description string) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, name, description)
	}
	return "generated_tool_name", nil
}

func TestCreateAction_Success(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	input := CreateActionInput{
		Name:        "Test Action",
		Description: "Test Description",
		ActionType:  "http",
		Config:      []byte(`{"url": "https://example.com"}`),
		Enabled:     true,
	}

	action, err := service.CreateAction(context.Background(), "chatbot-1", input)

	if err != nil {
		t.Fatalf("CreateAction() error = %v", err)
	}
	if action == nil {
		t.Fatal("CreateAction() returned nil action")
	}
	if action.Name != "Test Action" {
		t.Errorf("action.Name = %q; want %q", action.Name, "Test Action")
	}
	if action.ActionType != models.ActionTypeHTTP {
		t.Errorf("action.ActionType = %q; want %q", action.ActionType, models.ActionTypeHTTP)
	}
	if action.ToolName == nil || *action.ToolName != "generated_tool_name" {
		t.Errorf("action.ToolName = %v; want %q", action.ToolName, "generated_tool_name")
	}
	if action.ChatbotID != "chatbot-1" {
		t.Errorf("action.ChatbotID = %q; want %q", action.ChatbotID, "chatbot-1")
	}
}

func TestCreateAction_MissingName(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	input := CreateActionInput{
		Name:       "",
		ActionType: "http",
	}

	_, err := service.CreateAction(context.Background(), "chatbot-1", input)

	if !errors.Is(err, ErrActionNameRequired) {
		t.Errorf("CreateAction() error = %v; want ErrActionNameRequired", err)
	}
}

func TestCreateAction_MissingActionType(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	input := CreateActionInput{
		Name:       "Test Action",
		ActionType: "",
	}

	_, err := service.CreateAction(context.Background(), "chatbot-1", input)

	if !errors.Is(err, ErrActionTypeRequired) {
		t.Errorf("CreateAction() error = %v; want ErrActionTypeRequired", err)
	}
}

func TestCreateAction_ToolNameGenerationError(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{
		generateFunc: func(ctx context.Context, name, description string) (string, error) {
			return "", errors.New("LLM error")
		},
	}
	service := NewActionService(repo, generator)

	input := CreateActionInput{
		Name:       "Test Action",
		ActionType: "http",
	}

	_, err := service.CreateAction(context.Background(), "chatbot-1", input)

	if err == nil {
		t.Fatal("CreateAction() expected error, got nil")
	}
}

func TestUpdateAction_Success_NoToolNameRegen(t *testing.T) {
	existingAction := &models.ChatbotAction{
		ID:          "action-1",
		ChatbotID:   "chatbot-1",
		Name:        "Original Name",
		Description: strPtr("Original Description"),
		ActionType:  models.ActionTypeHTTP,
		ToolName:    strPtr("original_tool_name"),
		Version:     1,
	}
	repo := newMockActionRepository()
	repo.actions["action-1"] = existingAction

	generator := &mockToolNameGenerator{
		generateFunc: func(ctx context.Context, name, description string) (string, error) {
			t.Error("ToolNameGenerator.Generate() should not be called when name/description unchanged")
			return "should_not_be_called", nil
		},
	}
	service := NewActionService(repo, generator)

	enabled := true
	input := UpdateActionInput{
		Enabled: &enabled,
	}

	action, err := service.UpdateAction(context.Background(), "action-1", input)

	if err != nil {
		t.Fatalf("UpdateAction() error = %v", err)
	}
	if action.ToolName == nil || *action.ToolName != "original_tool_name" {
		t.Errorf("action.ToolName = %v; want %q", action.ToolName, "original_tool_name")
	}
	if action.Enabled != true {
		t.Errorf("action.Enabled = %v; want %v", action.Enabled, true)
	}
}

func TestUpdateAction_RegeneratesToolNameOnNameChange(t *testing.T) {
	existingAction := &models.ChatbotAction{
		ID:          "action-1",
		ChatbotID:   "chatbot-1",
		Name:        "Original Name",
		Description: strPtr("Description"),
		ActionType:  models.ActionTypeHTTP,
		ToolName:    strPtr("original_tool_name"),
		Version:     1,
	}
	repo := newMockActionRepository()
	repo.actions["action-1"] = existingAction

	generator := &mockToolNameGenerator{
		generateFunc: func(ctx context.Context, name, description string) (string, error) {
			if name != "New Name" {
				t.Errorf("Generate() name = %q; want %q", name, "New Name")
			}
			return "new_generated_tool_name", nil
		},
	}
	service := NewActionService(repo, generator)

	newName := "New Name"
	input := UpdateActionInput{
		Name: &newName,
	}

	action, err := service.UpdateAction(context.Background(), "action-1", input)

	if err != nil {
		t.Fatalf("UpdateAction() error = %v", err)
	}
	if action.ToolName == nil || *action.ToolName != "new_generated_tool_name" {
		t.Errorf("action.ToolName = %v; want %q", action.ToolName, "new_generated_tool_name")
	}
}

func TestUpdateAction_RegeneratesToolNameOnDescChange(t *testing.T) {
	existingAction := &models.ChatbotAction{
		ID:          "action-1",
		ChatbotID:   "chatbot-1",
		Name:        "Action Name",
		Description: strPtr("Old Description"),
		ActionType:  models.ActionTypeHTTP,
		ToolName:    strPtr("existing_tool_name"),
		Version:     1,
	}
	repo := newMockActionRepository()
	repo.actions["action-1"] = existingAction

	generator := &mockToolNameGenerator{
		generateFunc: func(ctx context.Context, name, description string) (string, error) {
			if description != "New Description" {
				t.Errorf("Generate() description = %q; want %q", description, "New Description")
			}
			return "new_tool_name", nil
		},
	}
	service := NewActionService(repo, generator)

	newDesc := "New Description"
	input := UpdateActionInput{
		Description: &newDesc,
	}

	action, err := service.UpdateAction(context.Background(), "action-1", input)

	if err != nil {
		t.Fatalf("UpdateAction() error = %v", err)
	}
	if action.ToolName == nil || *action.ToolName != "new_tool_name" {
		t.Errorf("action.ToolName = %v; want %q", action.ToolName, "new_tool_name")
	}
}

func TestUpdateAction_RegeneratesToolNameWhenMissing(t *testing.T) {
	existingAction := &models.ChatbotAction{
		ID:          "action-1",
		ChatbotID:   "chatbot-1",
		Name:        "Action Name",
		Description: strPtr("Description"),
		ActionType:  models.ActionTypeHTTP,
		ToolName:    nil,
		Version:     1,
	}
	repo := newMockActionRepository()
	repo.actions["action-1"] = existingAction

	generator := &mockToolNameGenerator{
		generateFunc: func(ctx context.Context, name, description string) (string, error) {
			return "regenerated_tool_name", nil
		},
	}
	service := NewActionService(repo, generator)

	enabled := true
	input := UpdateActionInput{
		Enabled: &enabled,
	}

	action, err := service.UpdateAction(context.Background(), "action-1", input)

	if err != nil {
		t.Fatalf("UpdateAction() error = %v", err)
	}
	if action.ToolName == nil || *action.ToolName != "regenerated_tool_name" {
		t.Errorf("action.ToolName = %v; want %q", action.ToolName, "regenerated_tool_name")
	}
}

func TestUpdateAction_ActionNotFound(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	input := UpdateActionInput{}
	_, err := service.UpdateAction(context.Background(), "non-existent", input)

	if !errors.Is(err, ErrActionNotFound) {
		t.Errorf("UpdateAction() error = %v; want ErrActionNotFound", err)
	}
}

func TestUpdateAction_VersionConflict(t *testing.T) {
	existingAction := &models.ChatbotAction{
		ID:        "action-1",
		ChatbotID: "chatbot-1",
		Name:      "Action Name",
		Version:   1,
	}
	repo := newMockActionRepository()
	repo.actions["action-1"] = existingAction
	repo.updateErr = repository.ErrVersionConflict

	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	input := UpdateActionInput{}
	_, err := service.UpdateAction(context.Background(), "action-1", input)

	if !errors.Is(err, ErrVersionConflict) {
		t.Errorf("UpdateAction() error = %v; want ErrVersionConflict", err)
	}
}

func TestGetAction_Success(t *testing.T) {
	existingAction := &models.ChatbotAction{
		ID:        "action-1",
		ChatbotID: "chatbot-1",
		Name:      "Test Action",
	}
	repo := newMockActionRepository()
	repo.actions["action-1"] = existingAction

	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	action, err := service.GetAction(context.Background(), "action-1")

	if err != nil {
		t.Fatalf("GetAction() error = %v", err)
	}
	if action == nil {
		t.Fatal("GetAction() returned nil")
	}
	if action.ID != "action-1" {
		t.Errorf("action.ID = %q; want %q", action.ID, "action-1")
	}
}

func TestGetAction_NotFound(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	action, err := service.GetAction(context.Background(), "non-existent")

	if err != nil {
		t.Fatalf("GetAction() error = %v", err)
	}
	if action != nil {
		t.Errorf("GetAction() = %v; want nil", action)
	}
}

func TestListActions_ReturnsEmptySlice(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	actions, err := service.ListActions(context.Background(), "chatbot-1")

	if err != nil {
		t.Fatalf("ListActions() error = %v", err)
	}
	if actions == nil {
		t.Fatal("ListActions() returned nil, want empty slice")
	}
	if len(actions) != 0 {
		t.Errorf("ListActions() length = %d; want 0", len(actions))
	}
}

func TestListActions_ReturnsActions(t *testing.T) {
	repo := newMockActionRepository()
	repo.actions["action-1"] = &models.ChatbotAction{
		ID:        "action-1",
		ChatbotID: "chatbot-1",
		Name:      "Action 1",
	}
	repo.actions["action-2"] = &models.ChatbotAction{
		ID:        "action-2",
		ChatbotID: "chatbot-1",
		Name:      "Action 2",
	}
	repo.actions["action-3"] = &models.ChatbotAction{
		ID:        "action-3",
		ChatbotID: "chatbot-2",
		Name:      "Action 3",
	}

	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	actions, err := service.ListActions(context.Background(), "chatbot-1")

	if err != nil {
		t.Fatalf("ListActions() error = %v", err)
	}
	if len(actions) != 2 {
		t.Errorf("ListActions() length = %d; want 2", len(actions))
	}
}

func TestDeleteAction_Success(t *testing.T) {
	repo := newMockActionRepository()
	repo.actions["action-1"] = &models.ChatbotAction{
		ID:        "action-1",
		ChatbotID: "chatbot-1",
	}

	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	err := service.DeleteAction(context.Background(), "action-1")

	if err != nil {
		t.Fatalf("DeleteAction() error = %v", err)
	}
	if _, exists := repo.actions["action-1"]; exists {
		t.Error("DeleteAction() did not remove action from repository")
	}
}

func TestGetActionLogs_ReturnsEmptySlice(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	logs, err := service.GetActionLogs(context.Background(), "chatbot-1", 20, 0)

	if err != nil {
		t.Fatalf("GetActionLogs() error = %v", err)
	}
	if logs == nil {
		t.Fatal("GetActionLogs() returned nil, want empty slice")
	}
	if len(logs) != 0 {
		t.Errorf("GetActionLogs() length = %d; want 0", len(logs))
	}
}

func TestGetActionLogs_ReturnsLogs(t *testing.T) {
	repo := newMockActionRepository()
	repo.logs = []*models.ActionExecutionLog{
		{ID: "log-1", ChatbotID: "chatbot-1"},
		{ID: "log-2", ChatbotID: "chatbot-1"},
	}

	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	logs, err := service.GetActionLogs(context.Background(), "chatbot-1", 20, 0)

	if err != nil {
		t.Fatalf("GetActionLogs() error = %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("GetActionLogs() length = %d; want 2", len(logs))
	}
}

func TestCreateAction_OptionalFields(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	input := CreateActionInput{
		Name:        "Test Action",
		Description: "",
		ActionType:  "http",
		Config:      nil,
		Parameters:  nil,
		Enabled:     false,
	}

	action, err := service.CreateAction(context.Background(), "chatbot-1", input)

	if err != nil {
		t.Fatalf("CreateAction() error = %v", err)
	}
	if action.Description != nil {
		t.Errorf("action.Description = %v; want nil", action.Description)
	}
	if action.Config != nil {
		t.Errorf("action.Config = %v; want nil", action.Config)
	}
	if action.Parameters != nil {
		t.Errorf("action.Parameters = %v; want nil", action.Parameters)
	}
}

func TestCreateAction_WithJSONConfig(t *testing.T) {
	repo := newMockActionRepository()
	generator := &mockToolNameGenerator{}
	service := NewActionService(repo, generator)

	config := json.RawMessage(`{"method": "GET", "headers": {"Authorization": "Bearer token"}}`)
	input := CreateActionInput{
		Name:       "Test Action",
		ActionType: "http",
		Config:     config,
	}

	action, err := service.CreateAction(context.Background(), "chatbot-1", input)

	if err != nil {
		t.Fatalf("CreateAction() error = %v", err)
	}
	if action.Config == nil {
		t.Fatal("action.Config is nil")
	}
	var dummy interface{}
	if err := json.Unmarshal([]byte(*action.Config), &dummy); err != nil {
		t.Error("action.Config is not valid JSON")
	}
}
