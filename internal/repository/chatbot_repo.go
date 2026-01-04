// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresChatbotRepo implements ChatbotRepository using PostgreSQL.
// SQL queries are built using Squirrel for type safety and maintainability.
type PostgresChatbotRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresChatbotRepo implements ChatbotRepository.
var _ ChatbotRepository = (*PostgresChatbotRepo)(nil)

// NewPostgresChatbotRepo creates a new PostgresChatbotRepo instance.
func NewPostgresChatbotRepo(pool *sql.DB) *PostgresChatbotRepo {
	return &PostgresChatbotRepo{pool: pool}
}

// normalizeLocale normalizes language codes to full locale format.
func (r *PostgresChatbotRepo) normalizeLocale(code string) string {
	s := strings.TrimSpace(code)
	if s == "" {
		return "tr-TR"
	}
	switch s {
	case "tr":
		return "tr-TR"
	case "en":
		return "en-US"
	}
	return s
}

// scanChatbot scans a single chatbot row from the result set.
func (r *PostgresChatbotRepo) scanChatbot(rows *sql.Rows) (*models.Chatbot, error) {
	var c models.Chatbot
	var sj, mqj, ipj, epj, swj, cbj, tcj, fmj, trj, hcj []byte
	if err := rows.Scan(
		&c.ID, &c.UserID, &c.WorkspaceID, &c.OrganizationID, &c.Name, &c.Description, &c.CustomInstruction, &c.LanguageCode, &c.Model,
		&c.Temperature, &c.MaxTokens, &c.ThemeColor, &c.WelcomeMessage,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		&c.Position, &c.BotMessageColor, &c.UserMessageColor,
		&c.BotMessageTextColor, &c.UserMessageTextColor,
		&c.ChatFontFamily, &c.ChatHeaderColor, &c.ChatHeaderTextColor,
		&c.ChatBackgroundColor, &c.BubbleRadius, &c.InputBackgroundColor, &c.InputTextColor, &c.SendButtonColor,
		&c.BotIcon, &c.BotDisplayName, &c.AllowedDomains, &c.EmbedSecret, &c.SecureEmbedEnabled,
		&sj, &mqj, &c.SuggestionsEnabled,
		&ipj, &epj, &swj, &c.DiscoveryMode,
		&c.RefreshPolicy, &c.RefreshFrequency, &c.NextRefreshAt, &c.LastRefreshAt,
		&c.HideBranding, &cbj,
		&c.ConfidenceThreshold, &tcj, &fmj, &trj,
		&c.HandoffEnabled, &c.HandoffType, &hcj,
	); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan chatbot")
	}

	// Unmarshal JSONB fields
	if len(sj) > 0 {
		var arr []string
		_ = json.Unmarshal(sj, &arr)
		c.SuggestedQuestions = arr
	}
	if len(mqj) > 0 {
		var arr []string
		_ = json.Unmarshal(mqj, &arr)
		c.ManualQuestions = arr
	}
	if len(ipj) > 0 {
		var arr []string
		_ = json.Unmarshal(ipj, &arr)
		c.IncludePaths = arr
	}
	if len(epj) > 0 {
		var arr []string
		_ = json.Unmarshal(epj, &arr)
		c.ExcludePaths = arr
	}
	if len(swj) > 0 {
		var arr []string
		_ = json.Unmarshal(swj, &arr)
		c.SelectorWhitelist = arr
	}
	if len(cbj) > 0 {
		var cb models.CustomBranding
		_ = json.Unmarshal(cbj, &cb)
		c.CustomBranding = &cb
	}
	if len(tcj) > 0 {
		var tc models.ThresholdConfig
		_ = json.Unmarshal(tcj, &tc)
		c.ThresholdConfig = &tc
	}
	if len(fmj) > 0 {
		var fm models.FallbackMessages
		_ = json.Unmarshal(fmj, &fm)
		c.FallbackMessages = &fm
	}
	if len(trj) > 0 {
		var tr models.TopicConfig
		_ = json.Unmarshal(trj, &tr)
		c.TopicRestrictions = &tr
	}
	if len(hcj) > 0 {
		var hc models.HandoffConfig
		_ = json.Unmarshal(hcj, &hc)
		c.HandoffConfig = &hc
	}
	return &c, nil
}

