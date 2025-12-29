// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
)

// PostgresActionRepo implements ActionRepository using PostgreSQL.
// It delegates to the existing db package functions to provide a gradual migration path.
type PostgresActionRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresActionRepo implements ActionRepository.
var _ ActionRepository = (*PostgresActionRepo)(nil)

// NewPostgresActionRepo creates a new PostgresActionRepo instance.
func NewPostgresActionRepo(pool *sql.DB) *PostgresActionRepo {
	return &PostgresActionRepo{pool: pool}
}

// List returns all actions (enabled and disabled) for a chatbot, ordered by creation date descending.
func (r *PostgresActionRepo) List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	return db.GetActions(ctx, r.pool, chatbotID)
}

// ListEnabled returns only enabled actions for a chatbot.
func (r *PostgresActionRepo) ListEnabled(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	return db.GetEnabledActions(ctx, r.pool, chatbotID)
}

// GetByID retrieves an action by its unique identifier.
// Returns nil, nil if the action is not found.
func (r *PostgresActionRepo) GetByID(ctx context.Context, id string) (*models.ChatbotAction, error) {
	return db.GetActionByID(ctx, r.pool, id)
}

// GetByToolName finds an enabled action by its tool_name within a chatbot.
// Returns nil, nil if no matching action is found.
func (r *PostgresActionRepo) GetByToolName(ctx context.Context, chatbotID, toolName string) (*models.ChatbotAction, error) {
	return db.GetActionByToolName(ctx, r.pool, chatbotID, toolName)
}

// Create persists a new action. The action's ID, Version, CreatedAt, and UpdatedAt
// fields are populated after successful creation.
func (r *PostgresActionRepo) Create(ctx context.Context, action *models.ChatbotAction) error {
	return db.CreateAction(ctx, r.pool, action)
}

// Update modifies an existing action with optimistic locking.
// Returns ErrVersionConflict if the action was modified by another request.
func (r *PostgresActionRepo) Update(ctx context.Context, action *models.ChatbotAction) error {
	err := db.UpdateAction(ctx, r.pool, action)
	if err != nil {
		// Map db.ErrVersionConflict to repository.ErrVersionConflict
		if errors.Is(err, db.ErrVersionConflict) {
			return ErrVersionConflict
		}
		return err
	}
	return nil
}

// Delete permanently removes an action by its ID.
func (r *PostgresActionRepo) Delete(ctx context.Context, id string) error {
	return db.DeleteAction(ctx, r.pool, id)
}

// GetLogs retrieves action execution logs for a chatbot with pagination.
func (r *PostgresActionRepo) GetLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	return db.GetActionLogs(ctx, r.pool, chatbotID, limit, offset)
}

// CreateLog persists an action execution log entry.
func (r *PostgresActionRepo) CreateLog(ctx context.Context, log *models.ActionExecutionLog) error {
	return db.CreateActionLog(ctx, r.pool, log)
}
