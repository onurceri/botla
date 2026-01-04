// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresConversationRepo implements ConversationRepository using PostgreSQL.
type PostgresConversationRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresConversationRepo implements ConversationRepository.
var _ ConversationRepository = (*PostgresConversationRepo)(nil)

// NewPostgresConversationRepo creates a new PostgresConversationRepo instance.
func NewPostgresConversationRepo(pool *sql.DB) *PostgresConversationRepo {
	return &PostgresConversationRepo{pool: pool}
}

// Pool returns the underlying database connection pool.
// Used by integration tests to set up fixtures.
func (r *PostgresConversationRepo) Pool() *sql.DB {
	return r.pool
}

// GetOrCreateBySessionID finds an existing conversation or creates a new one.
// Uses session_id as the unique identifier within a chatbot.
// The UPSERT pattern prevents race conditions when concurrent requests arrive for the same session.
func (r *PostgresConversationRepo) GetOrCreateBySessionID(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error) {
	query, args, err := psql.
		Insert("conversations").
		Columns("chatbot_id", "session_id").
		Values(chatbotID, sessionID).
		Suffix("ON CONFLICT (chatbot_id, session_id) DO UPDATE SET updated_at = NOW()").
		Suffix("RETURNING id, chatbot_id, session_id, message_count, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get or create conversation query")
	}

	var c models.Conversation
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&c.ID, &c.ChatbotID, &c.SessionID, &c.MessageCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get or create conversation")
	}
	return &c, nil
}

// GetByID retrieves a conversation by its unique identifier.
func (r *PostgresConversationRepo) GetByID(ctx context.Context, id string) (*models.Conversation, error) {
	query, args, err := psql.
		Select("id", "chatbot_id", "session_id", "message_count", "created_at", "updated_at").
		From("conversations").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get conversation by id query")
	}

	var c models.Conversation
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&c.ID, &c.ChatbotID, &c.SessionID, &c.MessageCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get conversation by id")
	}
	return &c, nil
}

// CreateMessage persists a new message in a conversation.
// Returns the generated message ID.
func (r *PostgresConversationRepo) CreateMessage(ctx context.Context, msg *models.Message) (string, error) {
	query, args, err := psql.
		Insert("messages").
		Columns("conversation_id", "role", "content", "tokens_used", "thumbs_up", "type").
		Values(msg.ConversationID, msg.Role, msg.Content, msg.TokensUsed, msg.ThumbsUp, msg.Type).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", pkgerrors.Wrapf(err, "build create message query")
	}

	var id string
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "create message")
	}
	return id, nil
}

// GetMessages retrieves messages for a conversation with pagination.
// Messages are ordered by created_at ascending (chronological order).
func (r *PostgresConversationRepo) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error) {
	query, args, err := psql.
		Select("id", "conversation_id", "role", "content", "tokens_used", "thumbs_up", "created_at", "type").
		From("messages").
		Where(sq.Eq{"conversation_id": conversationID}).
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get messages query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query messages")
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

// IncrementMessageCount atomically increments the message count for a conversation.
func (r *PostgresConversationRepo) IncrementMessageCount(ctx context.Context, conversationID string) error {
	query, args, err := psql.
		Update("conversations").
		Set("message_count", sq.Expr("message_count + 1")).
		Where(sq.Eq{"id": conversationID}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build increment message count query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment message count")
	}
	return nil
}

// ListRecentMessages retrieves recent messages for a conversation.
func (r *PostgresConversationRepo) ListRecentMessages(ctx context.Context, conversationID string, limit int) ([]models.Message, error) {
	query, args, err := psql.
		Select("id", "conversation_id", "role", "content", "tokens_used", "thumbs_up", "created_at", "type").
		From("messages").
		Where(sq.Eq{"conversation_id": conversationID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build list recent messages query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query recent messages")
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
		return nil, pkgerrors.Wrapf(err, "recent messages rows error")
	}
	return messages, nil
}

// SaveMessageSources persists source usage for a message.
func (r *PostgresConversationRepo) SaveMessageSources(ctx context.Context, messageID string, sources []models.ChunkMetadata) error {
	if len(sources) == 0 {
		return nil
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return pkgerrors.Wrapf(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO message_sources (message_id, source_id, chunk_index, source_type, relevance_score)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (message_id, source_id, chunk_index) DO NOTHING
    `)
	if err != nil {
		return pkgerrors.Wrapf(err, "prepare stmt")
	}
	defer func() { _ = stmt.Close() }()

	for _, src := range sources {
		if src.SourceID == "" {
			continue
		}
		_, err = stmt.ExecContext(ctx, messageID, src.SourceID, src.ChunkIndex, src.SourceType, src.Score)
		if err != nil {
			return pkgerrors.Wrapf(err, "exec stmt")
		}
	}

	if err := tx.Commit(); err != nil {
		return pkgerrors.Wrapf(err, "commit tx")
	}
	return nil
}
