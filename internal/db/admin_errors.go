package db

import (
	"context"
	"database/sql"
	"time"
)

type ErrorLogEntry struct {
	ID             string    `json:"id"`
	ErrorType      string    `json:"error_type"`
	Message        string    `json:"message"`
	StackTrace     string    `json:"stack_trace,omitempty"`
	RequestPath    string    `json:"request_path,omitempty"`
	RequestMethod  string    `json:"request_method,omitempty"`
	UserID         *string   `json:"user_id,omitempty"`
	ChatbotID      *string   `json:"chatbot_id,omitempty"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	Severity       string    `json:"severity"`
	Context        []byte    `json:"context,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ListErrorLogs returns a paginated list of error logs with optional severity filtering.
func ListErrorLogs(ctx context.Context, pool *sql.DB, severity string, limit, offset int) ([]ErrorLogEntry, int, error) {
	query := `
		SELECT id, error_type, message, stack_trace, request_path, request_method, 
		       user_id, chatbot_id, organization_id, severity, context, created_at,
		       COUNT(*) OVER() as total_count
		FROM error_logs
		WHERE ($1 = '' OR severity = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := pool.QueryContext(ctx, query, severity, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

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
			return nil, 0, err
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

	return logs, total, nil
}

// GetErrorLogByID returns details for a single error log entry.
func GetErrorLogByID(ctx context.Context, pool *sql.DB, id string) (*ErrorLogEntry, error) {
	var l ErrorLogEntry
	var stack, path, method sql.NullString
	var userID, botID, orgID sql.NullString
	var ctxData []byte

	err := pool.QueryRowContext(ctx, `
		SELECT id, error_type, message, stack_trace, request_path, request_method, 
		       user_id, chatbot_id, organization_id, severity, context, created_at
		FROM error_logs WHERE id = $1
	`, id).Scan(
		&l.ID, &l.ErrorType, &l.Message, &stack, &path, &method,
		&userID, &botID, &orgID, &l.Severity, &ctxData, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
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

// GetErrorStats returns counts of errors by severity for the last 24 hours.
func GetErrorStats(ctx context.Context, pool *sql.DB) (map[string]int, error) {
	rows, err := pool.QueryContext(ctx, `
		SELECT severity, COUNT(*) as count
		FROM error_logs
		WHERE created_at > NOW() - INTERVAL '24 hours'
		GROUP BY severity
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var sev string
		var count int
		if err := rows.Scan(&sev, &count); err != nil {
			return nil, err
		}
		stats[sev] = count
	}
	return stats, nil
}
