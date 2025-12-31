package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

type AuditLogEntry struct {
	ID          string         `json:"id"`
	AdminUserID string         `json:"admin_user_id"`
	Action      string         `json:"action"`
	TargetType  string         `json:"target_type"`
	TargetID    *string        `json:"target_id"`
	Details     map[string]any `json:"details"`
	IPAddress   string         `json:"ip_address"`
	UserAgent   string         `json:"user_agent"`
	CreatedAt   time.Time      `json:"created_at"`
}

type AuditFilter struct {
	AdminUserID *string
	Action      *string
	TargetType  *string
	TargetID    *string
	StartDate   *time.Time
	EndDate     *time.Time
}

func InsertAuditLog(ctx context.Context, pool *sql.DB, entry AuditLogEntry) error {
	detailsJSON, err := json.Marshal(entry.Details)
	if err != nil {
		return pkgerrors.Wrapf(err, "marshal details")
	}

	query := `
		INSERT INTO admin_audit_logs (
			admin_user_id, action, target_type, target_id, details, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = pool.ExecContext(ctx, query,
		entry.AdminUserID,
		entry.Action,
		entry.TargetType,
		entry.TargetID,
		detailsJSON,
		entry.IPAddress,
		entry.UserAgent,
	)
	if err != nil {
		return pkgerrors.Wrapf(err, "insert audit log")
	}

	return nil
}

func ListAuditLogs(ctx context.Context, pool *sql.DB, filter AuditFilter, limit, offset int) ([]AuditLogEntry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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
		Limit(uint64(limit)).
		Offset(uint64(offset)) // #nosec G115 -- limit/offset validated to be non-negative above

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build audit logs query")
	}

	rows, err := pool.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "query audit logs")
	}
	defer func() {
		_ = rows.Close()
	}()

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
				return nil, 0, pkgerrors.Wrapf(err, "unmarshal details")
			}
		}

		logs = append(logs, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "rows error")
	}

	return logs, totalCount, nil
}
