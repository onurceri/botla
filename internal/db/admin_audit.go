package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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
	query := `
		SELECT id, admin_user_id, action, target_type, target_id, details, ip_address, user_agent, created_at, COUNT(*) OVER() as total_count
		FROM admin_audit_logs
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

	if filter.AdminUserID != nil {
		query += fmt.Sprintf(" AND admin_user_id = $%d", argIdx)
		args = append(args, *filter.AdminUserID)
		argIdx++
	}
	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *filter.Action)
		argIdx++
	}
	if filter.TargetType != nil {
		query += fmt.Sprintf(" AND target_type = $%d", argIdx)
		args = append(args, *filter.TargetType)
		argIdx++
	}
	if filter.TargetID != nil {
		query += fmt.Sprintf(" AND target_id = $%d", argIdx)
		args = append(args, *filter.TargetID)
		argIdx++
	}
	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := pool.QueryContext(ctx, query, args...)
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
