package handlers

import (
	"database/sql"

	"github.com/onurceri/botla-co/pkg/config"
)

// ChatbotHandlers handles chatbot-related HTTP endpoints
type ChatbotHandlers struct {
	DB  *sql.DB
	Cfg *config.Config
}

type createChatbotRequest struct {
	Name                 string    `json:"name"`
	Description          *string   `json:"description"`
	SystemPrompt         *string   `json:"system_prompt"`
	Language             *string   `json:"language"`
	Model                *string   `json:"model"`
	Temperature          *float32  `json:"temperature"`
	MaxTokens            *int      `json:"max_tokens"`
	ThemeColor           *string   `json:"theme_color"`
	WelcomeMessage       *string   `json:"welcome_message"`
	Position             *string   `json:"position"`
	BotMessageColor      *string   `json:"bot_message_color"`
	UserMessageColor     *string   `json:"user_message_color"`
	BotMessageTextColor  *string   `json:"bot_message_text_color"`
	UserMessageTextColor *string   `json:"user_message_text_color"`
	ChatFontFamily       *string   `json:"chat_font_family"`
	ChatHeaderColor      *string   `json:"chat_header_color"`
	ChatHeaderTextColor  *string   `json:"chat_header_text_color"`
	ChatBackgroundColor  *string   `json:"chat_background_color"`
	BotIcon              *string   `json:"bot_icon"`
	BotDisplayName       *string   `json:"bot_display_name"`
	SecureEmbedEnabled   *bool     `json:"secure_embed_enabled"`
	AllowedDomains       *string   `json:"allowed_domains"`
	EmbedSecret          *string   `json:"embed_secret"`
	SuggestedQuestions   *[]string `json:"suggested_questions"`
	SuggestionsEnabled   *bool     `json:"suggestions_enabled"`
	IncludePaths         *[]string `json:"include_paths"`
	ExcludePaths         *[]string `json:"exclude_paths"`
	SelectorWhitelist    *[]string `json:"selector_whitelist"`
	DiscoveryMode        *string   `json:"discovery_mode"`
	RefreshPolicy        *string   `json:"refresh_policy"`
	RefreshFrequency     *string   `json:"refresh_frequency"`
}
