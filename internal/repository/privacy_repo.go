// Package repository provides data access layer implementations for privacy operations.
package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PrivacyRequest represents a user privacy request (deletion, export, correction).
type PrivacyRequest struct {
	ID              string     `json:"id"`
	UserID          *string    `json:"user_id,omitempty"`
	UserEmail       string     `json:"user_email"`
	RequestType     string     `json:"request_type"` // "deletion", "export", "correction"
	Status          string     `json:"status"`       // "pending", "processing", "completed", "denied"
	Reason          string     `json:"reason,omitempty"`
	DenialReason    *string    `json:"denial_reason,omitempty"`
	ProcessedBy     *string    `json:"processed_by,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ExportURL       *string    `json:"export_url,omitempty"`
	ExportExpiresAt *time.Time `json:"export_expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// DataExport represents a data export record.
type DataExport struct {
	ID            string     `json:"id"`
	UserID        *string    `json:"user_id,omitempty"`
	RequestedBy   *string    `json:"requested_by,omitempty"`
	Format        string     `json:"format"` // "json"
	Status        string     `json:"status"` // "pending", "processing", "completed", "failed"
	DownloadURL   *string    `json:"download_url,omitempty"`
	FileSizeBytes *int64     `json:"file_size_bytes,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// UserDataExport represents all user data for export.
type UserDataExport struct {
	User          ExportUser        `json:"user"`
	Organizations []ExportOrg       `json:"organizations"`
	Chatbots      []ExportChatbot   `json:"chatbots"`
	Conversations []ExportConv      `json:"conversations"`
	Messages      []ExportMessage   `json:"messages"`
	ActionLogs    []ExportActionLog `json:"action_logs"`
	Consents      []ExportConsent   `json:"consents"`
	ExportedAt    time.Time         `json:"exported_at"`
}

// ExportUser contains user data for GDPR export (excludes internal fields).
type ExportUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  *string   `json:"full_name,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	PlanName  *string   `json:"plan_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ExportOrg contains organization data for GDPR export.
type ExportOrg struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ExportChatbot contains chatbot data for GDPR export (excludes style/config).
type ExportChatbot struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	Sources   []ExportSource `json:"sources,omitempty"`
}

// ExportSource contains data source info for GDPR export.
type ExportSource struct {
	Type    string    `json:"type"`
	Name    string    `json:"name,omitempty"`
	URL     string    `json:"url,omitempty"`
	AddedAt time.Time `json:"added_at"`
}

// ExportConv contains conversation data for GDPR export.
type ExportConv struct {
	ID        string    `json:"id"`
	ChatbotID string    `json:"chatbot_id"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ExportMessage contains message data for GDPR export.
type ExportMessage struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

// ExportActionLog contains action execution log for GDPR export.
type ExportActionLog struct {
	ID        string    `json:"id"`
	ChatbotID string    `json:"chatbot_id"`
	ActionID  string    `json:"action_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ExportConsent contains consent data for GDPR export.
type ExportConsent struct {
	ConsentType string     `json:"consent_type"`
	Granted     bool       `json:"granted"`
	GrantedAt   time.Time  `json:"granted_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

// UserConsent represents user consent records.
type UserConsent struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	ConsentType string     `json:"consent_type"` // "privacy", "terms", "marketing"
	Granted     bool       `json:"granted"`
	GrantedAt   time.Time  `json:"granted_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PostgresPrivacyRepo implements PrivacyRepository using PostgreSQL.
type PostgresPrivacyRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresPrivacyRepo implements PrivacyRepository.
var _ PrivacyRepository = (*PostgresPrivacyRepo)(nil)

// NewPostgresPrivacyRepo creates a new PostgresPrivacyRepo instance.
func NewPostgresPrivacyRepo(pool *sql.DB) *PostgresPrivacyRepo {
	return &PostgresPrivacyRepo{pool: pool}
}

// GetUserConsents retrieves all consent records for a user.
func (r *PostgresPrivacyRepo) GetUserConsents(ctx context.Context, userID string) ([]UserConsent, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, user_id, consent_type, granted, granted_at, revoked_at
		FROM user_consents
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get user consents")
	}
	defer func() { _ = rows.Close() }()

	var consents []UserConsent
	for rows.Next() {
		var c UserConsent
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ConsentType, &c.Granted, &c.GrantedAt, &c.RevokedAt,
		)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "scan user consent")
		}
		consents = append(consents, c)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "user consents rows error")
	}
	return consents, nil
}

