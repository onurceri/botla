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
	err := pool.QueryRowContext(
		ctx,
		`INSERT INTO chatbots (
            user_id, name, description, system_prompt, language_id, model,
            temperature, max_tokens, theme_color, welcome_message,
            position, bot_message_color, user_message_color,
            bot_message_text_color, user_message_text_color,
            chat_font_family, chat_header_color, chat_header_text_color,
            chat_background_color, bot_icon, bot_display_name, suggested_questions, suggestions_enabled
        ) VALUES ($1,$2,$3,$4,(SELECT id FROM languages WHERE code=$5),$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23) RETURNING id`,
		bot.UserID, bot.Name, bot.Description, bot.SystemPrompt, normalizeLocale(bot.LanguageCode), bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
		bot.ChatBackgroundColor, bot.BotIcon, bot.BotDisplayName, bot.SuggestedQuestions, bot.SuggestionsEnabled,
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
               c.suggested_questions, c.suggestions_enabled
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
		var sj []byte
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
		); err != nil {
			return nil, err
		}
		if len(sj) > 0 {
			var arr []string
			_ = json.Unmarshal(sj, &arr)
			c.SuggestedQuestions = arr
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
	var sj []byte
	err := pool.QueryRowContext(ctx, `
        SELECT c.id, c.user_id, c.name, c.description, c.system_prompt, COALESCE(l.code,'') AS language_code, c.model,
               temperature, max_tokens, theme_color, welcome_message,
               c.created_at, c.updated_at, c.deleted_at,
               c.position, c.bot_message_color, c.user_message_color,
               c.bot_message_text_color, c.user_message_text_color,
               c.chat_font_family, c.chat_header_color, c.chat_header_text_color,
               c.chat_background_color,
               c.bot_icon, c.bot_display_name, c.allowed_domains, c.embed_secret, c.secure_embed_enabled,
               c.suggested_questions, c.suggestions_enabled
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
	return &c, nil
}

func UpdateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) error {
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
            updated_at=NOW()
        WHERE id=$26 AND user_id=$27 AND deleted_at IS NULL`,
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
