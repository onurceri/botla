package db

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/models"
)

func GetOrCreateConversationBySessionID(ctx context.Context, pool *sql.DB, chatbotID string, sessionID string) (*models.Conversation, error) {
	var c models.Conversation
	err := pool.QueryRowContext(ctx, `
        SELECT id, chatbot_id, session_id, message_count, created_at, updated_at
        FROM conversations
        WHERE chatbot_id=$1 AND session_id=$2`, chatbotID, sessionID).Scan(
		&c.ID, &c.ChatbotID, &c.SessionID, &c.MessageCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		// create new conversation
		var id string
		err = pool.QueryRowContext(ctx, `
            INSERT INTO conversations (chatbot_id, session_id)
            VALUES ($1, $2) RETURNING id`, chatbotID, sessionID).Scan(&id)
		if err != nil {
			return nil, err
		}
		// re-read to fill timestamps
		err = pool.QueryRowContext(ctx, `
            SELECT id, chatbot_id, session_id, message_count, created_at, updated_at
            FROM conversations WHERE id=$1`, id).Scan(
			&c.ID, &c.ChatbotID, &c.SessionID, &c.MessageCount, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}
	return &c, nil
}

func CreateMessage(ctx context.Context, pool *sql.DB, m *models.Message) (string, error) {
	var id string
	err := pool.QueryRowContext(ctx, `
        INSERT INTO messages (
            conversation_id, role, content, tokens_used, thumbs_up
        ) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		m.ConversationID, m.Role, m.Content, m.TokensUsed, m.ThumbsUp,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func IncrementConversationMessageCount(ctx context.Context, pool *sql.DB, conversationID string) error {
	_, err := pool.ExecContext(ctx, `
        UPDATE conversations SET message_count = message_count + 1, updated_at=NOW()
        WHERE id=$1`, conversationID)
	return err
}

func ListRecentMessages(ctx context.Context, pool *sql.DB, conversationID string, limit int) ([]models.Message, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT id, conversation_id, role, content, tokens_used, thumbs_up, created_at
        FROM messages
        WHERE conversation_id=$1
        ORDER BY created_at DESC
        LIMIT $2`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.TokensUsed, &m.ThumbsUp, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// reverse to chronological ascending
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func UpdateMessageFeedback(ctx context.Context, pool *sql.DB, messageID string, thumbsUp bool) (string, error) {
	var chatbotID string
	// We need chatbot_id for analytics update, so we join tables
	err := pool.QueryRowContext(ctx, `
		UPDATE messages SET thumbs_up=$2
		WHERE id=$1
		RETURNING (
			SELECT c.chatbot_id 
			FROM conversations c 
			WHERE c.id = messages.conversation_id
		)
	`, messageID, thumbsUp).Scan(&chatbotID)
	return chatbotID, err
}