// scanChatbots scans multiple chatbot rows from the result set.
func (r *PostgresChatbotRepo) scanChatbots(rows *sql.Rows) ([]models.Chatbot, error) {
	var out []models.Chatbot
	for rows.Next() {
		c, err := r.scanChatbot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan chatbots rows")
	}
	return out, nil
}

// chatbotColumns returns the list of columns for SELECT queries.
func (r *PostgresChatbotRepo) chatbotColumns() []string {
	return []string{
		"c.id", "c.user_id", "c.workspace_id", "c.organization_id", "c.name", "c.description",
		"COALESCE(c.custom_instruction, '')", "COALESCE(l.code, '')", "c.model",
		"c.temperature", "c.max_tokens", "c.theme_color", "c.welcome_message",
		"c.created_at", "c.updated_at", "c.deleted_at",
		"c.position", "c.bot_message_color", "c.user_message_color",
		"c.bot_message_text_color", "c.user_message_text_color",
		"c.chat_font_family", "c.chat_header_color", "c.chat_header_text_color",
		"c.chat_background_color", "COALESCE(c.bubble_radius, '22px')",
		"COALESCE(c.input_background_color, '#ededed')",
		"COALESCE(c.input_text_color, '#000000')",
		"COALESCE(c.send_button_color, '#ebb800')",
		"c.bot_icon", "c.bot_display_name", "c.allowed_domains", "c.embed_secret", "c.secure_embed_enabled",
		"c.suggested_questions", "c.manual_questions", "c.suggestions_enabled",
		"c.include_paths", "c.exclude_paths", "c.selector_whitelist", "COALESCE(c.discovery_mode, 'auto')",
		"COALESCE(c.refresh_policy, 'manual')", "c.refresh_frequency", "c.next_refresh_at", "c.last_refresh_at",
		"COALESCE(c.hide_branding, false)", "c.custom_branding",
		"c.confidence_threshold", "c.threshold_config", "c.fallback_messages", "c.topic_restrictions",
		"COALESCE(c.handoff_enabled, false)", "COALESCE(c.handoff_type, 'email')", "c.handoff_config",
	}
}

// GetByID retrieves a chatbot by its unique identifier.
// Returns nil, nil if the chatbot is not found.
func (r *PostgresChatbotRepo) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
	query, args, err := psql.
		Select(r.chatbotColumns()...).
		From("chatbots c").
		LeftJoin("languages l ON l.id = c.language_id").
		Where(sq.Eq{"c.id": id}).
		Where(sq.Eq{"c.deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by id query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbot")
	}
	defer rows.Close()

	chatbots, err := r.scanChatbots(rows)
	if err != nil {
		return nil, err
	}
	if len(chatbots) == 0 {
		return nil, nil
	}
	return &chatbots[0], nil
}

// GetByUserID retrieves all non-deleted chatbots for a user.
func (r *PostgresChatbotRepo) GetByUserID(ctx context.Context, userID string) ([]models.Chatbot, error) {
	query, args, err := psql.
		Select(r.chatbotColumns()...).
		From("chatbots c").
		LeftJoin("languages l ON l.id = c.language_id").
		Where(sq.Eq{"c.user_id": userID}).
		Where(sq.Eq{"c.deleted_at": nil}).
		OrderBy("c.created_at DESC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by user id query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbots")
	}
	defer rows.Close()
	return r.scanChatbots(rows)
}

// GetByWorkspace retrieves all non-deleted chatbots for a workspace.
func (r *PostgresChatbotRepo) GetByWorkspace(ctx context.Context, workspaceID string) ([]models.Chatbot, error) {
	query, args, err := psql.
		Select(r.chatbotColumns()...).
		From("chatbots c").
		LeftJoin("languages l ON l.id = c.language_id").
		Where(sq.Eq{"c.workspace_id": workspaceID}).
		Where(sq.Eq{"c.deleted_at": nil}).
		OrderBy("c.created_at DESC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by workspace query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbots")
	}
	defer rows.Close()
	return r.scanChatbots(rows)
}