// UpsertConsent creates or updates a consent record.
func (r *PostgresPrivacyRepo) UpsertConsent(ctx context.Context, userID string, consentType string, granted bool, ipAddress, userAgent string) error {
	var revokedAt *time.Time
	if !granted {
		now := time.Now()
		revokedAt = &now
	}

	query, args, err := psql.
		Insert("user_consents").
		Columns("user_id", "consent_type", "granted", "ip_address", "user_agent", "revoked_at").
		Values(userID, consentType, granted, ipAddress, userAgent, revokedAt).
		Suffix(`
			ON CONFLICT (user_id, consent_type)
			DO UPDATE SET
				granted = EXCLUDED.granted,
				ip_address = EXCLUDED.ip_address,
				user_agent = EXCLUDED.user_agent,
				revoked_at = EXCLUDED.revoked_at,
				granted_at = CASE WHEN EXCLUDED.granted THEN NOW() ELSE user_consents.granted_at END
		`).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build upsert consent query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "upsert consent")
	}
	return nil
}

// CreateDataExport creates a new data export record.
func (r *PostgresPrivacyRepo) CreateDataExport(ctx context.Context, exp DataExport) (*DataExport, error) {
	query, args, err := psql.
		Insert("data_exports").
		Columns("user_id", "requested_by", "format", "status").
		Values(exp.UserID, exp.RequestedBy, exp.Format, exp.Status).
		Suffix("RETURNING id, user_id, requested_by, format, status, created_at").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build create data export query")
	}

	var created DataExport
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&created.ID, &created.UserID, &created.RequestedBy, &created.Format, &created.Status, &created.CreatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create data export")
	}
	return &created, nil
}

// UpdateDataExport updates an existing data export record.
func (r *PostgresPrivacyRepo) UpdateDataExport(ctx context.Context, exp DataExport) error {
	query, args, err := psql.
		Update("data_exports").
		Set("status", exp.Status).
		Set("download_url", exp.DownloadURL).
		Set("file_size_bytes", exp.FileSizeBytes).
		Set("expires_at", exp.ExpiresAt).
		Set("error_message", exp.ErrorMessage).
		Set("completed_at", exp.CompletedAt).
		Where(sq.Eq{"id": exp.ID}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update data export query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update data export")
	}
	return nil
}

// GetDataExport retrieves a data export by ID.
func (r *PostgresPrivacyRepo) GetDataExport(ctx context.Context, id string) (*DataExport, error) {
	query, args, err := psql.
		Select(
			"id", "user_id", "requested_by", "format", "status",
			"download_url", "file_size_bytes", "expires_at", "error_message",
			"created_at", "completed_at",
		).
		From("data_exports").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get data export query")
	}

	var exp DataExport
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&exp.ID, &exp.UserID, &exp.RequestedBy, &exp.Format, &exp.Status,
		&exp.DownloadURL, &exp.FileSizeBytes, &exp.ExpiresAt, &exp.ErrorMessage,
		&exp.CreatedAt, &exp.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get data export")
	}
	return &exp, nil
}

// CreatePrivacyRequest creates a new privacy request.
func (r *PostgresPrivacyRepo) CreatePrivacyRequest(ctx context.Context, req PrivacyRequest) (*PrivacyRequest, error) {
	query, args, err := psql.
		Insert("privacy_requests").
		Columns("user_id", "user_email", "request_type", "status", "reason").
		Values(req.UserID, req.UserEmail, req.RequestType, req.Status, req.Reason).
		Suffix("RETURNING id, user_id, user_email, request_type, status, reason, created_at").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build create privacy request query")
	}

	var created PrivacyRequest
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&created.ID, &created.UserID, &created.UserEmail, &created.RequestType,
		&created.Status, &created.Reason, &created.CreatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create privacy request")
	}
	return &created, nil
}

