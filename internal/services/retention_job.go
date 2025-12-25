package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

type RetentionConfig struct {
	ConversationRetentionDays int // Default: 90
	ErrorLogRetentionDays     int // Default: 30
	AuditLogRetentionDays     int // Default: 365
	DataExportRetentionDays   int // Default: 7
}

type RetentionJob struct {
	DB      *sql.DB
	Log     *logger.Logger
	Storage storage.StorageService
	Config  RetentionConfig
}

func NewRetentionJob(db *sql.DB, log *logger.Logger, storage storage.StorageService) *RetentionJob {
	return &RetentionJob{
		DB:      db,
		Log:     log,
		Storage: storage,
		Config: RetentionConfig{
			ConversationRetentionDays: 730, // 2 years as per policy
			ErrorLogRetentionDays:     30,
			AuditLogRetentionDays:     730, // 2 years as per policy
			DataExportRetentionDays:   7,
		},
	}
}

// Run executes the retention cleanup
func (j *RetentionJob) Run(ctx context.Context) error {
	j.Log.Info("starting_retention_cleanup", nil)

	// 1. Delete old conversations
	if count, err := j.cleanConversations(ctx); err != nil {
		j.Log.Error("failed_to_clean_conversations", map[string]any{"error": err.Error()})
	} else {
		j.Log.Info("cleaned_conversations", map[string]any{"count": count})
	}

	// 2. Delete old error logs
	if count, err := j.cleanErrorLogs(ctx); err != nil {
		j.Log.Error("failed_to_clean_error_logs", map[string]any{"error": err.Error()})
	} else {
		j.Log.Info("cleaned_error_logs", map[string]any{"count": count})
	}

	// 3. Delete expired data exports
	if count, err := j.cleanExpiredExports(ctx); err != nil {
		j.Log.Error("failed_to_clean_expired_exports", map[string]any{"error": err.Error()})
	} else {
		j.Log.Info("cleaned_expired_exports", map[string]any{"count": count})
	}

	// 4. Delete old audit logs
	if count, err := j.cleanAuditLogs(ctx); err != nil {
		j.Log.Error("failed_to_clean_audit_logs", map[string]any{"error": err.Error()})
	} else {
		j.Log.Info("cleaned_audit_logs", map[string]any{"count": count})
	}

	return nil
}

func (j *RetentionJob) cleanConversations(ctx context.Context) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -j.Config.ConversationRetentionDays)

	// Assuming conversations table has created_at
	query := `DELETE FROM conversations WHERE created_at < $1`

	result, err := j.DB.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (j *RetentionJob) cleanErrorLogs(ctx context.Context) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -j.Config.ErrorLogRetentionDays)

	// Assuming error_logs table exists
	query := `DELETE FROM error_logs WHERE created_at < $1`

	result, err := j.DB.ExecContext(ctx, query, cutoff)
	if err != nil {
		// If table doesn't exist, ignore error or log it.
		// For now, return error so we know.
		return 0, err
	}

	return result.RowsAffected()
}

func (j *RetentionJob) cleanExpiredExports(ctx context.Context) (int64, error) {
	now := time.Now()
	var totalDeleted int64

	if j.Storage != nil {
		// 1. Clean data_exports table files
		rows, err := j.DB.QueryContext(ctx, `SELECT download_url FROM data_exports WHERE expires_at < $1 AND download_url IS NOT NULL`, now)
		if err == nil {
			defer func() {
				closeErr := rows.Close()
				if closeErr != nil {
					j.Log.Error("failed_to_close_rows", map[string]any{"error": closeErr.Error()})
				}
			}()
			for rows.Next() {
				var key string
				if scanErr := rows.Scan(&key); scanErr == nil && key != "" {
					_ = j.Storage.DeleteFile(ctx, key)
				}
			}
		}

		// 2. Clean privacy_requests table files
		rows2, err := j.DB.QueryContext(ctx, `SELECT export_url FROM privacy_requests WHERE export_expires_at < $1 AND export_url IS NOT NULL`, now)
		if err == nil {
			defer func() {
				closeErr := rows2.Close()
				if closeErr != nil {
					j.Log.Error("failed_to_close_rows2", map[string]any{"error": closeErr.Error()})
				}
			}()
			for rows2.Next() {
				var key string
				if scanErr := rows2.Scan(&key); scanErr == nil && key != "" {
					_ = j.Storage.DeleteFile(ctx, key)
				}
			}
		}
	}

	// 3. Clear URLs from privacy_requests (don't delete the request itself, just the sensitive URL)
	res1, err := j.DB.ExecContext(ctx, `UPDATE privacy_requests SET export_url = NULL WHERE export_expires_at < $1`, now)
	if err == nil {
		affected, _ := res1.RowsAffected()
		totalDeleted += affected
	}

	// 4. Delete from data_exports table
	res2, err := j.DB.ExecContext(ctx, `DELETE FROM data_exports WHERE expires_at < $1`, now)
	if err != nil {
		return totalDeleted, err
	}
	affected, _ := res2.RowsAffected()
	totalDeleted += affected

	return totalDeleted, nil
}

func (j *RetentionJob) cleanAuditLogs(ctx context.Context) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -j.Config.AuditLogRetentionDays)

	query := `DELETE FROM admin_audit_logs WHERE created_at < $1`

	result, err := j.DB.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