// Create persists a new chatbot and returns its generated ID.
func (r *PostgresChatbotRepo) Create(ctx context.Context, bot *models.Chatbot) (string, error) {
	// Set defaults for new appearance fields
	if bot.BubbleRadius == "" {
		bot.BubbleRadius = "22px"
	}
	if bot.InputBackgroundColor == "" {
		bot.InputBackgroundColor = "#ededed"
	}
	if bot.InputTextColor == "" {
		bot.InputTextColor = "#000000"
	}
	if bot.SendButtonColor == "" {
		bot.SendButtonColor = "#ebb800"
	}

	// Serialize JSONB fields
	var sqJSON, mqJSON, tcJSON, fmJSON, trJSON, hcJSON interface{}
	if bot.SuggestedQuestions != nil {
		sqJSON, _ = json.Marshal(bot.SuggestedQuestions)
	}
	if bot.ManualQuestions != nil {
		mqJSON, _ = json.Marshal(bot.ManualQuestions)
	}
	if bot.ThresholdConfig != nil {
		tcJSON, _ = json.Marshal(bot.ThresholdConfig)
	}
	if bot.FallbackMessages != nil {
		fmJSON, _ = json.Marshal(bot.FallbackMessages)
	}
	if bot.TopicRestrictions != nil {
		trJSON, _ = json.Marshal(bot.TopicRestrictions)
	}
	if bot.HandoffConfig != nil {
		hcJSON, _ = json.Marshal(bot.HandoffConfig)
	}

	// Get language_id from normalized locale code
	var languageID string
	normalizedLang := r.normalizeLocale(bot.LanguageCode)
	err := r.pool.QueryRowContext(ctx, "SELECT id FROM languages WHERE code=$1", normalizedLang).Scan(&languageID)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "lookup language id for code %s", normalizedLang)
	}

	query, args, err := psql.
		Insert("chatbots").
		Columns(
			"user_id", "workspace_id", "organization_id", "name", "description", "custom_instruction",
			"language_id", "model", "temperature", "max_tokens", "theme_color", "welcome_message",
			"position", "bot_message_color", "user_message_color",
			"bot_message_text_color", "user_message_text_color",
			"chat_font_family", "chat_header_color", "chat_header_text_color",
			"chat_background_color", "bubble_radius", "input_background_color", "input_text_color", "send_button_color",
			"bot_icon", "bot_display_name", "suggested_questions", "manual_questions", "suggestions_enabled",
			"include_paths", "exclude_paths", "selector_whitelist", "discovery_mode",
			"refresh_policy", "refresh_frequency", "next_refresh_at", "last_refresh_at",
			"confidence_threshold", "threshold_config", "fallback_messages", "topic_restrictions",
			"handoff_enabled", "handoff_type", "handoff_config",
		).
		Values(
			bot.UserID, bot.WorkspaceID, bot.OrganizationID, bot.Name, bot.Description, bot.CustomInstruction,
			languageID,
			bot.Model, bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
			bot.Position, bot.BotMessageColor, bot.UserMessageColor,
			bot.BotMessageTextColor, bot.UserMessageTextColor,
			bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
			bot.ChatBackgroundColor, bot.BubbleRadius, bot.InputBackgroundColor, bot.InputTextColor, bot.SendButtonColor,
			bot.BotIcon, bot.BotDisplayName, sqJSON, mqJSON, bot.SuggestionsEnabled,
			pq.Array(bot.IncludePaths), pq.Array(bot.ExcludePaths), pq.Array(bot.SelectorWhitelist), bot.DiscoveryMode,
			bot.RefreshPolicy, bot.RefreshFrequency, bot.NextRefreshAt, bot.LastRefreshAt,
			bot.ConfidenceThreshold, tcJSON, fmJSON, trJSON,
			bot.HandoffEnabled, bot.HandoffType, hcJSON,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", pkgerrors.Wrapf(err, "build create query")
	}

	var id string
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "create chatbot")
	}
	return id, nil
}

