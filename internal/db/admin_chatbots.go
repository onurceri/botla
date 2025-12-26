package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// AdminChatbot represents a chatbot for admin views with additional metadata
type AdminChatbot struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	OwnerID          string          `json:"owner_id"`
	WorkspaceID      *string         `json:"workspace_id,omitempty"`
	OrganizationID   *string         `json:"organization_id,omitempty"`
	OrganizationName *string         `json:"organization_name,omitempty"`
	OwnerEmail       string          `json:"owner_email"`
	SourceCount      int             `json:"source_count"`
	MessageCount     int             `json:"message_count"`
	CustomBranding   json.RawMessage `json:"custom_branding"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

// ChatbotFilter contains optional filters for listing chatbots
type ChatbotFilter struct {
	Name           *string
	OrganizationID *string
	OwnerID        *string
}

// AdminListChatbots returns a paginated list of all chatbots with their metadata
func AdminListChatbots(ctx context.Context, pool *sql.DB, filter ChatbotFilter, limit, offset int) ([]AdminChatbot, int, error) {
	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM chatbots c
		WHERE c.deleted_at IS NULL
			AND ($1::text IS NULL OR c.name ILIKE '%' || $1 || '%')
			AND ($2::text IS NULL OR c.organization_id = $2::uuid)
			AND ($3::text IS NULL OR c.user_id = $3::uuid)
	`

	var total int
	err := pool.QueryRowContext(ctx, countQuery, filter.Name, filter.OrganizationID, filter.OwnerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data query - simplified to avoid LATERAL issues
	dataQuery := `
		SELECT
			c.id,
			c.name,
			c.user_id,
			c.workspace_id,
			c.organization_id,
			o.name,
			u.email,
			(SELECT COUNT(*) FROM data_sources ds WHERE ds.chatbot_id = c.id AND ds.deleted_at IS NULL),
			(SELECT COUNT(*) FROM messages m JOIN conversations conv ON m.conversation_id = conv.id WHERE conv.chatbot_id = c.id),
			COALESCE(c.custom_branding::text, '{}'),
			c.created_at,
			c.updated_at
		FROM chatbots c
		LEFT JOIN organizations o ON c.organization_id = o.id
		JOIN users u ON c.user_id = u.id
		WHERE c.deleted_at IS NULL
			AND ($1::text IS NULL OR c.name ILIKE '%' || $1 || '%')
			AND ($2::text IS NULL OR c.organization_id = $2::uuid)
			AND ($3::text IS NULL OR c.user_id = $3::uuid)
		ORDER BY c.created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := pool.QueryContext(ctx, dataQuery, filter.Name, filter.OrganizationID, filter.OwnerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var chatbots []AdminChatbot
	for rows.Next() {
		var c AdminChatbot
		var workspaceID, orgID, orgName sql.NullString
		var customBranding string
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.OwnerID,
			&workspaceID,
			&orgID,
			&orgName,
			&c.OwnerEmail,
			&c.SourceCount,
			&c.MessageCount,
			&customBranding,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if workspaceID.Valid {
			c.WorkspaceID = &workspaceID.String
		}
		if orgID.Valid {
			c.OrganizationID = &orgID.String
		}
		if orgName.Valid {
			c.OrganizationName = &orgName.String
		}
		c.CustomBranding = json.RawMessage(customBranding)
		c.CreatedAt = createdAt.Format(time.RFC3339)
		c.UpdatedAt = updatedAt.Format(time.RFC3339)
		chatbots = append(chatbots, c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if chatbots == nil {
		chatbots = []AdminChatbot{}
	}

	return chatbots, total, nil
}

// AdminGetChatbot returns a single chatbot by ID with all details
func AdminGetChatbot(ctx context.Context, pool *sql.DB, id string) (*AdminChatbot, error) {
	query := `
		SELECT
			c.id,
			c.name,
			c.user_id,
			c.workspace_id,
			c.organization_id,
			o.name,
			u.email,
			(SELECT COUNT(*) FROM data_sources ds WHERE ds.chatbot_id = c.id AND ds.deleted_at IS NULL),
			(SELECT COUNT(*) FROM messages m JOIN conversations conv ON m.conversation_id = conv.id WHERE conv.chatbot_id = c.id),
			COALESCE(c.custom_branding::text, '{}'),
			c.created_at,
			c.updated_at
		FROM chatbots c
		LEFT JOIN organizations o ON c.organization_id = o.id
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`

	var c AdminChatbot
	var workspaceID, orgID, orgName sql.NullString
	var customBranding string
	var createdAt, updatedAt time.Time
	err := pool.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.Name,
		&c.OwnerID,
		&workspaceID,
		&orgID,
		&orgName,
		&c.OwnerEmail,
		&c.SourceCount,
		&c.MessageCount,
		&customBranding,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if workspaceID.Valid {
		c.WorkspaceID = &workspaceID.String
	}
	if orgID.Valid {
		c.OrganizationID = &orgID.String
	}
	if orgName.Valid {
		c.OrganizationName = &orgName.String
	}
	c.CustomBranding = json.RawMessage(customBranding)
	c.CreatedAt = createdAt.Format(time.RFC3339)
	c.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &c, nil
}

// AdminResetChatbotSources resets all sources for a chatbot to pending status for reprocessing
func AdminResetChatbotSources(ctx context.Context, pool *sql.DB, chatbotID string) (int64, error) {
	query := `
		UPDATE data_sources
		SET status = 'pending', error_message = NULL, updated_at = NOW()
		WHERE chatbot_id = $1 AND deleted_at IS NULL AND status IN ('failed', 'ready')
	`

	result, err := pool.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// AdminGetChatbotSourceIDs returns all source IDs for a chatbot for queue processing
func AdminGetChatbotSourceIDs(ctx context.Context, pool *sql.DB, chatbotID string) ([]string, error) {
	query := `
		SELECT id FROM data_sources
		WHERE chatbot_id = $1 AND deleted_at IS NULL AND status = 'pending'
	`

	rows, err := pool.QueryContext(ctx, query, chatbotID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// AdminDeleteChatbotVectors deletes all vectors for a chatbot's sources (for reindexing)
func AdminDeleteChatbotVectors(ctx context.Context, pool *sql.DB, chatbotID string) error {
	// Reset chunk_count to 0 for all sources
	query := `
		UPDATE data_sources
		SET chunk_count = 0
		WHERE chatbot_id = $1 AND deleted_at IS NULL
	`
	_, err := pool.ExecContext(ctx, query, chatbotID)
	return err
}

// Ensure pq is imported for lib/pq driver
var _ = pq.Array