// GetPrivacyRequest retrieves a privacy request by ID.
func (r *PostgresPrivacyRepo) GetPrivacyRequest(ctx context.Context, requestID string) (*PrivacyRequest, error) {
	query, args, err := psql.
		Select(
			"id", "user_id", "user_email", "request_type", "status", "reason",
			"denial_reason", "processed_by", "processed_at", "completed_at",
			"export_url", "export_expires_at", "created_at",
		).
		From("privacy_requests").
		Where(sq.Eq{"id": requestID}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get privacy request query")
	}

	var req PrivacyRequest
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
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

// ListPrivacyRequests retrieves privacy requests with optional status filter and pagination.
func (r *PostgresPrivacyRepo) ListPrivacyRequests(ctx context.Context, status string, limit, offset int) ([]PrivacyRequest, int, error) {
	baseQuery := psql.
		Select(
			"id", "user_id", "user_email", "request_type", "status", "reason",
			"denial_reason", "processed_by", "processed_at", "completed_at",
			"export_url", "export_expires_at", "created_at",
		).
		From("privacy_requests")

	var whereClause sq.Eq
	if status != "" {
		whereClause = sq.Eq{"status": status}
	}

	query, args, err := baseQuery.
		Where(whereClause).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build list privacy requests query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "list privacy requests")
	}
	defer func() { _ = rows.Close() }()

	var requests []PrivacyRequest
	for rows.Next() {
		var req PrivacyRequest
		err := rows.Scan(
			&req.ID, &req.UserID, &req.UserEmail, &req.RequestType, &req.Status,
			&req.Reason, &req.DenialReason, &req.ProcessedBy, &req.ProcessedAt,
			&req.CompletedAt, &req.ExportURL, &req.ExportExpiresAt, &req.CreatedAt,
		)
		if err != nil {
			return nil, 0, pkgerrors.Wrapf(err, "scan privacy request")
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "privacy requests rows error")
	}

	// Get total count
	countQuery, countArgs, err := psql.
		Select("COUNT(*)").
		From("privacy_requests").
		Where(whereClause).
		ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build count query")
	}

	var totalCount int
	err = r.pool.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "count privacy requests")
	}

	return requests, totalCount, nil
}

// UpdatePrivacyRequestStatus updates the status of a privacy request.
func (r *PostgresPrivacyRepo) UpdatePrivacyRequestStatus(ctx context.Context, requestID, status, adminID string, denialReason *string) error {
	var query string
	var args []interface{}

	if status == "completed" {
		query = `
			UPDATE privacy_requests
			SET status = $2, processed_by = $3, processed_at = NOW(), completed_at = NOW(), updated_at = NOW()
			WHERE id = $1
		`
		args = []interface{}{requestID, status, adminID}
	} else {
		query = `
			UPDATE privacy_requests
			SET status = $2, processed_by = $3, processed_at = NOW(), denial_reason = $4, updated_at = NOW()
			WHERE id = $1
		`
		args = []interface{}{requestID, status, adminID, denialReason}
	}

	_, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update privacy request status")
	}
	return nil
}

// CompletePrivacyExportRequest marks a privacy export request as completed.
func (r *PostgresPrivacyRepo) CompletePrivacyExportRequest(ctx context.Context, requestID, adminID, exportURL string, expiresAt time.Time) error {
	query, args, err := psql.
		Update("privacy_requests").
		Set("status", "completed").
		Set("processed_by", sq.Expr("COALESCE(processed_by, ?)", adminID)).
		Set("processed_at", sq.Expr("COALESCE(processed_at, NOW())")).
		Set("completed_at", sq.Expr("NOW()")).
		Set("export_url", exportURL).
		Set("export_expires_at", expiresAt).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": requestID}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build complete privacy export request query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "complete privacy export request")
	}
	return nil
}