// Update modifies an existing chatbot's fields.
func (r *PostgresChatbotRepo) Update(ctx context.Context, bot *models.Chatbot) error {
	// suggested_questions and manual_questions are JSONB, so marshal to JSON
	var sj, mqj []byte
	if bot.SuggestedQuestions != nil {
		sj, _ = json.Marshal(bot.SuggestedQuestions)
	}
	if bot.ManualQuestions != nil {
		mqj, _ = json.Marshal(bot.ManualQuestions)
	}

	// For struct pointer fields (JSONB), use explicit type handling
	var cbj, tcj, fmj, trj, hcj []byte
	if bot.CustomBranding != nil {
		cbj, _ = json.Marshal(bot.CustomBranding)
	}
	if bot.ThresholdConfig != nil {
		tcj, _ = json.Marshal(bot.ThresholdConfig)
	}
	if bot.FallbackMessages != nil {
		fmj, _ = json.Marshal(bot.FallbackMessages)
	}
	if bot.TopicRestrictions != nil {
		trj, _ = json.Marshal(bot.TopicRestrictions)
	}
	if bot.HandoffConfig != nil {
		hcj, _ = json.Marshal(bot.HandoffConfig)
	}

	// Get language_id from normalized locale code
	var languageID string
	normalizedLang := r.normalizeLocale(bot.LanguageCode)
	err := r.pool.QueryRowContext(ctx, "SELECT id FROM languages WHERE code=$1", normalizedLang).Scan(&languageID)
	if err != nil {
		return pkgerrors.Wrapf(err, "lookup language id for code %s", normalizedLang)
	}

	query, args, err := psql.
		Update("chatbots").
		Set("name", bot.Name).
		Set("description", bot.Description).
		Set("custom_instruction", bot.CustomInstruction).
		Set("language_id", languageID).
		Set("model", bot.Model).
		Set("temperature", bot.Temperature).
		Set("max_tokens", bot.MaxTokens).
		Set("theme_color", bot.ThemeColor).
		Set("welcome_message", bot.WelcomeMessage).
		Set("position", bot.Position).
		Set("bot_message_color", bot.BotMessageColor).
		Set("user_message_color", bot.UserMessageColor).
		Set("bot_message_text_color", bot.BotMessageTextColor).
		Set("user_message_text_color", bot.UserMessageTextColor).
		Set("chat_font_family", bot.ChatFontFamily).
		Set("chat_header_color", bot.ChatHeaderColor).
		Set("chat_header_text_color", bot.ChatHeaderTextColor).
		Set("chat_background_color", bot.ChatBackgroundColor).
		Set("bubble_radius", bot.BubbleRadius).
		Set("input_background_color", bot.InputBackgroundColor).
		Set("input_text_color", bot.InputTextColor).
		Set("send_button_color", bot.SendButtonColor).
		Set("bot_icon", bot.BotIcon).
		Set("bot_display_name", bot.BotDisplayName).
		Set("allowed_domains", bot.AllowedDomains).
		Set("embed_secret", bot.EmbedSecret).
		Set("secure_embed_enabled", bot.SecureEmbedEnabled).
		Set("suggested_questions", sj).
		Set("suggestions_enabled", bot.SuggestionsEnabled).
		Set("manual_questions", mqj).
		Set("include_paths", pq.Array(bot.IncludePaths)).
		Set("exclude_paths", pq.Array(bot.ExcludePaths)).
		Set("selector_whitelist", pq.Array(bot.SelectorWhitelist)).
		Set("discovery_mode", bot.DiscoveryMode).
		Set("refresh_policy", bot.RefreshPolicy).
		Set("refresh_frequency", bot.RefreshFrequency).
		Set("hide_branding", bot.HideBranding).
		Set("custom_branding", cbj).
		Set("confidence_threshold", bot.ConfidenceThreshold).
		Set("threshold_config", tcj).
		Set("fallback_messages", fmj).
		Set("topic_restrictions", trj).
		Set("handoff_enabled", bot.HandoffEnabled).
		Set("handoff_type", bot.HandoffType).
		Set("handoff_config", hcj).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": bot.ID}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update chatbot")
	}
	return nil
}

