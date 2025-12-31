package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
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
	// Validate pagination parameters
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Configure Squirrel for PostgreSQL (use $1, $2... placeholders)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note: limit and offset are validated to be non-negative above, so uint64 conversion is safe

	// Build COUNT query
	countQuery := psql.Select("COUNT(*)").From("chatbots c").Where(sq.Eq{"c.deleted_at": nil})

	// Apply optional filters
	if filter.Name != nil {
		namePattern := "%" + *filter.Name + "%"
		countQuery = countQuery.Where("c.name ILIKE ?", namePattern)
	}
	if filter.OrganizationID != nil {
		countQuery = countQuery.Where(sq.Eq{"c.organization_id": *filter.OrganizationID})
	}
	if filter.OwnerID != nil {
		countQuery = countQuery.Where(sq.Eq{"c.user_id": *filter.OwnerID})
	}

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build count query")
	}

	var total int
	err = pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "count admin chatbots")
	}

	// Build DATA query with subqueries
	dataQuery := psql.Select(
		"c.id",
		"c.name",
		"c.user_id",
		"c.workspace_id",
		"c.organization_id",
		"o.name as organization_name",
		"u.email as owner_email",
		"(SELECT COUNT(*) FROM data_sources ds WHERE ds.chatbot_id = c.id AND ds.deleted_at IS NULL)",
		"(SELECT COUNT(*) FROM messages m JOIN conversations conv ON m.conversation_id = conv.id WHERE conv.chatbot_id = c.id)",
		"COALESCE(c.custom_branding::text, '{}')",
		"c.created_at",
		"c.updated_at",
	).From("chatbots c").
		LeftJoin("organizations o ON c.organization_id = o.id").
		Join("users u ON c.user_id = u.id").
		Where(sq.Eq{"c.deleted_at": nil})

	// Apply same filters as count query
	if filter.Name != nil {
		namePattern := "%" + *filter.Name + "%"
		dataQuery = dataQuery.Where("c.name ILIKE ?", namePattern)
	}
	if filter.OrganizationID != nil {
		dataQuery = dataQuery.Where(sq.Eq{"c.organization_id": *filter.OrganizationID})
	}
	if filter.OwnerID != nil {
		dataQuery = dataQuery.Where(sq.Eq{"c.user_id": *filter.OwnerID})
	}

	// Sorting and pagination
	dataQuery = dataQuery.OrderBy("c.created_at DESC").Limit(uint64(int64(limit))).Offset(uint64(int64(offset))) // #nosec G115 -- limit/offset validated to be non-negative above

	dataSQL, dataArgs, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build data query")
	}

	rows, err := pool.QueryContext(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "query admin chatbots")
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
			return nil, 0, pkgerrors.Wrapf(err, "scan admin chatbot")
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
		return nil, 0, pkgerrors.Wrapf(err, "admin chatbots rows err")
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
		return nil, pkgerrors.Wrapf(err, "get admin chatbot")
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
		return 0, pkgerrors.Wrapf(err, "reset chatbot sources")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "rows affected")
	}
	return rowsAffected, nil
}

// AdminGetChatbotSourceIDs returns all source IDs for a chatbot for queue processing
func AdminGetChatbotSourceIDs(ctx context.Context, pool *sql.DB, chatbotID string) ([]string, error) {
	query := `
		SELECT id FROM data_sources
		WHERE chatbot_id = $1 AND deleted_at IS NULL AND status = 'pending'
	`

	rows, err := pool.QueryContext(ctx, query, chatbotID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query source ids")
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan source id")
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "source ids rows err")
	}
	return ids, nil
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
	if err != nil {
		return pkgerrors.Wrapf(err, "delete chatbot vectors")
	}
	return nil
}

// Ensure pq is imported for lib/pq driver
var _ = pq.Array