// AnonymizeUserData anonymizes a user's personal data and deletes their content.
func (r *PostgresPrivacyRepo) AnonymizeUserData(ctx context.Context, userID string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return pkgerrors.Wrapf(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	// Anonymize user fields
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

	// Delete user's chatbots
	_, err = tx.ExecContext(ctx, `DELETE FROM chatbots WHERE user_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user chatbots")
	}

	// Delete user's organizations
	_, err = tx.ExecContext(ctx, `DELETE FROM organizations WHERE owner_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user organizations")
	}

	// Delete refresh tokens
	_, err = tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user refresh tokens")
	}

	// Delete memberships
	_, err = tx.ExecContext(ctx, `DELETE FROM memberships WHERE user_id = $1`, userID)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete user memberships")
	}

	if err := tx.Commit(); err != nil {
		return pkgerrors.Wrapf(err, "commit tx")
	}
	return nil
}

// GetUserFilesForDeletion returns file paths that should be deleted from storage.
func (r *PostgresPrivacyRepo) GetUserFilesForDeletion(ctx context.Context, userID string) ([]string, error) {
	var files []string

	// Get user avatar
	var avatarURL sql.NullString
	err := r.pool.QueryRowContext(ctx, `SELECT avatar_url FROM users WHERE id = $1`, userID).Scan(&avatarURL)
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgerrors.Wrapf(err, "query user avatar")
	}
	if avatarURL.Valid && avatarURL.String != "" {
		files = append(files, avatarURL.String)
	}

	// Get chatbot bot_icons
	rows, err := r.pool.QueryContext(ctx, `SELECT bot_icon FROM chatbots WHERE user_id = $1 AND bot_icon IS NOT NULL AND bot_icon != ''`, userID)
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
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "chatbot icons rows error")
	}

	return files, nil
}

