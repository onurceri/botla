package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/lib/pq"
	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

func CreateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) (string, error) {
	var id string

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

	err := pool.QueryRowContext(
		ctx,
		`INSERT INTO chatbots (
            user_id, workspace_id, organization_id, name, description, custom_instruction, language_id, model,
            temperature, max_tokens, theme_color, welcome_message,
            position, bot_message_color, user_message_color,
            bot_message_text_color, user_message_text_color,
            chat_font_family, chat_header_color, chat_header_text_color,
            chat_background_color, bubble_radius, input_background_color, input_text_color, send_button_color,
            bot_icon, bot_display_name, suggested_questions, manual_questions, suggestions_enabled,
            include_paths, exclude_paths, selector_whitelist, discovery_mode,
            refresh_policy, refresh_frequency, next_refresh_at, last_refresh_at,
            confidence_threshold, threshold_config, fallback_messages, topic_restrictions,
            handoff_enabled, handoff_type, handoff_config
        ) VALUES ($1,$2,$3,$4,$5,$6,(SELECT id FROM languages WHERE code=$7),$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45) RETURNING id`,
		bot.UserID, bot.WorkspaceID, bot.OrganizationID, bot.Name, bot.Description, bot.CustomInstruction, normalizeLocale(bot.LanguageCode), bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor, bot.BubbleRadius, bot.InputBackgroundColor, bot.InputTextColor, bot.SendButtonColor,
		bot.BotIcon, bot.BotDisplayName, sqJSON, mqJSON, bot.SuggestionsEnabled,
		pq.Array(bot.IncludePaths), pq.Array(bot.ExcludePaths), pq.Array(bot.SelectorWhitelist), bot.DiscoveryMode,
		bot.RefreshPolicy, bot.RefreshFrequency, bot.NextRefreshAt, bot.LastRefreshAt,
		bot.ConfidenceThreshold, tcJSON, fmJSON, trJSON,
		bot.HandoffEnabled, bot.HandoffType, hcJSON,
	).Scan(&id)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "create chatbot")
	}
	return id, nil
}

func GetChatbotsByUserID(ctx context.Context, pool *sql.DB, userID string) ([]models.Chatbot, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT c.id, c.user_id, c.workspace_id, c.organization_id, c.name, c.description, COALESCE(c.custom_instruction,'') AS custom_instruction, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color, COALESCE(c.bubble_radius, '22px') AS bubble_radius,
               COALESCE(c.input_background_color, '#ededed') AS input_background_color,
               COALESCE(c.input_text_color, '#000000') AS input_text_color,
               COALESCE(c.send_button_color, '#ebb800') AS send_button_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.manual_questions, c.suggestions_enabled,
               c.include_paths, c.exclude_paths, c.selector_whitelist, COALESCE(c.discovery_mode, 'auto') AS discovery_mode,
               COALESCE(c.refresh_policy, 'manual') AS refresh_policy, c.refresh_frequency, c.next_refresh_at, c.last_refresh_at,
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding,
               c.confidence_threshold, c.threshold_config, c.fallback_messages, c.topic_restrictions,
               COALESCE(c.handoff_enabled, false) AS handoff_enabled, COALESCE(c.handoff_type, 'email') AS handoff_type, c.handoff_config
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.user_id=$1 AND c.deleted_at IS NULL
        ORDER BY c.created_at DESC`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbots by user id")
	}
	defer func() { _ = rows.Close() }()
	return scanChatbots(rows)
}

// GetChatbotsByWorkspace returns all chatbots for a specific workspace
func GetChatbotsByWorkspace(ctx context.Context, pool *sql.DB, workspaceID string) ([]models.Chatbot, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT c.id, c.user_id, c.workspace_id, c.organization_id, c.name, c.description, COALESCE(c.custom_instruction,'') AS custom_instruction, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color, COALESCE(c.bubble_radius, '22px') AS bubble_radius,
               COALESCE(c.input_background_color, '#ededed') AS input_background_color,
               COALESCE(c.input_text_color, '#000000') AS input_text_color,
               COALESCE(c.send_button_color, '#ebb800') AS send_button_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.manual_questions, c.suggestions_enabled,
               c.include_paths, c.exclude_paths, c.selector_whitelist, COALESCE(c.discovery_mode, 'auto') AS discovery_mode,
               COALESCE(c.refresh_policy, 'manual') AS refresh_policy, c.refresh_frequency, c.next_refresh_at, c.last_refresh_at,
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding,
               c.confidence_threshold, c.threshold_config, c.fallback_messages, c.topic_restrictions,
               COALESCE(c.handoff_enabled, false) AS handoff_enabled, COALESCE(c.handoff_type, 'email') AS handoff_type, c.handoff_config
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.workspace_id=$1 AND c.deleted_at IS NULL
        ORDER BY c.created_at DESC`, workspaceID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbots by workspace")
	}
	defer func() { _ = rows.Close() }()
	return scanChatbots(rows)
}

