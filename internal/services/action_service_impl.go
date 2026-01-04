package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

var (
	ErrActionNameRequired = errors.New("action name is required")
	ErrActionTypeRequired = errors.New("action type is required")
	ErrActionNotFound     = errors.New("action not found")
)

type ActionServiceImpl struct {
	actionRepo        ActionRepository
	toolNameGenerator ToolNameGenerator
}

func NewActionService(
	actionRepo ActionRepository,
	toolNameGenerator ToolNameGenerator,
) *ActionServiceImpl {
	return &ActionServiceImpl{
		actionRepo:        actionRepo,
		toolNameGenerator: toolNameGenerator,
	}
}

func (s *ActionServiceImpl) CreateAction(ctx context.Context, chatbotID string, input CreateActionInput) (*models.ChatbotAction, error) {
	if input.Name == "" {
		return nil, ErrActionNameRequired
	}
	if input.ActionType == "" {
		return nil, ErrActionTypeRequired
	}

	toolName, err := s.toolNameGenerator.Generate(ctx, input.Name, input.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tool name: %w", err)
	}

	action := &models.ChatbotAction{
		ChatbotID:  chatbotID,
		Name:       input.Name,
		ActionType: models.ActionType(input.ActionType),
		ToolName:   &toolName,
		Enabled:    input.Enabled,
	}

	if input.Description != "" {
		action.Description = &input.Description
	}
	if len(input.Config) > 0 {
		cfg := json.RawMessage(input.Config)
		action.Config = &cfg
	}
	if len(input.Parameters) > 0 {
		params := json.RawMessage(input.Parameters)
		action.Parameters = &params
	}

	if err := s.actionRepo.Create(ctx, action); err != nil {
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	return action, nil
}

func (s *ActionServiceImpl) UpdateAction(ctx context.Context, actionID string, input UpdateActionInput) (*models.ChatbotAction, error) {
	action, err := s.actionRepo.GetByID(ctx, actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
	}
	if action == nil {
		return nil, ErrActionNotFound
	}

	nameChanged := input.Name != nil && *input.Name != action.Name
	descChanged := input.Description != nil && (action.Description == nil || *input.Description != *action.Description)
	toolNameMissing := action.ToolName == nil || *action.ToolName == ""

	if nameChanged || descChanged || toolNameMissing {
		newName := action.Name
		if input.Name != nil {
			newName = *input.Name
		}

		newDesc := ""
		if input.Description != nil {
			newDesc = *input.Description
		} else if action.Description != nil {
			newDesc = *action.Description
		}

		toolName, err := s.toolNameGenerator.Generate(ctx, newName, newDesc)
		if err != nil {
			return nil, fmt.Errorf("failed to generate tool name: %w", err)
		}
		action.ToolName = &toolName
	}

	if input.Name != nil {
		action.Name = *input.Name
	}
	if input.Description != nil {
		action.Description = input.Description
	}
	if input.ActionType != nil {
		action.ActionType = models.ActionType(*input.ActionType)
	}
	if len(input.Config) > 0 {
		cfg := json.RawMessage(input.Config)
		action.Config = &cfg
	}
	if len(input.Parameters) > 0 {
		params := json.RawMessage(input.Parameters)
		action.Parameters = &params
	}
	if input.Enabled != nil {
		action.Enabled = *input.Enabled
	}

	if err := s.actionRepo.Update(ctx, action); err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("failed to update action: %w", err)
	}

	return action, nil
}

func (s *ActionServiceImpl) GetAction(ctx context.Context, actionID string) (*models.ChatbotAction, error) {
	return s.actionRepo.GetByID(ctx, actionID)
}

func (s *ActionServiceImpl) ListActions(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	actions, err := s.actionRepo.List(ctx, chatbotID)
	if err != nil {
		return nil, err
	}
	if actions == nil {
		return []*models.ChatbotAction{}, nil
	}
	return actions, nil
}

func (s *ActionServiceImpl) DeleteAction(ctx context.Context, actionID string) error {
	return s.actionRepo.Delete(ctx, actionID)
}

func (s *ActionServiceImpl) GetActionLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	logs, err := s.actionRepo.GetLogs(ctx, chatbotID, limit, offset)
	if err != nil {
		return nil, err
	}
	if logs == nil {
		return []*models.ActionExecutionLog{}, nil
	}
	return logs, nil
}
