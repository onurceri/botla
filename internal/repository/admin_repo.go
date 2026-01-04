// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresAdminRepo implements AdminRepository using PostgreSQL.
type PostgresAdminRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresAdminRepo implements AdminRepository.
var _ AdminRepository = (*PostgresAdminRepo)(nil)

// NewPostgresAdminRepo creates a new PostgresAdminRepo instance.
func NewPostgresAdminRepo(pool *sql.DB) *PostgresAdminRepo {
	return &PostgresAdminRepo{pool: pool}
}

// InsertAuditLog persists a new audit log entry.
func (r *PostgresAdminRepo) InsertAuditLog(ctx context.Context, entry AuditLogEntry) error {
	detailsJSON, err := json.Marshal(entry.Details)
	if err != nil {
		return pkgerrors.Wrapf(err, "marshal audit log details")
	}

	query, args, err := psql.
		Insert("admin_audit_logs").
		Columns("admin_user_id", "action", "target_type", "target_id", "details", "ip_address", "user_agent").
		Values(entry.AdminUserID, entry.Action, entry.TargetType, entry.TargetID, detailsJSON, entry.IPAddress, entry.UserAgent).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build insert audit log query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "insert audit log")
	}

	return nil
}

// ListAuditLogs returns a paginated list of audit logs with optional filtering.
func (r *PostgresAdminRepo) ListAuditLogs(ctx context.Context, filter AuditFilter, limit, offset int) ([]AuditLogEntry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := psql.Select(
		"id", "admin_user_id", "action", "target_type", "target_id",
		"details", "ip_address", "user_agent", "created_at",
		"COUNT(*) OVER() as total_count",
	).From("admin_audit_logs")

	if filter.AdminUserID != nil {
		query = query.Where(sq.Eq{"admin_user_id": *filter.AdminUserID})
	}
	if filter.Action != nil {
		query = query.Where(sq.Eq{"action": *filter.Action})
	}
	if filter.TargetType != nil {
		query = query.Where(sq.Eq{"target_type": *filter.TargetType})
	}
	if filter.TargetID != nil {
		query = query.Where(sq.Eq{"target_id": *filter.TargetID})
	}
	if filter.StartDate != nil {
		query = query.Where(sq.GtOrEq{"created_at": *filter.StartDate})
	}
	if filter.EndDate != nil {
		query = query.Where(sq.LtOrEq{"created_at": *filter.EndDate})
	}

	query = query.OrderBy("created_at DESC").
		Limit(uint64(limit)).   // #nosec G115 -- limit validated above
		Offset(uint64(offset)) // #nosec G115 -- offset validated above

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build audit logs query")
	}

	rows, err := r.pool.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "query audit logs")
	}
	defer func() { _ = rows.Close() }()

	var logs []AuditLogEntry
	var totalCount int

	for rows.Next() {
		var entry AuditLogEntry
		var detailsJSON []byte
		err := rows.Scan(
			&entry.ID,
			&entry.AdminUserID,
			&entry.Action,
			&entry.TargetType,
			&entry.TargetID,
			&detailsJSON,
			&entry.IPAddress,
			&entry.UserAgent,
			&entry.CreatedAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, pkgerrors.Wrapf(err, "scan audit log")
		}

		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &entry.Details); err != nil {
				return nil, 0, pkgerrors.Wrapf(err, "unmarshal audit log details")
			}
		}

		logs = append(logs, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "audit logs rows error")
	}

	return logs, totalCount, nil
}

