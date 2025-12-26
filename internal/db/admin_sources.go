package db

import (
	"context"
	"database/sql"
	"time"
)

// AdminSource represents a data source for admin views with additional metadata
type AdminSource struct {
	ID               string     `json:"id"`
	ChatbotID        string     `json:"chatbot_id"`
	ChatbotName      string     `json:"chatbot_name"`
	OrganizationName *string    `json:"organization_name,omitempty"`
	OwnerEmail       string     `json:"owner_email"`
	SourceType       string     `json:"source_type"`
	SourceURL        *string    `json:"source_url,omitempty"`
	OriginalFilename *string    `json:"original_filename,omitempty"`
	Status           string     `json:"status"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	ChunkCount       int        `json:"chunk_count"`
	SizeBytes        *int64     `json:"size_bytes,omitempty"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	CreatedAt        string     `json:"created_at"`
}

// SourceFilter contains optional filters for listing sources
type SourceFilter struct {
	ChatbotID  *string
	SourceType *string
	Status     *string
	OwnerID    *string
}

// AdminListSources returns a paginated list of all data sources with metadata
func AdminListSources(ctx context.Context, pool *sql.DB, filter SourceFilter, limit, offset int) ([]AdminSource, int, error) {
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
	err := pool.QueryRowContext(ctx, countQuery, filter.ChatbotID, filter.SourceType, filter.Status, filter.OwnerID).Scan(&total)
	if err != nil {
		return nil, 0, err
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

	rows, err := pool.QueryContext(ctx, dataQuery, filter.ChatbotID, filter.SourceType, filter.Status, filter.OwnerID, limit, offset)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
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
		return nil, 0, err
	}

	if sources == nil {
		sources = []AdminSource{}
	}

	return sources, total, nil
}

// AdminGetSourceStats returns aggregated statistics for data sources
func AdminGetSourceStats(ctx context.Context, pool *sql.DB) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*)
		FROM data_sources
		WHERE deleted_at IS NULL
		GROUP BY status
	`

	rows, err := pool.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, rows.Err()
}

// AdminReprocessSource resets a source to pending status for reprocessing
func AdminReprocessSource(ctx context.Context, pool *sql.DB, id string) error {
	query := `
		UPDATE data_sources
		SET status = 'pending', error_message = NULL, chunk_count = 0, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// AdminGetSourceByID returns a single source with all details
func AdminGetSourceByID(ctx context.Context, pool *sql.DB, id string) (*AdminSource, error) {
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
	err := pool.QueryRowContext(ctx, query, id).Scan(
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
		return nil, err
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

// AdminListSourceStatuses returns available source statuses
func AdminListSourceStatuses() []string {
	return []string{"pending", "processing", "ready", "failed"}
}

// AdminListSourceTypes returns available source types
func AdminListSourceTypes() []string {
	return []string{"url", "file", "pdf", "text"}
}