// scanChatbots is a helper function to scan chatbot rows
func scanChatbots(rows *sql.Rows) ([]models.Chatbot, error) {
	var out []models.Chatbot
	for rows.Next() {
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
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan chatbots rows err")
	}
	return out, nil
}

func GetChatbotByID(ctx context.Context, pool *sql.DB, id string) (*models.Chatbot, error) {
	var c models.Chatbot
	var sj, mqj, ipj, epj, swj, cbj, tcj, fmj, trj, hcj []byte
	err := pool.QueryRowContext(ctx, `
        SELECT c.id, c.user_id, c.workspace_id, c.organization_id, c.name, c.description, COALESCE(c.custom_instruction,'') AS custom_instruction, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color, COALESCE(c.bubble_radius, '22px') AS bubble_radius,
               COALESCE(c.input_background_color, '#ededed') AS input_background_color,
               COALESCE(c.input_text_color, '#000000') AS input_text_color,
               COALESCE(c.send_button_color, '#ebb800') AS send_button_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.manual_questions, c.suggestions_enabled,
               c.include_paths, c.exclude_paths, c.selector_whitelist, COALESCE(c.discovery_mode, 'auto') AS discovery_mode,
               COALESCE(c.refresh_policy, 'manual') AS refresh_policy, c.refresh_frequency, c.next_refresh_at, c.last_refresh_at,
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding,
               c.confidence_threshold, c.threshold_config, c.fallback_messages, c.topic_restrictions,
               COALESCE(c.handoff_enabled, false) AS handoff_enabled, COALESCE(c.handoff_type, 'email') AS handoff_type, c.handoff_config
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.id=$1 AND c.deleted_at IS NULL`, id).Scan(
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
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get chatbot by id")
	}
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

func GetUserByBotID(ctx context.Context, pool *sql.DB, botID string) (*models.User, error) {
	var u models.User
	err := pool.QueryRowContext(ctx, `
        SELECT u.id, u.email, u.full_name, u.plan_id, u.preferred_language_id
        FROM users u
        JOIN chatbots c ON c.user_id = u.id
        WHERE c.id = $1`, botID).Scan(
		&u.ID, &u.Email, &u.FullName, &u.PlanID, &u.PreferredLanguageID,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get user by bot id")
	}
	return &u, nil
}

func UpdateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) error {
	// suggested_questions and manual_questions are JSONB, so marshal to JSON
	var sj, mqj []byte
	if bot.SuggestedQuestions != nil {
		sj, _ = json.Marshal(bot.SuggestedQuestions)
	}
	if bot.ManualQuestions != nil {
		mqj, _ = json.Marshal(bot.ManualQuestions)
	}

	// For struct pointer fields (JSONB), keep as interface{} to allow NULL
	var cbj, tcj, fmj, trj, hcj interface{}
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

	// include_paths, exclude_paths, selector_whitelist are TEXT[] columns
	// Use pq.Array() to convert Go slices to PostgreSQL array format
	_, err := pool.ExecContext(ctx, `
        UPDATE chatbots SET
            name=$1, description=$2, custom_instruction=$3, language_id=(SELECT id FROM languages WHERE code=$4), model=$5,
            temperature=$6, max_tokens=$7, theme_color=$8, welcome_message=$9,
            position=$10, bot_message_color=$11, user_message_color=$12,
            bot_message_text_color=$13, user_message_text_color=$14,
            chat_font_family=$15, chat_header_color=$16, chat_header_text_color=$17,
            chat_background_color=$18, bubble_radius=$19, input_background_color=$20, input_text_color=$21, send_button_color=$22,
            bot_icon=$23, bot_display_name=$24,
            updated_at=NOW(), allowed_domains=$25, embed_secret=$26, secure_embed_enabled=$27,
            suggested_questions=COALESCE($28, suggested_questions), suggestions_enabled=$29,
            manual_questions=COALESCE($30, manual_questions),
            include_paths=$31, exclude_paths=$32, selector_whitelist=$33, discovery_mode=$34,
            refresh_policy=$35, refresh_frequency=$36,
            hide_branding=$37, custom_branding=$38,
            confidence_threshold=$39, threshold_config=$40, fallback_messages=$41, topic_restrictions=$42,
            handoff_enabled=$43, handoff_type=$44, handoff_config=$45
        WHERE id=$46`,
		bot.Name, bot.Description, bot.CustomInstruction, normalizeLocale(bot.LanguageCode), bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor, bot.BubbleRadius, bot.InputBackgroundColor, bot.InputTextColor, bot.SendButtonColor,
		bot.BotIcon, bot.BotDisplayName,
		bot.AllowedDomains, bot.EmbedSecret, bot.SecureEmbedEnabled,
		sj, bot.SuggestionsEnabled,
		mqj,
		pq.Array(bot.IncludePaths), pq.Array(bot.ExcludePaths), pq.Array(bot.SelectorWhitelist), bot.DiscoveryMode,
		bot.RefreshPolicy, bot.RefreshFrequency,
		bot.HideBranding, cbj,
		bot.ConfidenceThreshold, tcj, fmj, trj,
		bot.HandoffEnabled, bot.HandoffType, hcj,
		bot.ID,
	)
	if err != nil {
		return pkgerrors.Wrapf(err, "update chatbot")
	}
	return nil
}

