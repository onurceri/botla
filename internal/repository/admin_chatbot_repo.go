// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresAdminChatbotRepo implements AdminChatbotRepository using PostgreSQL.
// It uses Squirrel SQL builder for type-safe query construction.
type PostgresAdminChatbotRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresAdminChatbotRepo implements AdminChatbotRepository.
var _ AdminChatbotRepository = (*PostgresAdminChatbotRepo)(nil)

// NewPostgresAdminChatbotRepo creates a new PostgresAdminChatbotRepo instance.
func NewPostgresAdminChatbotRepo(pool *sql.DB) *PostgresAdminChatbotRepo {
	return &PostgresAdminChatbotRepo{pool: pool}
}

// Pool returns the underlying database connection pool for testing purposes.
func (r *PostgresAdminChatbotRepo) Pool() *sql.DB {
	return r.pool
}

// ListChatbots returns a paginated list of all chatbots with their metadata.
func (r *PostgresAdminChatbotRepo) ListChatbots(ctx context.Context, filter AdminChatbotFilter, limit, offset int) ([]AdminChatbot, int, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

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
	err = r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total)
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

	rows, err := r.pool.QueryContext(ctx, dataSQL, dataArgs...)
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
		c.CustomBranding = []byte(customBranding)
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

// GetByID retrieves a single chatbot by ID with all admin-visible details.
func (r *PostgresAdminChatbotRepo) GetByID(ctx context.Context, id string) (*AdminChatbot, error) {
	query := psql.Select(
		"c.id",
		"c.name",
		"c.user_id",
		"c.workspace_id",
		"c.organization_id",
		"o.name",
		"u.email",
		"(SELECT COUNT(*) FROM data_sources ds WHERE ds.chatbot_id = c.id AND ds.deleted_at IS NULL)",
		"(SELECT COUNT(*) FROM messages m JOIN conversations conv ON m.conversation_id = conv.id WHERE conv.chatbot_id = c.id)",
		"COALESCE(c.custom_branding::text, '{}')",
		"c.created_at",
		"c.updated_at",
	).From("chatbots c").
		LeftJoin("organizations o ON c.organization_id = o.id").
		Join("users u ON c.user_id = u.id").
		Where(sq.Eq{"c.id": id, "c.deleted_at": nil})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get chatbot query")
	}

	var c AdminChatbot
	var workspaceID, orgID, orgName sql.NullString
	var customBranding string
	var createdAt, updatedAt time.Time
	err = r.pool.QueryRowContext(ctx, sqlQuery, args...).Scan(
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
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
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
	c.CustomBranding = []byte(customBranding)
	c.CreatedAt = createdAt.Format(time.RFC3339)
	c.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &c, nil
}

// ResetSources resets all sources for a chatbot to pending status for reprocessing.
func (r *PostgresAdminChatbotRepo) ResetSources(ctx context.Context, chatbotID string) (int64, error) {
	query := `
		UPDATE data_sources
		SET status = 'pending', error_message = NULL
		WHERE chatbot_id = $1 AND deleted_at IS NULL AND status IN ('failed', 'ready')
	`

	result, err := r.pool.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "reset chatbot sources")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "rows affected")
	}
	return rowsAffected, nil
}

// GetSourceIDs returns all pending source IDs for a chatbot for queue processing.
func (r *PostgresAdminChatbotRepo) GetSourceIDs(ctx context.Context, chatbotID string) ([]string, error) {
	query := `
		SELECT id FROM data_sources
		WHERE chatbot_id = $1 AND deleted_at IS NULL AND status = 'pending'
	`

	rows, err := r.pool.QueryContext(ctx, query, chatbotID)
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

// DeleteVectors resets chunk counts to 0 for all sources (for reindexing).
func (r *PostgresAdminChatbotRepo) DeleteVectors(ctx context.Context, chatbotID string) error {
	// Reset chunk_count to 0 for all sources
	query := `
		UPDATE data_sources
		SET chunk_count = 0
		WHERE chatbot_id = $1 AND deleted_at IS NULL
	`
	_, err := r.pool.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete chatbot vectors")
	}
	return nil
}
