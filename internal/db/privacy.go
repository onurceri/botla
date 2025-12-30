package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

type PrivacyRequest struct {
	ID              string     `json:"id"`
	UserID          *string    `json:"user_id"`
	UserEmail       string     `json:"user_email"`
	RequestType     string     `json:"request_type"`
	Status          string     `json:"status"`
	Reason          string     `json:"reason,omitempty"`
	DenialReason    *string    `json:"denial_reason,omitempty"`
	ProcessedBy     *string    `json:"processed_by,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ExportURL       *string    `json:"export_url,omitempty"`
	ExportExpiresAt *time.Time `json:"export_expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type DataExport struct {
	ID            string     `json:"id"`
	UserID        *string    `json:"user_id"`
	RequestedBy   *string    `json:"requested_by"`
	Format        string     `json:"format"`
	Status        string     `json:"status"`
	DownloadURL   *string    `json:"download_url"`
	FileSizeBytes *int64     `json:"file_size_bytes"`
	ExpiresAt     *time.Time `json:"expires_at"`
	ErrorMessage  *string    `json:"error_message"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type UserDataExport struct {
	User          models.User                 `json:"user"`
	Organizations []models.Organization       `json:"organizations"`
	Chatbots      []models.Chatbot            `json:"chatbots"`
	Conversations []models.Conversation       `json:"conversations"`
	Messages      []models.Message            `json:"messages"`
	ActionLogs    []models.ActionExecutionLog `json:"action_logs"`
	Consents      []UserConsent               `json:"consents"`
	ExportedAt    time.Time                   `json:"exported_at"`
}

func CreateDataExport(ctx context.Context, pool *sql.DB, exp DataExport) (*DataExport, error) {
	var created DataExport
	err := pool.QueryRowContext(ctx, `
		INSERT INTO data_exports (
			user_id, requested_by, format, status
		) VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, requested_by, format, status, created_at
	`, exp.UserID, exp.RequestedBy, exp.Format, exp.Status).Scan(
		&created.ID, &created.UserID, &created.RequestedBy, &created.Format, &created.Status, &created.CreatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create data export")
	}
	return &created, nil
}

func GetDataExport(ctx context.Context, pool *sql.DB, id string) (*DataExport, error) {
	var exp DataExport
	err := pool.QueryRowContext(ctx, `
		SELECT id, user_id, requested_by, format, status, download_url, file_size_bytes, expires_at,
		       error_message, created_at, completed_at
		FROM data_exports
		WHERE id = $1
	`, id).Scan(
		&exp.ID, &exp.UserID, &exp.RequestedBy, &exp.Format, &exp.Status, &exp.DownloadURL, &exp.FileSizeBytes,
		&exp.ExpiresAt, &exp.ErrorMessage, &exp.CreatedAt, &exp.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get data export")
	}
	return &exp, nil
}

func UpdateDataExport(ctx context.Context, pool *sql.DB, exp DataExport) error {
	_, err := pool.ExecContext(ctx, `
		UPDATE data_exports
		SET status = $2, download_url = $3, file_size_bytes = $4, expires_at = $5, 
		    error_message = $6, completed_at = $7
		WHERE id = $1
	`, exp.ID, exp.Status, exp.DownloadURL, exp.FileSizeBytes, exp.ExpiresAt, exp.ErrorMessage, exp.CompletedAt)
	if err != nil {
		return pkgerrors.Wrapf(err, "update data export")
	}
	return nil
}

func CreatePrivacyRequest(ctx context.Context, pool *sql.DB, req PrivacyRequest) (*PrivacyRequest, error) {
	var created PrivacyRequest
	err := pool.QueryRowContext(ctx, `
		INSERT INTO privacy_requests (
			user_id, user_email, request_type, status, reason
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, user_email, request_type, status, reason, created_at
	`, req.UserID, req.UserEmail, req.RequestType, req.Status, req.Reason).Scan(
		&created.ID, &created.UserID, &created.UserEmail, &created.RequestType, &created.Status, &created.Reason, &created.CreatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create privacy request")
	}
	return &created, nil
}

func GetPrivacyRequest(ctx context.Context, pool *sql.DB, id string) (*PrivacyRequest, error) {
	var req PrivacyRequest
	err := pool.QueryRowContext(ctx, `
		SELECT id, user_id, user_email, request_type, status, reason, denial_reason, 
		       processed_by, processed_at, completed_at, export_url, export_expires_at, created_at
		FROM privacy_requests WHERE id = $1
	`, id).Scan(
		&req.ID, &req.UserID, &req.UserEmail, &req.RequestType, &req.Status,
		&req.Reason, &req.DenialReason, &req.ProcessedBy, &req.ProcessedAt,
		&req.CompletedAt, &req.ExportURL, &req.ExportExpiresAt, &req.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get privacy request")
	}
	return &req, nil
}

func ListPrivacyRequests(ctx context.Context, pool *sql.DB, status string, limit, offset int) ([]PrivacyRequest, int, error) {
	query := `
		SELECT id, user_id, user_email, request_type, status, reason, denial_reason,
		       processed_by, processed_at, completed_at, export_url, export_expires_at, created_at,
		       COUNT(*) OVER() as total_count
		FROM privacy_requests
	`
	args := []any{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "list privacy requests")
	}
	defer func() { _ = rows.Close() }()

	var requests []PrivacyRequest
	var totalCount int

	for rows.Next() {
		var req PrivacyRequest
		err := rows.Scan(
			&req.ID, &req.UserID, &req.UserEmail, &req.RequestType, &req.Status,
			&req.Reason, &req.DenialReason, &req.ProcessedBy, &req.ProcessedAt,
			&req.CompletedAt, &req.ExportURL, &req.ExportExpiresAt, &req.CreatedAt, &totalCount,
		)
		if err != nil {
			return nil, 0, pkgerrors.Wrapf(err, "scan privacy request")
		}
		requests = append(requests, req)
	}

	return requests, totalCount, nil
}

func UpdatePrivacyRequestStatus(ctx context.Context, pool *sql.DB, id, status, adminID string, denialReason *string) error {
	query := `
		UPDATE privacy_requests 
		SET status = $2, processed_by = $3, processed_at = NOW(), denial_reason = $4, updated_at = NOW()
	`
	args := []any{id, status, adminID, denialReason}

	if status == "completed" {
		query = `
			UPDATE privacy_requests 
			SET status = $2, processed_by = $3, processed_at = NOW(), completed_at = NOW(), updated_at = NOW()
		`
		args = []any{id, status, adminID}
	}

	query += " WHERE id = $1"

	_, err := pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update privacy request status")
	}
	return nil
}

func CompletePrivacyExportRequest(ctx context.Context, pool *sql.DB, id, adminID, exportURL string, exportExpiresAt time.Time) error {
	_, err := pool.ExecContext(ctx, `
		UPDATE privacy_requests
		SET status = 'completed',
		    processed_by = COALESCE(processed_by, $2),
		    processed_at = COALESCE(processed_at, NOW()),
		    completed_at = NOW(),
		    export_url = $3,
		    export_expires_at = $4,
		    updated_at = NOW()
		WHERE id = $1
	`, id, adminID, exportURL, exportExpiresAt)
	if err != nil {
		return pkgerrors.Wrapf(err, "complete privacy export request")
	}
	return nil
}

func AnonymizeUserData(ctx context.Context, pool *sql.DB, userID string) error {
	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		return pkgerrors.Wrapf(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	// Anonymize user fields and soft delete
	// We replace sensitive data with random/placeholder values and mark as deleted
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET full_name = 'Anonymized User',
		    email = 'anonymized-' || id || '@botla.co',
		    password_hash = 'ANONYMIZED',
		    avatar_url = NULL,
		    payment_customer_id = NULL,
		    deleted_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "anonymize user")
	}

	// Delete user's chatbots (this will cascade to data_sources, conversations, messages, etc.)
	_, err = tx.ExecContext(ctx, `DELETE FROM chatbots WHERE user_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user chatbots")
	}

	// Delete user's owned organizations (cascades to workspaces, etc.)
	_, err = tx.ExecContext(ctx, `DELETE FROM organizations WHERE owner_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user organizations")
	}

	// Delete user's refresh tokens
	_, err = tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user refresh tokens")
	}

	// Delete user's memberships
	_, err = tx.ExecContext(ctx, `DELETE FROM memberships WHERE user_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user memberships")
	}

	if err := tx.Commit(); err != nil {
		return pkgerrors.Wrapf(err, "commit tx")
	}
	return nil
}

