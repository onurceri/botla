// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// psql is the PostgreSQL-compatible Squirrel SQL builder.
// It uses $1, $2, etc. placeholders instead of ? for PostgreSQL.
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// PostgresActionRepo implements ActionRepository using PostgreSQL.
// SQL queries are built using Squirrel for type safety and maintainability.
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
	query, args, err := psql.
		Select("id", "chatbot_id", "name", "description", "action_type", "config", "parameters", "tool_name", "enabled", "version", "created_at", "updated_at").
		From("chatbot_actions").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build list actions query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query actions")
	}
	defer func() { _ = rows.Close() }()

	var actions []*models.ChatbotAction
	for rows.Next() {
		var a models.ChatbotAction
		if err := rows.Scan(
			&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
			&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan action")
		}
		actions = append(actions, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "actions rows error")
	}
	return actions, nil
}

// ListEnabled returns only enabled actions for a chatbot.
func (r *PostgresActionRepo) ListEnabled(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
	query, args, err := psql.
		Select("id", "chatbot_id", "name", "description", "action_type", "config", "parameters", "tool_name", "enabled", "version", "created_at", "updated_at").
		From("chatbot_actions").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"enabled": true}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build list enabled actions query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query enabled actions")
	}
	defer func() { _ = rows.Close() }()

	var actions []*models.ChatbotAction
	for rows.Next() {
		var a models.ChatbotAction
		if err := rows.Scan(
			&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
			&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan enabled action")
		}
		actions = append(actions, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "enabled actions rows error")
	}
	return actions, nil
}

// GetByID retrieves an action by its unique identifier.
// Returns nil, nil if the action is not found.
func (r *PostgresActionRepo) GetByID(ctx context.Context, id string) (*models.ChatbotAction, error) {
	query, args, err := psql.
		Select("id", "chatbot_id", "name", "description", "action_type", "config", "parameters", "tool_name", "enabled", "version", "created_at", "updated_at").
		From("chatbot_actions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get action by id query")
	}

	var a models.ChatbotAction
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
		&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get action by id")
	}
	return &a, nil
}

// GetByToolName finds an enabled action by its tool_name within a chatbot.
// Returns nil, nil if no matching action is found.
func (r *PostgresActionRepo) GetByToolName(ctx context.Context, chatbotID, toolName string) (*models.ChatbotAction, error) {
	query, args, err := psql.
		Select("id", "chatbot_id", "name", "description", "action_type", "config", "parameters", "tool_name", "enabled", "version", "created_at", "updated_at").
		From("chatbot_actions").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"tool_name": toolName}).
		Where(sq.Eq{"enabled": true}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get action by tool name query")
	}

	var a models.ChatbotAction
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&a.ID, &a.ChatbotID, &a.Name, &a.Description, &a.ActionType,
		&a.Config, &a.Parameters, &a.ToolName, &a.Enabled, &a.Version, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get action by tool name")
	}
	return &a, nil
}

// Create persists a new action. The action's ID, Version, CreatedAt, and UpdatedAt
// fields are populated after successful creation.
func (r *PostgresActionRepo) Create(ctx context.Context, action *models.ChatbotAction) error {
	query, args, err := psql.
		Insert("chatbot_actions").
		Columns("chatbot_id", "name", "description", "action_type", "config", "parameters", "tool_name", "enabled").
		Values(action.ChatbotID, action.Name, action.Description, action.ActionType, action.Config, action.Parameters, action.ToolName, action.Enabled).
		Suffix("RETURNING id, version, created_at, updated_at").
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build create action query")
	}

	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&action.ID, &action.Version, &action.CreatedAt, &action.UpdatedAt)
	if err != nil {
		return pkgerrors.Wrapf(err, "create action")
	}
	return nil
}

// Update modifies an existing action with optimistic locking.
// Returns ErrVersionConflict if the action was modified by another request.
func (r *PostgresActionRepo) Update(ctx context.Context, action *models.ChatbotAction) error {
	query, args, err := psql.
		Update("chatbot_actions").
		Set("name", action.Name).
		Set("description", action.Description).
		Set("action_type", action.ActionType).
		Set("config", action.Config).
		Set("parameters", action.Parameters).
		Set("tool_name", action.ToolName).
		Set("enabled", action.Enabled).
		Set("version", sq.Expr("version + 1")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": action.ID}).
		Where(sq.Eq{"version": action.Version}).
		Suffix("RETURNING version, updated_at").
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update action query")
	}

	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&action.Version, &action.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrVersionConflict
		}
		return pkgerrors.Wrapf(err, "update action")
	}
	return nil
}

// Delete permanently removes an action by its ID.
func (r *PostgresActionRepo) Delete(ctx context.Context, id string) error {
	query, args, err := psql.
		Delete("chatbot_actions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build delete action query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete action")
	}
	return nil
}

// GetLogs retrieves action execution logs for a chatbot with pagination.
func (r *PostgresActionRepo) GetLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	query, args, err := psql.
		Select("id", "chatbot_id", "action_id", "conversation_id", "message_id", "status", "request_payload", "response_payload", "error_message", "duration_ms", "created_at").
		From("action_execution_logs").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get action logs query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query action logs")
	}
	defer func() { _ = rows.Close() }()

	var logs []*models.ActionExecutionLog
	for rows.Next() {
		var l models.ActionExecutionLog
		if err := rows.Scan(
			&l.ID, &l.ChatbotID, &l.ActionID, &l.ConversationID, &l.MessageID,
			&l.Status, &l.RequestPayload, &l.ResponsePayload, &l.ErrorMessage,
			&l.DurationMs, &l.CreatedAt,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan action log")
		}
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "action logs rows error")
	}
	return logs, nil
}

// CreateLog persists an action execution log entry.
func (r *PostgresActionRepo) CreateLog(ctx context.Context, log *models.ActionExecutionLog) error {
	query, args, err := psql.
		Insert("action_execution_logs").
		Columns("chatbot_id", "action_id", "conversation_id", "message_id", "status", "request_payload", "response_payload", "error_message", "duration_ms").
		Values(log.ChatbotID, log.ActionID, log.ConversationID, log.MessageID, log.Status, log.RequestPayload, log.ResponsePayload, log.ErrorMessage, log.DurationMs).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build create action log query")
	}

	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return pkgerrors.Wrapf(err, "create action log")
	}
	return nil
}
