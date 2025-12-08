package models

import "time"

type Chatbot struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	WorkspaceID          *string    `json:"workspace_id,omitempty"`
	OrganizationID       *string    `json:"organization_id,omitempty"`
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
	DiscoveryMode        string     `json:"discovery_mode"`    // auto, pending, disabled
	RefreshPolicy        string     `json:"refresh_policy"`    // manual, auto
	RefreshFrequency     *string    `json:"refresh_frequency"` // daily, weekly, monthly (only for auto)
	NextRefreshAt        *time.Time `json:"next_refresh_at,omitempty"`
	LastRefreshAt        *time.Time `json:"last_refresh_at,omitempty"`
	HideBranding         bool            `json:"hide_branding"`
	CustomBranding       *CustomBranding `json:"custom_branding,omitempty"`
	ConfidenceThreshold  float64         `json:"confidence_threshold"`
	FallbackMessages     *FallbackMessages `json:"fallback_messages,omitempty"`
	TopicRestrictions    *TopicConfig    `json:"topic_restrictions,omitempty"`
	HandoffEnabled       bool            `json:"handoff_enabled"`
	HandoffType          string          `json:"handoff_type"`
	HandoffConfig        *HandoffConfig  `json:"handoff_config,omitempty"`
}

// CustomBranding represents custom branding configuration (Enterprise plan feature)
type CustomBranding struct {
	LogoURL string `json:"logo_url,omitempty"`
	Text    string `json:"text,omitempty"`
	Link    string `json:"link,omitempty"`
}

type FallbackMessages struct {
	NoInfoFound    string `json:"no_info_found"`
	ErrorMessage   string `json:"error_message"`
	HandoffMessage string `json:"handoff_message"`
}

type TopicConfig struct {
	AllowedTopics  []string `json:"allowed_topics,omitempty"`
	BlockedTopics  []string `json:"blocked_topics,omitempty"`
	BlockedMessage string   `json:"blocked_message,omitempty"`
}