func GetUserFilesForDeletion(ctx context.Context, pool *sql.DB, userID string) ([]string, error) {
	var files []string

	// 1. Get user avatar
	var avatarURL sql.NullString
	err := pool.QueryRowContext(ctx, `SELECT avatar_url FROM users WHERE id = $1`, userID).Scan(&avatarURL)
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgerrors.Wrapf(err, "query user avatar")
	}
	if avatarURL.Valid && avatarURL.String != "" {
		files = append(files, avatarURL.String)
	}

	// 2. Get chatbot bot_icons
	rows, err := pool.QueryContext(ctx, `SELECT bot_icon FROM chatbots WHERE user_id = $1 AND bot_icon IS NOT NULL AND bot_icon != ''`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbot icons")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var icon string
		if scanErr := rows.Scan(&icon); scanErr != nil {
			return nil, pkgerrors.Wrapf(scanErr, "scan chatbot icon")
		}
		files = append(files, icon)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, pkgerrors.Wrapf(rowsErr, "chatbot icons rows err")
	}

	// 3. Get data source file paths
	rows, err = pool.QueryContext(ctx, `
		SELECT file_path 
		FROM data_sources ds
		JOIN chatbots cb ON ds.chatbot_id = cb.id
		WHERE cb.user_id = $1 AND ds.file_path IS NOT NULL AND ds.file_path != ''
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query data source file paths")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var path string
		if scanErr := rows.Scan(&path); scanErr != nil {
			return nil, pkgerrors.Wrapf(scanErr, "scan data source file path")
		}
		files = append(files, path)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, pkgerrors.Wrapf(rowsErr, "data source file paths rows err")
	}

	// 4. Get organization branding logos
	rows, err = pool.QueryContext(ctx, `
		SELECT branding 
		FROM organizations 
		WHERE owner_id = $1 AND branding IS NOT NULL
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query organization branding")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var brandingBytes []byte
		if scanErr := rows.Scan(&brandingBytes); scanErr != nil {
			return nil, pkgerrors.Wrapf(scanErr, "scan organization branding")
		}
		if len(brandingBytes) > 0 {
			var cb struct {
				LogoURL string `json:"logo_url"`
			}
			if unmarshalErr := json.Unmarshal(brandingBytes, &cb); unmarshalErr == nil && cb.LogoURL != "" {
				files = append(files, cb.LogoURL)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "organization branding rows err")
	}

	return files, nil
}
func GetUserDataForExport(ctx context.Context, pool *sql.DB, userID string) (*UserDataExport, error) {
	export := &UserDataExport{
		ExportedAt: time.Now(),
	}

	// 1. Get User
	user, err := GetUserByID(ctx, pool, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get user")
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	export.User = *user

	// 2. Get Organizations
	rows, err := pool.QueryContext(ctx, `
		SELECT o.id, o.name, o.slug, o.owner_id, o.plan_id, o.created_at, o.updated_at
		FROM organizations o
		JOIN memberships m ON o.id = m.organization_id
		WHERE m.user_id = $1
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query organizations")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var org models.Organization
		err = rows.Scan(
			&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.PlanID,
			&org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "scan organization")
		}
		export.Organizations = append(export.Organizations, org)
	}

	// 3. Get Chatbots
	chatbots, err := GetChatbotsByUserID(ctx, pool, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get chatbots")
	}
	export.Chatbots = chatbots

	// 4. Get Conversations and Messages
	// We get conversations for all chatbots owned by the user
	convRows, err := pool.QueryContext(ctx, `
		SELECT c.id, c.chatbot_id, c.session_id, c.message_count, c.created_at, c.updated_at
		FROM conversations c
		JOIN chatbots b ON c.chatbot_id = b.id
		WHERE b.user_id = $1
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query conversations")
	}
	defer func() { _ = convRows.Close() }()

	var conversationIDs []string
	for convRows.Next() {
		var conv models.Conversation
		err = convRows.Scan(
			&conv.ID, &conv.ChatbotID, &conv.SessionID, &conv.MessageCount,
			&conv.CreatedAt, &conv.UpdatedAt,
		)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "scan conversation")
		}
		export.Conversations = append(export.Conversations, conv)
		conversationIDs = append(conversationIDs, conv.ID)
	}

	// 5. Get Messages for those conversations
	if len(conversationIDs) > 0 {
		// Using a simple loop or IN clause. For export, a single query with ANY is efficient.
		msgRows, queryErr := pool.QueryContext(ctx, `
			SELECT id, conversation_id, role, content, tokens_used, thumbs_up, created_at
			FROM messages
			WHERE conversation_id = ANY($1)
			ORDER BY created_at ASC
		`, conversationIDs)
		if queryErr != nil {
			return nil, pkgerrors.Wrapf(queryErr, "query messages")
		}
		defer func() { _ = msgRows.Close() }()

		for msgRows.Next() {
			var msg models.Message
			err = msgRows.Scan(
				&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
				&msg.TokensUsed, &msg.ThumbsUp, &msg.CreatedAt,
			)
			if err != nil {
				return nil, pkgerrors.Wrapf(err, "scan message")
			}
			export.Messages = append(export.Messages, msg)
		}
	}

	// 6. Get Action Logs
	actionRows, err := pool.QueryContext(ctx, `
		SELECT l.id, l.chatbot_id, l.action_id, l.conversation_id, l.message_id,
		       l.status, l.request_payload, l.response_payload, l.error_message,
		       l.duration_ms, l.created_at
		FROM action_execution_logs l
		JOIN chatbots b ON l.chatbot_id = b.id
		WHERE b.user_id = $1
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query action logs")
	}
	defer func() { _ = actionRows.Close() }()

	for actionRows.Next() {
		var l models.ActionExecutionLog
		err = actionRows.Scan(
			&l.ID, &l.ChatbotID, &l.ActionID, &l.ConversationID, &l.MessageID,
			&l.Status, &l.RequestPayload, &l.ResponsePayload, &l.ErrorMessage,
			&l.DurationMs, &l.CreatedAt,
		)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "scan action log")
		}
		export.ActionLogs = append(export.ActionLogs, l)
	}

	// 7. Get Consents
	consents, err := GetUserConsents(ctx, pool, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get consents")
	}
	export.Consents = consents

	return export, nil
}