// UpdateChatbotSuggestedQuestions updates only the AI-generated suggestions
func UpdateChatbotSuggestedQuestions(ctx context.Context, pool *sql.DB, id string, suggestions []string) error {
	js, err := json.Marshal(suggestions)
	if err != nil {
		return pkgerrors.Wrapf(err, "marshal suggestions")
	}
	_, err = pool.ExecContext(ctx, `UPDATE chatbots SET suggested_questions=$1, updated_at=NOW() WHERE id=$2`, js, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "update chatbot suggestions")
	}
	return nil
}

func SoftDeleteChatbot(ctx context.Context, pool *sql.DB, id, userID string) ([]string, error) {
	tx, err := pool.BeginTx(ctx, nil)
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
	// We only care about sources that are not already deleted
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

	// 4. Hard delete analytics (since they don't support soft delete and we want to cascade)
	_, err = tx.ExecContext(ctx, `DELETE FROM analytics WHERE chatbot_id=$1`, id)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "hard delete analytics")
	}

	if err := tx.Commit(); err != nil {
		return nil, pkgerrors.Wrapf(err, "commit tx")
	}

	return sourceIDs, nil
}

func CountChatbotsByUserID(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM chatbots
        WHERE user_id=$1 AND deleted_at IS NULL
    `, userID).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "count chatbots by user id")
	}
	return count, nil
}

func CountChatbotsByWorkspace(ctx context.Context, pool *sql.DB, workspaceID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM chatbots
        WHERE workspace_id=$1 AND deleted_at IS NULL
    `, workspaceID).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "count chatbots by workspace")
	}
	return count, nil
}

func normalizeLocale(code string) string {
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
