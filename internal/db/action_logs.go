package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/onurceri/botla-co/internal/models"
)

func CreateActionLog(ctx context.Context, db *sql.DB, log *models.ActionExecutionLog) error {
	query := `
		INSERT INTO action_execution_logs (
			chatbot_id, action_id, conversation_id, message_id,
			status, request_payload, response_payload, error_message,
			duration_ms
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	err := db.QueryRowContext(ctx, query,
		log.ChatbotID, log.ActionID, log.ConversationID, log.MessageID,
		log.Status, log.RequestPayload, log.ResponsePayload, log.ErrorMessage,
		log.DurationMs,
	).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return fmt.Errorf("create action log: %w", err)
	}
	return nil
}

func GetActionLogs(ctx context.Context, db *sql.DB, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
	query := `
		SELECT id, chatbot_id, action_id, conversation_id, message_id,
			   status, request_payload, response_payload, error_message,
			   duration_ms, created_at
		FROM action_execution_logs
		WHERE chatbot_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := db.QueryContext(ctx, query, chatbotID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query action logs: %w", err)
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
			return nil, fmt.Errorf("scan action log: %w", err)
		}
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("action logs rows err: %w", err)
	}
	return logs, nil
}
