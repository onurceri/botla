package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

// CreateHandoffRequest creates a new handoff request
func CreateHandoffRequest(ctx context.Context, pool *sql.DB, req *models.HandoffRequest) (string, error) {
	var id string
	err := pool.QueryRowContext(ctx, `
		INSERT INTO handoff_requests (chatbot_id, conversation_id, status, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		req.ChatbotID, req.ConversationID, models.HandoffStatusPending, req.Notes,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetHandoffRequestsByBotID returns all handoff requests for a chatbot
func GetHandoffRequestsByBotID(ctx context.Context, pool *sql.DB, chatbotID string) ([]*models.HandoffRequest, error) {
	rows, err := pool.QueryContext(ctx, `
		SELECT id, chatbot_id, conversation_id, status, assigned_to, notes, user_email, created_at, resolved_at
		FROM handoff_requests
		WHERE chatbot_id = $1
		ORDER BY created_at DESC`, chatbotID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var requests []*models.HandoffRequest
	for rows.Next() {
		var req models.HandoffRequest
		if err := rows.Scan(
			&req.ID, &req.ChatbotID, &req.ConversationID, &req.Status,
			&req.AssignedTo, &req.Notes, &req.UserEmail, &req.CreatedAt, &req.ResolvedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

// GetHandoffRequestByID returns a single handoff request
func GetHandoffRequestByID(ctx context.Context, pool *sql.DB, id string) (*models.HandoffRequest, error) {
	var req models.HandoffRequest
	err := pool.QueryRowContext(ctx, `
		SELECT id, chatbot_id, conversation_id, status, assigned_to, notes, user_email, created_at, resolved_at
		FROM handoff_requests
		WHERE id = $1`, id).Scan(
		&req.ID, &req.ChatbotID, &req.ConversationID, &req.Status,
		&req.AssignedTo, &req.Notes, &req.UserEmail, &req.CreatedAt, &req.ResolvedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

// UpdateHandoffRequestStatus updates the status of a handoff request
func UpdateHandoffRequestStatus(ctx context.Context, pool *sql.DB, id, status string, assignedTo *string) error {
	var resolvedAt interface{}
	if status == models.HandoffStatusResolved {
		resolvedAt = time.Now()
	}
	_, err := pool.ExecContext(ctx, `
		UPDATE handoff_requests
		SET status = $1, assigned_to = $2, resolved_at = $3
		WHERE id = $4`,
		status, assignedTo, resolvedAt, id)
	return err
}

// CountPendingHandoffRequests returns the count of pending handoff requests for a chatbot
func CountPendingHandoffRequests(ctx context.Context, pool *sql.DB, chatbotID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM handoff_requests
		WHERE chatbot_id = $1 AND status = $2`,
		chatbotID, models.HandoffStatusPending).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// HasActiveHandoffRequest checks if there is any pending or assigned handoff request for the conversation
func HasActiveHandoffRequest(ctx context.Context, pool *sql.DB, conversationID string) (bool, error) {
	var exists bool
	err := pool.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM handoff_requests
			WHERE conversation_id = $1
			AND status IN ($2, $3)
		)`,
		conversationID, models.HandoffStatusPending, models.HandoffStatusAssigned,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// UpdateHandoffUserEmail updates the user email for a handoff request
func UpdateHandoffUserEmail(ctx context.Context, pool *sql.DB, requestID, email string) error {
	_, err := pool.ExecContext(ctx, `
		UPDATE handoff_requests
		SET user_email = $1
		WHERE id = $2`,
		email, requestID)
	return err
}

// HandoffRequestDetail contains a handoff request with its conversation messages
type HandoffRequestDetail struct {
	Request  *models.HandoffRequest `json:"request"`
	Messages []models.Message       `json:"messages"`
}

// GetHandoffRequestWithMessages returns a handoff request with its conversation messages
func GetHandoffRequestWithMessages(ctx context.Context, pool *sql.DB, requestID string) (*HandoffRequestDetail, error) {
	// Get the request
	req, err := GetHandoffRequestByID(ctx, pool, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, nil
	}

	// Get conversation messages
	msgs, err := ListRecentMessages(ctx, pool, req.ConversationID, 100)
	if err != nil {
		return nil, err
	}

	return &HandoffRequestDetail{
		Request:  req,
		Messages: msgs,
	}, nil
}
