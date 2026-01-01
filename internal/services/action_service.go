package services

import (
	"context"

	"github.com/onurceri/botla-co/internal/models"
)

type ActionRepository interface {
	List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)
	GetByID(ctx context.Context, id string) (*models.ChatbotAction, error)
	Create(ctx context.Context, action *models.ChatbotAction) error
	Update(ctx context.Context, action *models.ChatbotAction) error
	Delete(ctx context.Context, id string) error
	GetLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error)
}

type ToolNameGenerator interface {
	Generate(ctx context.Context, name, description string) (string, error)
}

var ErrVersionConflict = &serviceError{msg: "action was modified by another request, please refresh and try again"}

type serviceError struct {
	msg string
}

func (e *serviceError) Error() string {
	return e.msg
}

type CreateActionInput struct {
	Name        string
	Description string
	ActionType  string
	Config      []byte
	Parameters  []byte
	Enabled     bool
}

type UpdateActionInput struct {
	Name        *string
	Description *string
	ActionType  *string
	Config      []byte
	Parameters  []byte
	Enabled     *bool
}

type ActionService interface {
	CreateAction(ctx context.Context, chatbotID string, input CreateActionInput) (*models.ChatbotAction, error)
	UpdateAction(ctx context.Context, actionID string, input UpdateActionInput) (*models.ChatbotAction, error)
	GetAction(ctx context.Context, actionID string) (*models.ChatbotAction, error)
	ListActions(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)
	DeleteAction(ctx context.Context, actionID string) error
	GetActionLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error)
}