// SoftDelete marks a chatbot as deleted and returns the IDs of associated sources.
// This allows cleanup of related resources (e.g., vectors in Qdrant).
func (r *PostgresChatbotRepo) SoftDelete(ctx context.Context, id, userID string) ([]string, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Mark chatbot as deleted
	res, err := tx.ExecContext(ctx, `
		UPDATE chatbots SET deleted_at=NOW()
		WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "mark chatbot as deleted")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "rows affected")
	}
	if rows == 0 {
		return nil, nil
	}

	// 2. Get source IDs before soft deleting them
	sourceRows, err := tx.QueryContext(ctx, `
		SELECT id FROM data_sources
		WHERE chatbot_id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query sources for deletion")
	}
	defer func() { _ = sourceRows.Close() }()

	var sourceIDs []string
	for sourceRows.Next() {
		var sid string
		if scanErr := sourceRows.Scan(&sid); scanErr != nil {
			return nil, pkgerrors.Wrapf(scanErr, "scan source id for deletion")
		}
		sourceIDs = append(sourceIDs, sid)
	}
	if rowsErr := sourceRows.Err(); rowsErr != nil {
		return nil, pkgerrors.Wrapf(rowsErr, "source rows err")
	}

	// 3. Cascade soft delete sources
	_, err = tx.ExecContext(ctx, `
		UPDATE data_sources SET deleted_at=NOW()
		WHERE chatbot_id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "cascade soft delete sources")
	}

	// 4. Hard delete analytics
	_, err = tx.ExecContext(ctx, `DELETE FROM analytics WHERE chatbot_id=$1`, id)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "hard delete analytics")
	}

	if err := tx.Commit(); err != nil {
		return nil, pkgerrors.Wrapf(err, "commit tx")
	}

	return sourceIDs, nil
}

// CountByUserID returns the number of active chatbots for a user.
func (r *PostgresChatbotRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("chatbots").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count query")
	}

	var count int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "count chatbots")
	}
	return count, nil
}

// CountByWorkspace returns the number of active chatbots for a workspace.
func (r *PostgresChatbotRepo) CountByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("chatbots").
		Where(sq.Eq{"workspace_id": workspaceID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count query")
	}

	var count int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "count chatbots")
	}
	return count, nil
}

// UpdateSuggestedQuestions updates only the AI-generated suggestions.
func (r *PostgresChatbotRepo) UpdateSuggestedQuestions(ctx context.Context, id string, suggestions []string) error {
	js, err := json.Marshal(suggestions)
	if err != nil {
		return pkgerrors.Wrapf(err, "marshal suggestions")
	}

	query, args, err := psql.
		Update("chatbots").
		Set("suggested_questions", js).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update suggestions query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update suggestions")
	}
	return nil
}

// GetDueForRefresh returns chatbots with auto refresh enabled that are due for refresh.
func (r *PostgresChatbotRepo) GetDueForRefresh(ctx context.Context, now time.Time) ([]models.Chatbot, error) {
	query, args, err := psql.
		Select(r.chatbotColumns()...).
		From("chatbots c").
		LeftJoin("languages l ON l.id = c.language_id").
		Where(sq.Eq{"c.refresh_policy": "auto"}).
		Where(sq.LtOrEq{"c.next_refresh_at": now}).
		Where(sq.Eq{"c.deleted_at": nil}).
		OrderBy("c.next_refresh_at ASC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get due for refresh query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbots due for refresh")
	}
	defer rows.Close()
	return r.scanChatbots(rows)
}

// UpdateRefreshTimes updates the next_refresh_at and last_refresh_at for a chatbot.
func (r *PostgresChatbotRepo) UpdateRefreshTimes(ctx context.Context, botID string, nextRefresh, lastRefresh time.Time) error {
	query, args, err := psql.
		Update("chatbots").
		Set("next_refresh_at", nextRefresh).
		Set("last_refresh_at", lastRefresh).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": botID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update refresh times query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update chatbot refresh times")
	}
	return nil
}