// AdminListSources returns a paginated list of all data sources with metadata.
func (r *PostgresAdminRepo) AdminListSources(ctx context.Context, filter AdminSourceFilter, limit, offset int) ([]AdminSource, int, error) {
	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM data_sources ds
		JOIN chatbots c ON ds.chatbot_id = c.id
		WHERE ds.deleted_at IS NULL
			AND ($1::text IS NULL OR ds.chatbot_id = $1::uuid)
			AND ($2::text IS NULL OR ds.source_type = $2)
			AND ($3::text IS NULL OR ds.status = $3)
			AND ($4::text IS NULL OR c.user_id = $4::uuid)
	`

	var total int
	err := r.pool.QueryRowContext(ctx, countQuery, filter.ChatbotID, filter.SourceType, filter.Status, filter.OwnerID).Scan(&total)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "count admin sources")
	}

	// Data query
	dataQuery := `
		SELECT
			ds.id,
			ds.chatbot_id,
			c.name,
			o.name,
			u.email,
			ds.source_type,
			ds.source_url,
			ds.original_filename,
			ds.status,
			ds.error_message,
			COALESCE(ds.chunk_count, 0),
			ds.size_bytes,
			ds.processed_at,
			ds.created_at
		FROM data_sources ds
		JOIN chatbots c ON ds.chatbot_id = c.id
		LEFT JOIN organizations o ON c.organization_id = o.id
		JOIN users u ON c.user_id = u.id
		WHERE ds.deleted_at IS NULL
			AND ($1::text IS NULL OR ds.chatbot_id = $1::uuid)
			AND ($2::text IS NULL OR ds.source_type = $2)
			AND ($3::text IS NULL OR ds.status = $3)
			AND ($4::text IS NULL OR c.user_id = $4::uuid)
		ORDER BY ds.created_at DESC
		LIMIT $5 OFFSET $6
	`

	rows, err := r.pool.QueryContext(ctx, dataQuery, filter.ChatbotID, filter.SourceType, filter.Status, filter.OwnerID, limit, offset)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "query admin sources")
	}
	defer func() { _ = rows.Close() }()

	var sources []AdminSource
	for rows.Next() {
		var s AdminSource
		var orgName, sourceURL, origFilename, errMsg sql.NullString
		var sizeBytes sql.NullInt64
		var processedAt sql.NullTime
		var createdAt time.Time
		err := rows.Scan(
			&s.ID,
			&s.ChatbotID,
			&s.ChatbotName,
			&orgName,
			&s.OwnerEmail,
			&s.SourceType,
			&sourceURL,
			&origFilename,
			&s.Status,
			&errMsg,
			&s.ChunkCount,
			&sizeBytes,
			&processedAt,
			&createdAt,
		)
		if err != nil {
			return nil, 0, pkgerrors.Wrapf(err, "scan admin source")
		}
		if orgName.Valid {
			s.OrganizationName = &orgName.String
		}
		if sourceURL.Valid {
			s.SourceURL = &sourceURL.String
		}
		if origFilename.Valid {
			s.OriginalFilename = &origFilename.String
		}
		if errMsg.Valid {
			s.ErrorMessage = &errMsg.String
		}
		if sizeBytes.Valid {
			s.SizeBytes = &sizeBytes.Int64
		}
		if processedAt.Valid {
			s.ProcessedAt = &processedAt.Time
		}
		s.CreatedAt = createdAt.Format(time.RFC3339)
		sources = append(sources, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "admin sources rows error")
	}

	if sources == nil {
		sources = []AdminSource{}
	}

	return sources, total, nil
}

// AdminGetSourceByID retrieves a single source by ID with all admin-visible details.
func (r *PostgresAdminRepo) AdminGetSourceByID(ctx context.Context, id string) (*AdminSource, error) {
	query := `
		SELECT
			ds.id,
			ds.chatbot_id,
			c.name,
			o.name,
			u.email,
			ds.source_type,
			ds.source_url,
			ds.original_filename,
			ds.status,
			ds.error_message,
			COALESCE(ds.chunk_count, 0),
			ds.size_bytes,
			ds.processed_at,
			ds.created_at
		FROM data_sources ds
		JOIN chatbots c ON ds.chatbot_id = c.id
		LEFT JOIN organizations o ON c.organization_id = o.id
		JOIN users u ON c.user_id = u.id
		WHERE ds.id = $1 AND ds.deleted_at IS NULL
	`

	var s AdminSource
	var orgName, sourceURL, origFilename, errMsg sql.NullString
	var sizeBytes sql.NullInt64
	var processedAt sql.NullTime
	var createdAt time.Time
	err := r.pool.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.ChatbotID,
		&s.ChatbotName,
		&orgName,
		&s.OwnerEmail,
		&s.SourceType,
		&sourceURL,
		&origFilename,
		&s.Status,
		&errMsg,
		&s.ChunkCount,
		&sizeBytes,
		&processedAt,
		&createdAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get admin source by id")
	}
	if orgName.Valid {
		s.OrganizationName = &orgName.String
	}
	if sourceURL.Valid {
		s.SourceURL = &sourceURL.String
	}
	if origFilename.Valid {
		s.OriginalFilename = &origFilename.String
	}
	if errMsg.Valid {
		s.ErrorMessage = &errMsg.String
	}
	if sizeBytes.Valid {
		s.SizeBytes = &sizeBytes.Int64
	}
	if processedAt.Valid {
		s.ProcessedAt = &processedAt.Time
	}
	s.CreatedAt = createdAt.Format(time.RFC3339)

	return &s, nil
}

// AdminGetSourceStats returns aggregated statistics for data sources.
func (r *PostgresAdminRepo) AdminGetSourceStats(ctx context.Context) (*SourceStats, error) {
	query := `
		SELECT status, COUNT(*)
		FROM data_sources
		WHERE deleted_at IS NULL
		GROUP BY status
	`

	rows, err := r.pool.QueryContext(ctx, query)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query source stats")
	}
	defer func() { _ = rows.Close() }()

	stats := &SourceStats{
		StatusCounts: make(map[string]int),
	}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan source stat")
		}
		stats.StatusCounts[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "source stats rows error")
	}
	return stats, nil
}

// AdminReprocessSource resets a source to pending status for reprocessing.
func (r *PostgresAdminRepo) AdminReprocessSource(ctx context.Context, id string) error {
	query, args, err := psql.
		Update("data_sources").
		Set("status", "pending").
		Set("error_message", nil).
		Set("chunk_count", 0).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build reprocess source query")
	}

	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "reprocess source")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrapf(err, "rows affected")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ListErrorLogs returns a paginated list of error logs with optional severity filtering.
func (r *PostgresAdminRepo) ListErrorLogs(ctx context.Context, severity string, limit, offset int) ([]ErrorLogEntry, int, error) {
	query := `
		SELECT id, error_type, message, stack_trace, request_path, request_method, 
		       user_id, chatbot_id, organization_id, severity, context, created_at,
		       COUNT(*) OVER() as total_count
		FROM error_logs
		WHERE ($1 = '' OR severity = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.QueryContext(ctx, query, severity, limit, offset)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "query error logs")
	}
	defer func() { _ = rows.Close() }()

	var logs []ErrorLogEntry
	total := 0
	for rows.Next() {
		var l ErrorLogEntry
		var stack, path, method sql.NullString
		var userID, botID, orgID sql.NullString
		var ctxData []byte
		err := rows.Scan(
			&l.ID, &l.ErrorType, &l.Message, &stack, &path, &method,
			&userID, &botID, &orgID, &l.Severity, &ctxData, &l.CreatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, pkgerrors.Wrapf(err, "scan error log")
		}
		if stack.Valid {
			l.StackTrace = stack.String
		}
		if path.Valid {
			l.RequestPath = path.String
		}
		if method.Valid {
			l.RequestMethod = method.String
		}
		if userID.Valid {
			s := userID.String
			l.UserID = &s
		}
		if botID.Valid {
			s := botID.String
			l.ChatbotID = &s
		}
		if orgID.Valid {
			s := orgID.String
			l.OrganizationID = &s
		}
		l.Context = ctxData
		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "error logs rows error")
	}

	return logs, total, nil
}