// GetUserDataForExport retrieves all user data for GDPR export.
func (r *PostgresPrivacyRepo) GetUserDataForExport(ctx context.Context, userID string) (*UserDataExport, error) {
	export := &UserDataExport{
		Organizations: []ExportOrg{},
		Chatbots:      []ExportChatbot{},
		Conversations: []ExportConv{},
		Messages:      []ExportMessage{},
		ActionLogs:    []ExportActionLog{},
		Consents:      []ExportConsent{},
		ExportedAt:    time.Now(),
	}

	// 1. Get user with plan name
	userQuery, userArgs, err := psql.
		Select("u.id", "u.email", "u.full_name", "u.avatar_url", "p.code", "u.created_at").
		From("users u").
		LeftJoin("plans p ON p.id = u.plan_id").
		Where(sq.Eq{"u.id": userID}).
		Where(sq.Eq{"u.deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build user query")
	}
	err = r.pool.QueryRowContext(ctx, userQuery, userArgs...).Scan(
		&export.User.ID, &export.User.Email, &export.User.FullName, &export.User.AvatarURL,
		&export.User.PlanName, &export.User.CreatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgerrors.Wrapf(err, "get user")
	}

	// 2. Get organizations via memberships
	orgsQuery, orgsArgs, err := psql.
		Select("o.id", "o.name", "o.created_at").
		From("organizations o").
		Join("memberships m ON m.organization_id = o.id").
		Where(sq.Eq{"m.user_id": userID}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build orgs query")
	}
	orgsRows, err := r.pool.QueryContext(ctx, orgsQuery, orgsArgs...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query organizations")
	}
	defer func() { _ = orgsRows.Close() }()
	for orgsRows.Next() {
		var org ExportOrg
		if err := orgsRows.Scan(&org.ID, &org.Name, &org.CreatedAt); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan organization")
		}
		export.Organizations = append(export.Organizations, org)
	}

	// 3. Get chatbots
	botsQuery, botsArgs, err := psql.
		Select("id", "name", "created_at").
		From("chatbots").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build chatbots query")
	}
	botsRows, err := r.pool.QueryContext(ctx, botsQuery, botsArgs...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbots")
	}
	defer func() { _ = botsRows.Close() }()
	var chatbotIDs []string
	chatbotMap := make(map[string]int)
	for botsRows.Next() {
		var bot ExportChatbot
		if err := botsRows.Scan(&bot.ID, &bot.Name, &bot.CreatedAt); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan chatbot")
		}
		chatbotMap[bot.ID] = len(export.Chatbots)
		export.Chatbots = append(export.Chatbots, bot)
		chatbotIDs = append(chatbotIDs, bot.ID)
	}

	// 4. Get data sources for chatbots
	if len(chatbotIDs) > 0 {
		sourcesQuery, sourcesArgs, err := psql.
			Select("chatbot_id", "source_type", "original_filename", "source_url", "created_at").
			From("data_sources").
			Where(sq.Eq{"chatbot_id": chatbotIDs}).
			Where(sq.Eq{"deleted_at": nil}).
			ToSql()
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "build sources query")
		}
		sourcesRows, err := r.pool.QueryContext(ctx, sourcesQuery, sourcesArgs...)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "query sources")
		}
		defer func() { _ = sourcesRows.Close() }()
		for sourcesRows.Next() {
			var chatbotID string
			var source ExportSource
			var filename, url *string
			if err := sourcesRows.Scan(&chatbotID, &source.Type, &filename, &url, &source.AddedAt); err != nil {
				return nil, pkgerrors.Wrapf(err, "scan source")
			}
			if filename != nil {
				source.Name = *filename
			}
			if url != nil {
				source.URL = *url
			}
			if idx, ok := chatbotMap[chatbotID]; ok {
				export.Chatbots[idx].Sources = append(export.Chatbots[idx].Sources, source)
			}
		}
	}

	// 5. Get conversations for user's chatbots
	if len(chatbotIDs) > 0 {
		convsQuery, convsArgs, err := psql.
			Select("id", "chatbot_id", "session_id", "created_at").
			From("conversations").
			Where(sq.Eq{"chatbot_id": chatbotIDs}).
			ToSql()
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "build conversations query")
		}
		convsRows, err := r.pool.QueryContext(ctx, convsQuery, convsArgs...)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "query conversations")
		}
		defer func() { _ = convsRows.Close() }()
		var convIDs []string
		for convsRows.Next() {
			var conv ExportConv
			if err := convsRows.Scan(&conv.ID, &conv.ChatbotID, &conv.SessionID, &conv.CreatedAt); err != nil {
				return nil, pkgerrors.Wrapf(err, "scan conversation")
			}
			export.Conversations = append(export.Conversations, conv)
			convIDs = append(convIDs, conv.ID)
		}

		// 6. Get messages
		if len(convIDs) > 0 {
			msgsQuery, msgsArgs, err := psql.
				Select("id", "conversation_id", "role", "content", "created_at").
				From("messages").
				Where(sq.Eq{"conversation_id": convIDs}).
				OrderBy("created_at ASC").
				ToSql()
			if err != nil {
				return nil, pkgerrors.Wrapf(err, "build messages query")
			}
			msgsRows, err := r.pool.QueryContext(ctx, msgsQuery, msgsArgs...)
			if err != nil {
				return nil, pkgerrors.Wrapf(err, "query messages")
			}
			defer func() { _ = msgsRows.Close() }()
			for msgsRows.Next() {
				var msg ExportMessage
				if err := msgsRows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
					return nil, pkgerrors.Wrapf(err, "scan message")
				}
				export.Messages = append(export.Messages, msg)
			}
		}

		// 7. Get action execution logs
		logsQuery, logsArgs, err := psql.
			Select("id", "chatbot_id", "action_id", "status", "created_at").
			From("action_execution_logs").
			Where(sq.Eq{"chatbot_id": chatbotIDs}).
			ToSql()
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "build action logs query")
		}
		logsRows, err := r.pool.QueryContext(ctx, logsQuery, logsArgs...)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "query action logs")
		}
		defer func() { _ = logsRows.Close() }()
		for logsRows.Next() {
			var log ExportActionLog
			if err := logsRows.Scan(&log.ID, &log.ChatbotID, &log.ActionID, &log.Status, &log.CreatedAt); err != nil {
				return nil, pkgerrors.Wrapf(err, "scan action log")
			}
			export.ActionLogs = append(export.ActionLogs, log)
		}
	}

	// 8. Get user consents
	consentsQuery, consentsArgs, err := psql.
		Select("consent_type", "granted", "granted_at", "revoked_at").
		From("user_consents").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build consents query")
	}
	consentsRows, err := r.pool.QueryContext(ctx, consentsQuery, consentsArgs...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query consents")
	}
	defer func() { _ = consentsRows.Close() }()
	for consentsRows.Next() {
		var consent ExportConsent
		if err := consentsRows.Scan(&consent.ConsentType, &consent.Granted,
			&consent.GrantedAt, &consent.RevokedAt); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan consent")
		}
		export.Consents = append(export.Consents, consent)
	}

	return export, nil
}

