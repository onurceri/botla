package models

import "time"

type Chatbot struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	Name                 string     `json:"name"`
	Description          *string    `json:"description,omitempty"`
	SystemPrompt         string     `json:"system_prompt"`
	LanguageCode         string     `json:"language"`
	Model                string     `json:"model"`
	Temperature          float32    `json:"temperature"`
	MaxTokens            int        `json:"max_tokens"`
	ThemeColor           string     `json:"theme_color"`
	WelcomeMessage       string     `json:"welcome_message"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
	Position             string     `json:"position"`
	BotMessageColor      string     `json:"bot_message_color"`
	UserMessageColor     string     `json:"user_message_color"`
	BotMessageTextColor  string     `json:"bot_message_text_color"`
	UserMessageTextColor string     `json:"user_message_text_color"`
	ChatFontFamily       string     `json:"chat_font_family"`
	ChatHeaderColor      string     `json:"chat_header_color"`
	ChatHeaderTextColor  string     `json:"chat_header_text_color"`
	ChatBackgroundColor  string     `json:"chat_background_color"`
	BotIcon              *string    `json:"bot_icon,omitempty"`
	BotDisplayName       *string    `json:"bot_display_name,omitempty"`
	AllowedDomains       *string    `json:"allowed_domains,omitempty"`
	EmbedSecret          *string    `json:"embed_secret,omitempty"`
	SecureEmbedEnabled   bool       `json:"secure_embed_enabled"`
	SuggestedQuestions   []string   `json:"suggested_questions,omitempty"`
	SuggestionsEnabled   bool       `json:"suggestions_enabled"`
	IncludePaths         []string   `json:"include_paths,omitempty"`
	ExcludePaths         []string   `json:"exclude_paths,omitempty"`
	SelectorWhitelist    []string   `json:"selector_whitelist,omitempty"`
}