// GetErrorLogByID retrieves a single error log entry by ID.
func (r *PostgresAdminRepo) GetErrorLogByID(ctx context.Context, id string) (*ErrorLogEntry, error) {
	var l ErrorLogEntry
	var stack, path, method sql.NullString
	var userID, botID, orgID sql.NullString
	var ctxData []byte

	err := r.pool.QueryRowContext(ctx, `
		SELECT id, error_type, message, stack_trace, request_path, request_method, 
		       user_id, chatbot_id, organization_id, severity, context, created_at
		FROM error_logs WHERE id = $1
	`, id).Scan(
		&l.ID, &l.ErrorType, &l.Message, &stack, &path, &method,
		&userID, &botID, &orgID, &l.Severity, &ctxData, &l.CreatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get error log")
	}

	if stack.Valid {
		l.StackTrace = stack.String
	}
	if path.Valid {
		l.RequestPath = path.String
	}
	if method.Valid {
		l.RequestMethod = method.String
	}
	if userID.Valid {
		s := userID.String
		l.UserID = &s
	}
	if botID.Valid {
		s := botID.String
		l.ChatbotID = &s
	}
	if orgID.Valid {
		s := orgID.String
		l.OrganizationID = &s
	}
	l.Context = ctxData

	return &l, nil
}

// GetErrorStats returns aggregated error statistics for the last 24 hours.
func (r *PostgresAdminRepo) GetErrorStats(ctx context.Context) (*ErrorStats, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT severity, COUNT(*) as count
		FROM error_logs
		WHERE created_at > NOW() - INTERVAL '24 hours'
		GROUP BY severity
	`)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query error stats")
	}
	defer func() { _ = rows.Close() }()

	stats := &ErrorStats{
		SeverityCounts: make(map[string]int),
	}
	for rows.Next() {
		var sev string
		var count int
		if err := rows.Scan(&sev, &count); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan error stat")
		}
		stats.SeverityCounts[sev] = count
	}

	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "error stats rows error")
	}

	return stats, nil
}
