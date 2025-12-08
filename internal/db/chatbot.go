package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/onurceri/botla-co/internal/models"
)

func CreateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) (string, error) {
	var id string

	// Serialize JSON fields
	var fmJSON, trJSON interface{}
	if bot.FallbackMessages != nil {
		fmJSON, _ = json.Marshal(bot.FallbackMessages)
	}
	if bot.TopicRestrictions != nil {
		trJSON, _ = json.Marshal(bot.TopicRestrictions)
	}

	err := pool.QueryRowContext(
		ctx,
		`INSERT INTO chatbots (
            user_id, workspace_id, organization_id, name, description, system_prompt, language_id, model,
            temperature, max_tokens, theme_color, welcome_message,
            position, bot_message_color, user_message_color,
            bot_message_text_color, user_message_text_color,
            chat_font_family, chat_header_color, chat_header_text_color,
            chat_background_color, bot_icon, bot_display_name, suggested_questions, suggestions_enabled,
            include_paths, exclude_paths, selector_whitelist, discovery_mode,
            refresh_policy, refresh_frequency, next_refresh_at, last_refresh_at,
            confidence_threshold, fallback_messages, topic_restrictions
        ) VALUES ($1,$2,$3,$4,$5,$6,(SELECT id FROM languages WHERE code=$7),$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36) RETURNING id`,
		bot.UserID, bot.WorkspaceID, bot.OrganizationID, bot.Name, bot.Description, bot.SystemPrompt, normalizeLocale(bot.LanguageCode), bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor, bot.BotIcon, bot.BotDisplayName, bot.SuggestedQuestions, bot.SuggestionsEnabled,
		bot.IncludePaths, bot.ExcludePaths, bot.SelectorWhitelist, bot.DiscoveryMode,
		bot.RefreshPolicy, bot.RefreshFrequency, bot.NextRefreshAt, bot.LastRefreshAt,
		bot.ConfidenceThreshold, fmJSON, trJSON,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GetChatbotsByUserID(ctx context.Context, pool *sql.DB, userID string) ([]models.Chatbot, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT c.id, c.user_id, c.workspace_id, c.organization_id, c.name, c.description, c.system_prompt, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.suggestions_enabled,
               c.include_paths, c.exclude_paths, c.selector_whitelist, COALESCE(c.discovery_mode, 'auto') AS discovery_mode,
               COALESCE(c.refresh_policy, 'manual') AS refresh_policy, c.refresh_frequency, c.next_refresh_at, c.last_refresh_at,
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding,
               c.confidence_threshold, c.fallback_messages, c.topic_restrictions,
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
        SELECT c.id, c.user_id, c.workspace_id, c.organization_id, c.name, c.description, c.system_prompt, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.suggestions_enabled,
               c.include_paths, c.exclude_paths, c.selector_whitelist, COALESCE(c.discovery_mode, 'auto') AS discovery_mode,
               COALESCE(c.refresh_policy, 'manual') AS refresh_policy, c.refresh_frequency, c.next_refresh_at, c.last_refresh_at,
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding,
               c.confidence_threshold, c.fallback_messages, c.topic_restrictions,
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
		var sj, ipj, epj, swj, cbj, fmj, trj, hcj []byte
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.WorkspaceID, &c.OrganizationID, &c.Name, &c.Description, &c.SystemPrompt, &c.LanguageCode, &c.Model,
			&c.Temperature, &c.MaxTokens, &c.ThemeColor, &c.WelcomeMessage,
			&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
			&c.Position, &c.BotMessageColor, &c.UserMessageColor,
			&c.BotMessageTextColor, &c.UserMessageTextColor,
			&c.ChatFontFamily, &c.ChatHeaderColor, &c.ChatHeaderTextColor,
			&c.ChatBackgroundColor,
			&c.BotIcon, &c.BotDisplayName, &c.AllowedDomains, &c.EmbedSecret, &c.SecureEmbedEnabled,
			&sj, &c.SuggestionsEnabled,
			&ipj, &epj, &swj, &c.DiscoveryMode,
			&c.RefreshPolicy, &c.RefreshFrequency, &c.NextRefreshAt, &c.LastRefreshAt,
			&c.HideBranding, &cbj,
			&c.ConfidenceThreshold, &fmj, &trj,
			&c.HandoffEnabled, &c.HandoffType, &hcj,
		); err != nil {
			return nil, err
		}
		if len(sj) > 0 {
			var arr []string
			_ = json.Unmarshal(sj, &arr)
			c.SuggestedQuestions = arr
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
	var sj, ipj, epj, swj, cbj, fmj, trj, hcj []byte
	err := pool.QueryRowContext(ctx, `
        SELECT c.id, c.user_id, c.workspace_id, c.organization_id, c.name, c.description, c.system_prompt, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.suggestions_enabled,
               c.include_paths, c.exclude_paths, c.selector_whitelist, COALESCE(c.discovery_mode, 'auto') AS discovery_mode,
               COALESCE(c.refresh_policy, 'manual') AS refresh_policy, c.refresh_frequency, c.next_refresh_at, c.last_refresh_at,
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding,
               c.confidence_threshold, c.fallback_messages, c.topic_restrictions,
               COALESCE(c.handoff_enabled, false) AS handoff_enabled, COALESCE(c.handoff_type, 'email') AS handoff_type, c.handoff_config
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.id=$1 AND c.deleted_at IS NULL`, id).
		Scan(
			&c.ID, &c.UserID, &c.WorkspaceID, &c.OrganizationID, &c.Name, &c.Description, &c.SystemPrompt, &c.LanguageCode, &c.Model,
			&c.Temperature, &c.MaxTokens, &c.ThemeColor, &c.WelcomeMessage,
			&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
			&c.Position, &c.BotMessageColor, &c.UserMessageColor,
			&c.BotMessageTextColor, &c.UserMessageTextColor,
			&c.ChatFontFamily, &c.ChatHeaderColor, &c.ChatHeaderTextColor,
			&c.ChatBackgroundColor,
			&c.BotIcon, &c.BotDisplayName, &c.AllowedDomains, &c.EmbedSecret, &c.SecureEmbedEnabled,
			&sj, &c.SuggestionsEnabled,
			&ipj, &epj, &swj, &c.DiscoveryMode,
			&c.RefreshPolicy, &c.RefreshFrequency, &c.NextRefreshAt, &c.LastRefreshAt,
			&c.HideBranding, &cbj,
			&c.ConfidenceThreshold, &fmj, &trj,
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

func UpdateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) error {
	// Serialize custom_branding to JSON
	var cbJSON interface{}
	if bot.CustomBranding != nil {
		cbJSON, _ = json.Marshal(bot.CustomBranding)
	}
	// Serialize new fields
	var fmJSON, trJSON, hcJSON interface{}
	if bot.FallbackMessages != nil {
		fmJSON, _ = json.Marshal(bot.FallbackMessages)
	}
	if bot.TopicRestrictions != nil {
		trJSON, _ = json.Marshal(bot.TopicRestrictions)
	}
	if bot.HandoffConfig != nil {
		hcJSON, _ = json.Marshal(bot.HandoffConfig)
	}

	_, err := pool.ExecContext(ctx, `
        UPDATE chatbots SET
            name=$1,
            description=$2,
            system_prompt=$3,
            language_id=(SELECT id FROM languages WHERE code=$4),
            model=$5,
            temperature=$6,
            max_tokens=$7,
            theme_color=$8,
            welcome_message=$9,
            position=$10,
            bot_message_color=$11,
            user_message_color=$12,
            bot_message_text_color=$13,
            user_message_text_color=$14,
            chat_font_family=$15,
            chat_header_color=$16,
            chat_header_text_color=$17,
            chat_background_color=$18,
            bot_icon=$19,
            bot_display_name=$20,
            allowed_domains=$21,
            embed_secret=$22,
            secure_embed_enabled=$23,
            suggested_questions=$24,
            suggestions_enabled=$25,
            include_paths=$26,
            exclude_paths=$27,
            selector_whitelist=$28,
            discovery_mode=$29,
            refresh_policy=$30,
            refresh_frequency=$31,
            next_refresh_at=$32,
            last_refresh_at=$33,
            hide_branding=$34,
            custom_branding=$35,
            confidence_threshold=$36,
            fallback_messages=$37,
            topic_restrictions=$38,
            handoff_enabled=$39,
            handoff_type=$40,
            handoff_config=$41,
            updated_at=NOW()
        WHERE id=$42 AND user_id=$43 AND deleted_at IS NULL`,
		bot.Name,
		bot.Description,
		bot.SystemPrompt,
		normalizeLocale(bot.LanguageCode),
		bot.Model,
		bot.Temperature,
		bot.MaxTokens,
		bot.ThemeColor,
		bot.WelcomeMessage,
		bot.Position,
		bot.BotMessageColor,
		bot.UserMessageColor,
		bot.BotMessageTextColor,
		bot.UserMessageTextColor,
		bot.ChatFontFamily,
		bot.ChatHeaderColor,
		bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor,
		bot.BotIcon,
		bot.BotDisplayName,
		bot.AllowedDomains,
		bot.EmbedSecret,
		bot.SecureEmbedEnabled,
		bot.SuggestedQuestions,
		bot.SuggestionsEnabled,
		bot.IncludePaths,
		bot.ExcludePaths,
		bot.SelectorWhitelist,
		bot.DiscoveryMode,
		bot.RefreshPolicy,
		bot.RefreshFrequency,
		bot.NextRefreshAt,
		bot.LastRefreshAt,
		bot.HideBranding,
		cbJSON,
		bot.ConfidenceThreshold,
		fmJSON,
		trJSON,
		bot.HandoffEnabled,
		bot.HandoffType,
		hcJSON,
		bot.ID,
		bot.UserID,
	)
	return err
}

func UpdateChatbotSuggestions(ctx context.Context, pool *sql.DB, chatbotID string, suggestions []string) error {
	var js any
	if suggestions == nil {
		js = nil
	} else {
		js = suggestions
	}
	_, err := pool.ExecContext(ctx, `UPDATE chatbots SET suggested_questions=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`, js, chatbotID)
	return err
}

func SoftDeleteChatbot(ctx context.Context, pool *sql.DB, id, userID string) error {
	_, err := pool.ExecContext(ctx, `
        UPDATE chatbots SET deleted_at=NOW()
        WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, userID)
	return err
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
