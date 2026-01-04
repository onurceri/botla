package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresHandoffRepo implements HandoffRepository using PostgreSQL.
type PostgresHandoffRepo struct {
	pool *sql.DB
}

var _ HandoffRepository = (*PostgresHandoffRepo)(nil)

func NewPostgresHandoffRepo(pool *sql.DB) *PostgresHandoffRepo {
	return &PostgresHandoffRepo{pool: pool}
}

func (r *PostgresHandoffRepo) HasActiveHandoffRequest(ctx context.Context, conversationID string) (bool, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("handoff_requests").
		Where(sq.Eq{"conversation_id": conversationID}).
		Where(sq.Or{
			sq.Eq{"status": models.HandoffStatusPending},
			sq.Eq{"status": models.HandoffStatusAssigned},
		}).
		ToSql()
	if err != nil {
		return false, pkgerrors.Wrapf(err, "build has active handoff request query")
	}

	var count int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, pkgerrors.Wrapf(err, "has active handoff request")
	}
	return count > 0, nil
}

func (r *PostgresHandoffRepo) CreateHandoffRequest(ctx context.Context, req *models.HandoffRequest) (string, error) {
	query, args, err := psql.
		Insert("handoff_requests").
		Columns("chatbot_id", "conversation_id", "status", "notes", "user_email").
		Values(req.ChatbotID, req.ConversationID, models.HandoffStatusPending, req.Notes, req.UserEmail).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", pkgerrors.Wrapf(err, "build create handoff request query")
	}

	var id string
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "create handoff request")
	}
	return id, nil
}

func (r *PostgresHandoffRepo) GetHandoffRequestsByBotID(ctx context.Context, chatbotID string) ([]*models.HandoffRequest, error) {
	query, args, err := psql.
		Select(
			"id", "chatbot_id", "conversation_id", "status",
			"assigned_to", "notes", "user_email", "created_at", "resolved_at",
		).
		From("handoff_requests").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get handoff requests query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query handoff requests")
	}
	defer func() { _ = rows.Close() }()

	var requests []*models.HandoffRequest
	for rows.Next() {
		var req models.HandoffRequest
		if err := rows.Scan(
			&req.ID, &req.ChatbotID, &req.ConversationID, &req.Status,
			&req.AssignedTo, &req.Notes, &req.UserEmail, &req.CreatedAt, &req.ResolvedAt,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan handoff request")
		}
		requests = append(requests, &req)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "handoff requests rows error")
	}
	return requests, nil
}

func (r *PostgresHandoffRepo) GetHandoffRequestByID(ctx context.Context, id string) (*models.HandoffRequest, error) {
	query, args, err := psql.
		Select(
			"id", "chatbot_id", "conversation_id", "status",
			"assigned_to", "notes", "user_email", "created_at", "resolved_at",
		).
		From("handoff_requests").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get handoff request query")
	}

	var req models.HandoffRequest
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&req.ID, &req.ChatbotID, &req.ConversationID, &req.Status,
		&req.AssignedTo, &req.Notes, &req.UserEmail, &req.CreatedAt, &req.ResolvedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get handoff request")
	}
	return &req, nil
}

func (r *PostgresHandoffRepo) UpdateHandoffRequestStatus(ctx context.Context, id, status string, assignedTo *string) error {
	var resolvedAt interface{}
	if status == models.HandoffStatusResolved {
		now := time.Now()
		resolvedAt = &now
	}

	query, args, err := psql.
		Update("handoff_requests").
		Set("status", status).
		Set("assigned_to", assignedTo).
		Set("resolved_at", resolvedAt).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update handoff request status query")
	}

	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update handoff request status")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrapf(err, "rows affected")
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresHandoffRepo) CountPendingHandoffRequests(ctx context.Context, chatbotID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("handoff_requests").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"status": models.HandoffStatusPending}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count pending handoff requests query")
	}

	var count int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "count pending handoff requests")
	}
	return count, nil
}

func (r *PostgresHandoffRepo) ListHandoffMessages(ctx context.Context, conversationID string, limit int) ([]models.Message, error) {
	query, args, err := psql.
		Select("id", "conversation_id", "role", "content", "tokens_used", "thumbs_up", "created_at", "type").
		From("messages").
		Where(sq.Eq{"conversation_id": conversationID}).
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build list handoff messages query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query handoff messages")
	}
	defer func() { _ = rows.Close() }()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(
			&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.TokensUsed,
			&m.ThumbsUp, &m.CreatedAt, &m.Type,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan message")
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "messages rows error")
	}
	return messages, nil
}
