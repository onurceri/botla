package db

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/models"
)

func CreateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) (string, error) {
	var id string
	err := pool.QueryRowContext(
		ctx,
        `INSERT INTO chatbots (
            user_id, name, description, system_prompt, language, model,
            temperature, max_tokens, theme_color, welcome_message,
            position, bot_message_color, user_message_color,
            bot_message_text_color, user_message_text_color,
            chat_font_family, chat_header_color, chat_header_text_color,
            chat_background_color, bot_icon, bot_display_name
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21) RETURNING id`,
		bot.UserID, bot.Name, bot.Description, bot.SystemPrompt, bot.Language, bot.Model,
		bot.Temperature, bot.MaxTokens, bot.ThemeColor, bot.WelcomeMessage,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor,
		bot.BotMessageTextColor, bot.UserMessageTextColor,
        bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor,
        bot.ChatBackgroundColor, bot.BotIcon, bot.BotDisplayName,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GetChatbotsByUserID(ctx context.Context, pool *sql.DB, userID string) ([]models.Chatbot, error) {
    rows, err := pool.QueryContext(ctx, `
        SELECT id, user_id, name, description, system_prompt, language, model,
               temperature, max_tokens, theme_color, welcome_message,
               created_at, updated_at, deleted_at,
               position, bot_message_color, user_message_color,
               bot_message_text_color, user_message_text_color,
               chat_font_family, chat_header_color, chat_header_text_color,
               chat_background_color,
               bot_icon, bot_display_name, allowed_domains, embed_secret, secure_embed_enabled
        FROM chatbots
        WHERE user_id=$1 AND deleted_at IS NULL
        ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Chatbot
	for rows.Next() {
		var c models.Chatbot
        if err := rows.Scan(
            &c.ID, &c.UserID, &c.Name, &c.Description, &c.SystemPrompt, &c.Language, &c.Model,
            &c.Temperature, &c.MaxTokens, &c.ThemeColor, &c.WelcomeMessage,
            &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
            &c.Position, &c.BotMessageColor, &c.UserMessageColor,
            &c.BotMessageTextColor, &c.UserMessageTextColor,
            &c.ChatFontFamily, &c.ChatHeaderColor, &c.ChatHeaderTextColor,
            &c.ChatBackgroundColor,
            &c.BotIcon, &c.BotDisplayName, &c.AllowedDomains, &c.EmbedSecret, &c.SecureEmbedEnabled,
        ); err != nil {
            return nil, err
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
    err := pool.QueryRowContext(ctx, `
        SELECT id, user_id, name, description, system_prompt, language, model,
               temperature, max_tokens, theme_color, welcome_message,
               created_at, updated_at, deleted_at,
               position, bot_message_color, user_message_color,
               bot_message_text_color, user_message_text_color,
               chat_font_family, chat_header_color, chat_header_text_color,
               chat_background_color,
               bot_icon, bot_display_name, allowed_domains, embed_secret, secure_embed_enabled
        FROM chatbots WHERE id=$1 AND deleted_at IS NULL`, id).
        Scan(
            &c.ID, &c.UserID, &c.Name, &c.Description, &c.SystemPrompt, &c.Language, &c.Model,
            &c.Temperature, &c.MaxTokens, &c.ThemeColor, &c.WelcomeMessage,
            &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
            &c.Position, &c.BotMessageColor, &c.UserMessageColor,
            &c.BotMessageTextColor, &c.UserMessageTextColor,
            &c.ChatFontFamily, &c.ChatHeaderColor, &c.ChatHeaderTextColor,
            &c.ChatBackgroundColor,
            &c.BotIcon, &c.BotDisplayName, &c.AllowedDomains, &c.EmbedSecret, &c.SecureEmbedEnabled,
        )
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func UpdateChatbot(ctx context.Context, pool *sql.DB, bot *models.Chatbot) error {
    _, err := pool.ExecContext(ctx, `
        UPDATE chatbots SET
            name=$1,
            description=$2,
            system_prompt=$3,
            language=$4,
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
            updated_at=NOW()
        WHERE id=$24 AND user_id=$25 AND deleted_at IS NULL`,
        bot.Name,
        bot.Description,
        bot.SystemPrompt,
        bot.Language,
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
        bot.ID,
        bot.UserID,
    )
	return err
}

func SoftDeleteChatbot(ctx context.Context, pool *sql.DB, id, userID string) error {
	_, err := pool.ExecContext(ctx, `
        UPDATE chatbots SET deleted_at=NOW()
        WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, userID)
	return err
}
