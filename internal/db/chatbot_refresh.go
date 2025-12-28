package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

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
               COALESCE(c.hide_branding, false) AS hide_branding, c.custom_branding,
               c.confidence_threshold, c.fallback_messages, c.topic_restrictions,
               COALESCE(c.handoff_enabled, false) AS handoff_enabled, COALESCE(c.handoff_type, 'email') AS handoff_type, c.handoff_config
        FROM chatbots c
        LEFT JOIN languages l ON l.id = c.language_id
        WHERE c.refresh_policy = 'auto' 
          AND c.next_refresh_at <= $1
          AND c.deleted_at IS NULL
        ORDER BY c.next_refresh_at ASC`, now)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query chatbots due for refresh")
	}
	defer func() { _ = rows.Close() }()
	var out []models.Chatbot
	for rows.Next() {
		var c models.Chatbot
		var sj, ipj, epj, swj, cbj, fmj, trj, hcj []byte
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
			&c.ConfidenceThreshold, &fmj, &trj,
			&c.HandoffEnabled, &c.HandoffType, &hcj,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan chatbot for refresh")
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
		return nil, pkgerrors.Wrapf(err, "chatbots due for refresh rows err")
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
	if err != nil {
		return pkgerrors.Wrapf(err, "update chatbot refresh times")
	}
	return nil
}
