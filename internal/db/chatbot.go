package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

func CreateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) (string, error) {
	var id string
	err := pool.QueryRowContext(
		ctx,
		`INSERT INTO chatbots (
            user_id, name, description, system_prompt, language_id, model,
            temperature, max_tokens, theme_color, welcome_message,
            position, bot_message_color, user_message_color,
            bot_message_text_color, user_message_text_color,
            chat_font_family, chat_header_color, chat_header_text_color,
            chat_background_color, bot_icon, bot_display_name, suggested_questions, suggestions_enabled,
            include_paths, exclude_paths, selector_whitelist, discovery_mode,
            refresh_policy, refresh_frequency, next_refresh_at, last_refresh_at
        ) VALUES ($1,$2,$3,$4,(SELECT id FROM languages WHERE code=$5),$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31) RETURNING id`,
		bot.UserID, bot.Name, bot.Description, bot.SystemPrompt, normalizeLocale(bot.LanguageCode), bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor, bot.BotIcon, bot.BotDisplayName, bot.SuggestedQuestions, bot.SuggestionsEnabled,
		bot.IncludePaths, bot.ExcludePaths, bot.SelectorWhitelist, bot.DiscoveryMode,
		bot.RefreshPolicy, bot.RefreshFrequency, bot.NextRefreshAt, bot.LastRefreshAt,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GetChatbotsByUserID(ctx context.Context, pool *sql.DB, userID string) ([]models.Chatbot, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT c.id, c.user_id, c.name, c.description, c.system_prompt, COALESCE(l.code,'') AS language_code, c.model,
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
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.user_id=$1 AND c.deleted_at IS NULL
        ORDER BY c.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []models.Chatbot
	for rows.Next() {
		var c models.Chatbot
		var sj, ipj, epj, swj, cbj []byte
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &c.Description, &c.SystemPrompt, &c.LanguageCode, &c.Model,
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
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func GetChatbotByID(ctx context.Context, pool *sql.DB, id string) (*models.Chatbot, error) {
	var c models.Chatbot
	var sj, ipj, epj, swj, cbj []byte
	err := pool.QueryRowContext(ctx, `
        SELECT c.id, c.user_id, c.name, c.description, c.system_prompt, COALESCE(l.code,'') AS language_code, c.model,
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
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.id=$1 AND c.deleted_at IS NULL`, id).
		Scan(
			&c.ID, &c.UserID, &c.Name, &c.Description, &c.SystemPrompt, &c.LanguageCode, &c.Model,
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
	return &c, nil
}

func UpdateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) error {
	// Serialize custom_branding to JSON
	var cbJSON interface{}
	if bot.CustomBranding != nil {
		cbJSON, _ = json.Marshal(bot.CustomBranding)
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
            updated_at=NOW()
        WHERE id=$36 AND user_id=$37 AND deleted_at IS NULL`,
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
	if s == ""  {
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

// GetChatbotsDueForRefresh returns chatbots with auto refresh enabled that are due for refresh
func GetChatbotsDueForRefresh(ctx context.Context, pool *sql.DB, now time.Time) ([]models.Chatbot, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT c.id, c.user_id, c.name, c.description, c.system_prompt, COALESCE(l.code,'') AS language_code, c.model,
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
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.refresh_policy = 'auto' 
          AND c.next_refresh_at <= $1
          AND c.deleted_at IS NULL
        ORDER BY c.next_refresh_at ASC`, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []models.Chatbot
	for rows.Next() {
		var c models.Chatbot
		var sj, ipj, epj, swj, cbj []byte
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &c.Description, &c.SystemPrompt, &c.LanguageCode, &c.Model,
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
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateChatbotRefreshTimes updates the next_refresh_at and last_refresh_at for a chatbot
func UpdateChatbotRefreshTimes(ctx context.Context, pool *sql.DB, botID string, nextRefresh, lastRefresh time.Time) error {
	_, err := pool.ExecContext(ctx, `
        UPDATE chatbots SET 
            next_refresh_at = $1, 
            last_refresh_at = $2, 
            updated_at = NOW() 
        WHERE id = $3 AND deleted_at IS NULL`,
		nextRefresh, lastRefresh, botID)
	return err
}

// GetAutoRefreshCountForMonth returns the auto_refresh_count for a user in a given month
func GetAutoRefreshCountForMonth(ctx context.Context, pool *sql.DB, userID string, month time.Time) (int, error) {
	// Find the beginning of the month
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	
	var count int
	err := pool.QueryRowContext(ctx, `
        SELECT COALESCE(auto_refresh_count, 0) 
        FROM usage_ingestions 
        WHERE user_id = $1 AND month = $2`,
		userID, startOfMonth).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

// IncrementAutoRefreshCount increments the auto_refresh_count for a user in a given month
func IncrementAutoRefreshCount(ctx context.Context, pool *sql.DB, userID string, month time.Time, delta int) error {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	_, err := pool.ExecContext(ctx, `
        INSERT INTO usage_ingestions (user_id, month, auto_refresh_count, ingestion_count, embedding_tokens)
        VALUES ($1, $2, $3, 0, 0)
        ON CONFLICT(user_id, month) 
        DO UPDATE SET auto_refresh_count = usage_ingestions.auto_refresh_count + $3`,
		userID, startOfMonth, delta)
	return err
}

// GetURLSourcesForChatbot returns all URL sources for a chatbot
func GetURLSourcesForChatbot(ctx context.Context, pool *sql.DB, chatbotID string) ([]models.DataSource, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT id, chatbot_id, source_type, source_url, status, error_message, 
               chunk_count, created_at, hash, deleted_at, last_refreshed_at
        FROM data_sources
        WHERE chatbot_id = $1 AND source_type = 'url' AND deleted_at IS NULL
        ORDER BY created_at DESC`, chatbotID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []models.DataSource
	for rows.Next() {
		var s models.DataSource
		if err := rows.Scan(&s.ID, &s.ChatbotID, &s.SourceType, &s.SourceURL, &s.Status, &s.ErrorMessage,
			&s.ChunkCount, &s.CreatedAt, &s.Hash, &s.DeletedAt, &s.LastRefreshedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