// ListPrivacyRequestsByUserID retrieves privacy requests for a specific user with pagination and optional request type filter.
func (r *PostgresPrivacyRepo) ListPrivacyRequestsByUserID(ctx context.Context, userID, requestType string, limit, offset int) ([]PrivacyRequest, int, error) {
	var whereClause sq.Sqlizer = sq.Eq{"user_id": userID}
	if requestType != "" {
		whereClause = sq.And{whereClause, sq.Eq{"request_type": requestType}}
	}

	query, args, err := psql.
		Select(
			"id", "user_id", "user_email", "request_type", "status", "reason",
			"denial_reason", "processed_by", "processed_at", "completed_at",
			"export_url", "export_expires_at", "created_at",
		).
		From("privacy_requests").
		Where(whereClause).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build list privacy requests by user query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "list privacy requests by user")
	}
	defer func() { _ = rows.Close() }()

	var requests []PrivacyRequest
	for rows.Next() {
		var req PrivacyRequest
		err := rows.Scan(
			&req.ID, &req.UserID, &req.UserEmail, &req.RequestType, &req.Status,
			&req.Reason, &req.DenialReason, &req.ProcessedBy, &req.ProcessedAt,
			&req.CompletedAt, &req.ExportURL, &req.ExportExpiresAt, &req.CreatedAt,
		)
		if err != nil {
			return nil, 0, pkgerrors.Wrapf(err, "scan privacy request")
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "privacy requests rows error")
	}

	// Get total count
	countQuery, countArgs, err := psql.
		Select("COUNT(*)").
		From("privacy_requests").
		Where(whereClause).
		ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build count query")
	}

	var totalCount int
	err = r.pool.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "count privacy requests by user")
	}

	return requests, totalCount, nil
}

// HasActivePrivacyRequest checks if user has a pending or processing request of the given type.
func (r *PostgresPrivacyRepo) HasActivePrivacyRequest(ctx context.Context, userID, requestType string) (bool, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("privacy_requests").
		Where(sq.And{
			sq.Eq{"user_id": userID},
			sq.Eq{"request_type": requestType},
			sq.Or{
				sq.Eq{"status": "pending"},
				sq.Eq{"status": "processing"},
			},
		}).
		ToSql()
	if err != nil {
		return false, pkgerrors.Wrapf(err, "build has active privacy request query")
	}

	var count int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, pkgerrors.Wrapf(err, "has active privacy request")
	}

	return count > 0, nil
}

// DeletePrivacyRequest deletes a privacy request by ID.
func (r *PostgresPrivacyRepo) DeletePrivacyRequest(ctx context.Context, requestID string) error {
	query, args, err := psql.
		Delete("privacy_requests").
		Where(sq.Eq{"id": requestID}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build delete privacy request query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete privacy request")
	}
	return nil
}

// GetLastCompletedRequestDate returns the completion date of the last completed request of the given type for a user.
func (r *PostgresPrivacyRepo) GetLastCompletedRequestDate(ctx context.Context, userID, requestType string) (*time.Time, error) {
	query, args, err := psql.
		Select("completed_at").
		From("privacy_requests").
		Where(sq.And{
			sq.Eq{"user_id": userID},
			sq.Eq{"request_type": requestType},
			sq.Eq{"status": "completed"},
		}).
		OrderBy("completed_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get last completed request date query")
	}

	var completedAt *time.Time
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&completedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get last completed request date")
	}

	return completedAt, nil
}
