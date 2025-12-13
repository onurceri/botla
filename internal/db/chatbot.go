package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/lib/pq"
	"github.com/onurceri/botla-co/internal/models"
)

func CreateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) (string, error) {
	var id string

	// Serialize JSONB fields
	var sqJSON, tcJSON, fmJSON, trJSON, hcJSON interface{}
	if bot.SuggestedQuestions != nil {
		sqJSON, _ = json.Marshal(bot.SuggestedQuestions)
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
            chat_background_color, bot_icon, bot_display_name, suggested_questions, suggestions_enabled,
            include_paths, exclude_paths, selector_whitelist, discovery_mode,
            refresh_policy, refresh_frequency, next_refresh_at, last_refresh_at,
            confidence_threshold, threshold_config, fallback_messages, topic_restrictions,
            handoff_enabled, handoff_type, handoff_config
        ) VALUES ($1,$2,$3,$4,$5,$6,(SELECT id FROM languages WHERE code=$7),$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40) RETURNING id`,
		bot.UserID, bot.WorkspaceID, bot.OrganizationID, bot.Name, bot.Description, bot.CustomInstruction, normalizeLocale(bot.LanguageCode), bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor, bot.BotIcon, bot.BotDisplayName, sqJSON, bot.SuggestionsEnabled,
		pq.Array(bot.IncludePaths), pq.Array(bot.ExcludePaths), pq.Array(bot.SelectorWhitelist), bot.DiscoveryMode,
		bot.RefreshPolicy, bot.RefreshFrequency, bot.NextRefreshAt, bot.LastRefreshAt,
		bot.ConfidenceThreshold, tcJSON, fmJSON, trJSON,
		bot.HandoffEnabled, bot.HandoffType, hcJSON,
	).Scan(&id)
	if err != nil {
		return "", err
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
               c.chat_background_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.all_suggested_questions, c.suggestions_enabled,
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
		return nil, err
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
               c.chat_background_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.all_suggested_questions, c.suggestions_enabled,
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
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanChatbots(rows)
}

// scanChatbots is a helper function to scan chatbot rows
func scanChatbots(rows *sql.Rows) ([]models.Chatbot, error) {
	var out []models.Chatbot
	for rows.Next() {
		var c models.Chatbot
		var sj, asj, ipj, epj, swj, cbj, tcj, fmj, trj, hcj []byte
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.WorkspaceID, &c.OrganizationID, &c.Name, &c.Description, &c.CustomInstruction, &c.LanguageCode, &c.Model,
			&c.Temperature, &c.MaxTokens, &c.ThemeColor, &c.WelcomeMessage,
			&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
			&c.Position, &c.BotMessageColor, &c.UserMessageColor,
			&c.BotMessageTextColor, &c.UserMessageTextColor,
			&c.ChatFontFamily, &c.ChatHeaderColor, &c.ChatHeaderTextColor,
			&c.ChatBackgroundColor,
			&c.BotIcon, &c.BotDisplayName, &c.AllowedDomains, &c.EmbedSecret, &c.SecureEmbedEnabled,
			&sj, &asj, &c.SuggestionsEnabled,
			&ipj, &epj, &swj, &c.DiscoveryMode,
			&c.RefreshPolicy, &c.RefreshFrequency, &c.NextRefreshAt, &c.LastRefreshAt,
			&c.HideBranding, &cbj,
			&c.ConfidenceThreshold, &tcj, &fmj, &trj,
			&c.HandoffEnabled, &c.HandoffType, &hcj,
		); err != nil {
			return nil, err
		}
		if len(sj) > 0 {
			var arr []string
			_ = json.Unmarshal(sj, &arr)
			c.SuggestedQuestions = arr
		}
		if len(asj) > 0 {
			var arr []string
			_ = json.Unmarshal(asj, &arr)
			c.AllSuggestedQuestions = arr
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
		return nil, err
	}
	return out, nil
}

func GetChatbotByID(ctx context.Context, pool *sql.DB, id string) (*models.Chatbot, error) {
	var c models.Chatbot
	var sj, asj, ipj, epj, swj, cbj, tcj, fmj, trj, hcj []byte
	err := pool.QueryRowContext(ctx, `
        SELECT c.id, c.user_id, c.workspace_id, c.organization_id, c.name, c.description, COALESCE(c.custom_instruction,'') AS custom_instruction, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.all_suggested_questions, c.suggestions_enabled,
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
		&c.ChatBackgroundColor,
		&c.BotIcon, &c.BotDisplayName, &c.AllowedDomains, &c.EmbedSecret, &c.SecureEmbedEnabled,
		&sj, &asj, &c.SuggestionsEnabled,
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
		return nil, err
	}
	if len(sj) > 0 {
		var arr []string
		_ = json.Unmarshal(sj, &arr)
		c.SuggestedQuestions = arr
	}
	if len(asj) > 0 {
		var arr []string
		_ = json.Unmarshal(asj, &arr)
		c.AllSuggestedQuestions = arr
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
		return nil, err
	}
	return &u, nil
}

func UpdateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) error {
	// suggested_questions is JSONB, so marshal to JSON
	var sj []byte
	if bot.SuggestedQuestions != nil {
		sj, _ = json.Marshal(bot.SuggestedQuestions)
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
            chat_background_color=$18, bot_icon=$19, bot_display_name=$20,
            updated_at=NOW(), allowed_domains=$21, embed_secret=$22, secure_embed_enabled=$23,
            suggested_questions=$24, suggestions_enabled=$25,
            include_paths=$26, exclude_paths=$27, selector_whitelist=$28, discovery_mode=$29,
            refresh_policy=$30, refresh_frequency=$31,
            hide_branding=$32, custom_branding=$33,
            confidence_threshold=$34, threshold_config=$35, fallback_messages=$36, topic_restrictions=$37,
            handoff_enabled=$38, handoff_type=$39, handoff_config=$40
        WHERE id=$41`,
		bot.Name, bot.Description, bot.CustomInstruction, normalizeLocale(bot.LanguageCode), bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor, bot.BotIcon, bot.BotDisplayName,
		bot.AllowedDomains, bot.EmbedSecret, bot.SecureEmbedEnabled,
		sj, bot.SuggestionsEnabled,
		pq.Array(bot.IncludePaths), pq.Array(bot.ExcludePaths), pq.Array(bot.SelectorWhitelist), bot.DiscoveryMode,
		bot.RefreshPolicy, bot.RefreshFrequency,
		bot.HideBranding, cbj,
		bot.ConfidenceThreshold, tcj, fmj, trj,
		bot.HandoffEnabled, bot.HandoffType, hcj,
		bot.ID,
	)
	return err
}

func UpdateChatbotSuggestions(ctx context.Context, pool *sql.DB, id string, suggestions []string) error {
	js, err := json.Marshal(suggestions)
	if err != nil {
		return err
	}
	_, err = pool.ExecContext(ctx, `UPDATE chatbots SET suggested_questions=$1, updated_at=NOW() WHERE id=$2`, js, id)
	return err
}

func SoftDeleteChatbot(ctx context.Context, pool *sql.DB, id, userID string) ([]string, error) {
	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Mark chatbot as deleted
	res, err := tx.ExecContext(ctx, `
        UPDATE chatbots SET deleted_at=NOW()
        WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer func() { _ = sourceRows.Close() }()

	var sourceIDs []string
	for sourceRows.Next() {
		var sid string
		if scanErr := sourceRows.Scan(&sid); scanErr != nil {
			return nil, scanErr
		}
		sourceIDs = append(sourceIDs, sid)
	}
	if rowsErr := sourceRows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	// 3. Cascade soft delete sources
	_, err = tx.ExecContext(ctx, `
        UPDATE data_sources SET deleted_at=NOW()
        WHERE chatbot_id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}

	// 4. Hard delete analytics (since they don't support soft delete and we want to cascade)
	_, err = tx.ExecContext(ctx, `DELETE FROM analytics WHERE chatbot_id=$1`, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return sourceIDs, nil
}

func CountChatbotsByUserID(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM chatbots
        WHERE user_id=$1 AND deleted_at IS NULL
    `, userID).Scan(&count)
	return count, err
}

func CountChatbotsByWorkspace(ctx context.Context, pool *sql.DB, workspaceID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM chatbots
        WHERE workspace_id=$1 AND deleted_at IS NULL
    `, workspaceID).Scan(&count)
	return count, err
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
