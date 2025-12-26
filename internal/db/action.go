package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/onurceri/botla-co/internal/models"
)

// ErrVersionConflict is returned when an optimistic lock fails due to concurrent modification
var ErrVersionConflict = errors.New("version conflict: action was modified by another request")

func GetEnabledActions(ctx context.Context, db *sql.DB, chatbotID string) ([]*models.ChatbotAction, error) {
	query := `
		SELECT id, chatbot_id, name, description, action_type, config, parameters, tool_name, enabled, version, created_at, updated_at
		FROM chatbot_actions
		WHERE chatbot_id = $1 AND enabled = true
	`
	rows, err := db.QueryContext(ctx, query, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("query enabled actions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var actions []*models.ChatbotAction
	for rows.Next() {
		var a models.ChatbotAction
		if err := rows.Scan(
			&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
			&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan enabled action: %w", err)
		}
		actions = append(actions, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("enabled actions rows err: %w", err)
	}
	return actions, nil
}

func GetActions(ctx context.Context, db *sql.DB, chatbotID string) ([]*models.ChatbotAction, error) {
	query := `
		SELECT id, chatbot_id, name, description, action_type, config, parameters, tool_name, enabled, version, created_at, updated_at
		FROM chatbot_actions
		WHERE chatbot_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.QueryContext(ctx, query, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("query actions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var actions []*models.ChatbotAction
	for rows.Next() {
		var a models.ChatbotAction
		if err := rows.Scan(
			&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
			&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		actions = append(actions, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("actions rows err: %w", err)
	}
	return actions, nil
}

func CreateAction(ctx context.Context, db *sql.DB, action *models.ChatbotAction) error {
	query := `
		INSERT INTO chatbot_actions (chatbot_id, name, description, action_type, config, parameters, tool_name, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, version, created_at, updated_at
	`
	err := db.QueryRowContext(ctx, query,
		action.ChatbotID, action.Name, action.Description, action.ActionType,
		action.Config, action.Parameters, action.ToolName, action.Enabled,
	).Scan(&action.ID, &action.Version, &action.CreatedAt, &action.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create action: %w", err)
	}
	return nil
}

// UpdateAction updates an action with optimistic locking.
// Returns ErrVersionConflict if the action was modified by another request.
func UpdateAction(ctx context.Context, db *sql.DB, action *models.ChatbotAction) error {
	query := `
		UPDATE chatbot_actions
		SET name = $2, description = $3, action_type = $4, config = $5, parameters = $6, tool_name = $7, enabled = $8, version = version + 1, updated_at = NOW()
		WHERE id = $1 AND version = $9
		RETURNING version, updated_at
	`
	err := db.QueryRowContext(ctx, query,
		action.ID, action.Name, action.Description, action.ActionType,
		action.Config, action.Parameters, action.ToolName, action.Enabled, action.Version,
	).Scan(&action.Version, &action.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrVersionConflict
		}
		return fmt.Errorf("update action: %w", err)
	}
	return nil
}

func DeleteAction(ctx context.Context, db *sql.DB, id string) error {
	query := `DELETE FROM chatbot_actions WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete action: %w", err)
	}
	return nil
}

func GetActionByID(ctx context.Context, db *sql.DB, id string) (*models.ChatbotAction, error) {
	query := `
		SELECT id, chatbot_id, name, description, action_type, config, parameters, tool_name, enabled, version, created_at, updated_at
		FROM chatbot_actions
		WHERE id = $1
	`
	var a models.ChatbotAction
	err := db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
		&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get action by id: %w", err)
	}
	return &a, nil
}

// GetActionByToolName finds an action by its tool_name within a chatbot
func GetActionByToolName(ctx context.Context, db *sql.DB, chatbotID, toolName string) (*models.ChatbotAction, error) {
	query := `
		SELECT id, chatbot_id, name, description, action_type, config, parameters, tool_name, enabled, version, created_at, updated_at
		FROM chatbot_actions
		WHERE chatbot_id = $1 AND tool_name = $2 AND enabled = true
	`
	var a models.ChatbotAction
	err := db.QueryRowContext(ctx, query, chatbotID, toolName).Scan(
		&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
		&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get action by tool name: %w", err)
	}
	return &a, nil
}
