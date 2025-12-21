package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/validation"
	"github.com/onurceri/botla-co/pkg/logger"
)

// =============================================================================
// CHATBOT SERVICE - Business logic for chatbot operations
// =============================================================================

// ChatbotService handles chatbot CRUD operations with validation.
type ChatbotService struct {
	DB        *sql.DB
	Validator *validation.ChatbotValidator
	Log       *logger.Logger
}

// NewChatbotService creates a new ChatbotService.
func NewChatbotService(db *sql.DB, log *logger.Logger) *ChatbotService {
	return &ChatbotService{
		DB:        db,
		Validator: validation.NewChatbotValidator(db),
		Log:       log,
	}
}

// ChatbotUpdateRequest contains all updatable chatbot fields.
type ChatbotUpdateRequest struct {
	Name                 string                   `json:"name,omitempty"`
	Description          *string                  `json:"description,omitempty"`
	CustomInstruction    *string                  `json:"custom_instruction,omitempty"`
	Language             *string                  `json:"language,omitempty"`
	Model                *string                  `json:"model,omitempty"`
	Temperature          *float64                 `json:"temperature,omitempty"`
	MaxTokens            *int                     `json:"max_tokens,omitempty"`
	ThemeColor           *string                  `json:"theme_color,omitempty"`
	WelcomeMessage       *string                  `json:"welcome_message,omitempty"`
	Position             *string                  `json:"position,omitempty"`
	BotMessageColor      *string                  `json:"bot_message_color,omitempty"`
	UserMessageColor     *string                  `json:"user_message_color,omitempty"`
	BotMessageTextColor  *string                  `json:"bot_message_text_color,omitempty"`
	UserMessageTextColor *string                  `json:"user_message_text_color,omitempty"`
	ChatFontFamily       *string                  `json:"chat_font_family,omitempty"`
	ChatHeaderColor      *string                  `json:"chat_header_color,omitempty"`
	ChatHeaderTextColor  *string                  `json:"chat_header_text_color,omitempty"`
	ChatBackgroundColor  *string                  `json:"chat_background_color,omitempty"`
	BubbleRadius         *string                  `json:"bubble_radius,omitempty"`
	InputBackgroundColor *string                  `json:"input_background_color,omitempty"`
	InputTextColor       *string                  `json:"input_text_color,omitempty"`
	SendButtonColor      *string                  `json:"send_button_color,omitempty"`
	BotIcon              *string                  `json:"bot_icon,omitempty"`
	BotDisplayName       *string                  `json:"bot_display_name,omitempty"`
	SecureEmbedEnabled   *bool                    `json:"secure_embed_enabled,omitempty"`
	AllowedDomains       []string                 `json:"allowed_domains,omitempty"`
	EmbedSecret          *string                  `json:"embed_secret,omitempty"`
	SuggestedQuestions   *[]string                `json:"suggested_questions,omitempty"`
	SuggestionsEnabled   *bool                    `json:"suggestions_enabled,omitempty"`
	IncludePaths         *[]string                `json:"include_paths,omitempty"`
	ExcludePaths         *[]string                `json:"exclude_paths,omitempty"`
	SelectorWhitelist    *[]string                `json:"selector_whitelist,omitempty"`
	DiscoveryMode        *string                  `json:"discovery_mode,omitempty"`
	RefreshPolicy        *string                  `json:"refresh_policy,omitempty"`
	RefreshFrequency     *string                  `json:"refresh_frequency,omitempty"`
	HideBranding         *bool                    `json:"hide_branding,omitempty"`
	CustomBranding       *models.CustomBranding   `json:"custom_branding,omitempty"`
	ConfidenceThreshold  *float64                 `json:"confidence_threshold,omitempty"`
	FallbackMessages     *models.FallbackMessages `json:"fallback_messages,omitempty"`
	TopicRestrictions    *models.TopicConfig      `json:"topic_restrictions,omitempty"`
	ThresholdConfig      *models.ThresholdConfig  `json:"threshold_config,omitempty"`
	HandoffEnabled       *bool                    `json:"handoff_enabled,omitempty"`
	HandoffType          *string                  `json:"handoff_type,omitempty"`
	HandoffConfig        *models.HandoffConfig    `json:"handoff_config,omitempty"`
}

// Update validates and applies updates to a chatbot.
func (s *ChatbotService) Update(ctx context.Context, chatbot *models.Chatbot, req ChatbotUpdateRequest) (*models.Chatbot, *validation.FeatureError) {
	// Convert to validation request and validate
	valReq := validation.ChatbotUpdateRequest{
		Model:               req.Model,
		HideBranding:        req.HideBranding,
		CustomBranding:      req.CustomBranding,
		RefreshPolicy:       req.RefreshPolicy,
		DiscoveryMode:       req.DiscoveryMode,
		SecureEmbedEnabled:  req.SecureEmbedEnabled,
		AllowedDomains:      req.AllowedDomains,
		ThresholdConfig:     req.ThresholdConfig,
		HandoffEnabled:      req.HandoffEnabled,
		TopicRestrictions:   req.TopicRestrictions,
		ChatBackgroundColor: req.ChatBackgroundColor,
	}

	if err := s.Validator.ValidateUpdate(ctx, valReq, chatbot.UserID); err != nil {
		return nil, err
	}

	// Apply updates
	s.applyUpdates(chatbot, req)

	// Save to database
	if err := db.UpdateChatbot(ctx, s.DB, chatbot); err != nil {
		if s.Log != nil {
			s.Log.Error("chatbot_update_failed", map[string]any{"chatbot_id": chatbot.ID, "error": err.Error()})
		}
		return nil, &validation.FeatureError{
			Feature: "database",
			Message: "Failed to update chatbot",
		}
	}

	// Re-fetch for updated_at
	updated, err := db.GetChatbotByID(ctx, s.DB, chatbot.ID)
	if err != nil || updated == nil {
		return chatbot, nil // Return original if re-fetch fails
	}

	return updated, nil
}

// =============================================================================
// DOMAIN SPECIFIC UPDATE METHODS
// =============================================================================

type BasicInfoRequest struct {
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	Language          *string `json:"language"`
	CustomInstruction *string `json:"custom_instruction"`
}

func (s *ChatbotService) UpdateBasicInfo(ctx context.Context, bot *models.Chatbot, req BasicInfoRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		Name:              req.Name,
		Description:       req.Description,
		Language:          req.Language,
		CustomInstruction: req.CustomInstruction,
	})
}

type AppearanceRequest struct {
	ThemeColor           *string                `json:"theme_color"`
	WelcomeMessage       *string                `json:"welcome_message"`
	Position             *string                `json:"position"`
	BotMessageColor      *string                `json:"bot_message_color"`
	UserMessageColor     *string                `json:"user_message_color"`
	BotMessageTextColor  *string                `json:"bot_message_text_color"`
	UserMessageTextColor *string                `json:"user_message_text_color"`
	ChatFontFamily       *string                `json:"chat_font_family"`
	ChatHeaderColor      *string                `json:"chat_header_color"`
	ChatHeaderTextColor  *string                `json:"chat_header_text_color"`
	ChatBackgroundColor  *string                `json:"chat_background_color"`
	BubbleRadius         *string                `json:"bubble_radius"`
	InputBackgroundColor *string                `json:"input_background_color"`
	InputTextColor       *string                `json:"input_text_color"`
	SendButtonColor      *string                `json:"send_button_color"`
	BotIcon              *string                `json:"bot_icon"`
	BotDisplayName       *string                `json:"bot_display_name"`
	HideBranding         *bool                  `json:"hide_branding"`
	CustomBranding       *models.CustomBranding `json:"custom_branding"`
	SuggestedQuestions   *[]string              `json:"suggested_questions"`
	SuggestionsEnabled   *bool                  `json:"suggestions_enabled"`
}

func (s *ChatbotService) UpdateAppearance(ctx context.Context, bot *models.Chatbot, req AppearanceRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		ThemeColor:           req.ThemeColor,
		WelcomeMessage:       req.WelcomeMessage,
		Position:             req.Position,
		BotMessageColor:      req.BotMessageColor,
		UserMessageColor:     req.UserMessageColor,
		BotMessageTextColor:  req.BotMessageTextColor,
		UserMessageTextColor: req.UserMessageTextColor,
		ChatFontFamily:       req.ChatFontFamily,
		ChatHeaderColor:      req.ChatHeaderColor,
		ChatHeaderTextColor:  req.ChatHeaderTextColor,
		ChatBackgroundColor:  req.ChatBackgroundColor,
		BubbleRadius:         req.BubbleRadius,
		InputBackgroundColor: req.InputBackgroundColor,
		InputTextColor:       req.InputTextColor,
		SendButtonColor:      req.SendButtonColor,
		BotIcon:              req.BotIcon,
		BotDisplayName:       req.BotDisplayName,
		HideBranding:         req.HideBranding,
		CustomBranding:       req.CustomBranding,
		SuggestedQuestions:   req.SuggestedQuestions,
		SuggestionsEnabled:   req.SuggestionsEnabled,
	})
}

type ModelSettingsRequest struct {
	Model             *string  `json:"model"`
	Temperature       *float64 `json:"temperature"`
	MaxTokens         *int     `json:"max_tokens"`
	CustomInstruction *string  `json:"custom_instruction"`
}

func (s *ChatbotService) UpdateModelSettings(ctx context.Context, bot *models.Chatbot, req ModelSettingsRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		Model:             req.Model,
		Temperature:       req.Temperature,
		MaxTokens:         req.MaxTokens,
		CustomInstruction: req.CustomInstruction,
	})
}

type SecuritySettingsRequest struct {
	SecureEmbedEnabled *bool    `json:"secure_embed_enabled"`
	AllowedDomains     []string `json:"allowed_domains"`
	EmbedSecret        *string  `json:"embed_secret"`
}

func (s *ChatbotService) UpdateSecuritySettings(ctx context.Context, bot *models.Chatbot, req SecuritySettingsRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		SecureEmbedEnabled: req.SecureEmbedEnabled,
		AllowedDomains:     req.AllowedDomains,
		EmbedSecret:        req.EmbedSecret,
	})
}

type GuardrailsRequest struct {
	ConfidenceThreshold *float64                 `json:"confidence_threshold"`
	FallbackMessages    *models.FallbackMessages `json:"fallback_messages"`
	TopicRestrictions   *models.TopicConfig      `json:"topic_restrictions"`
	ThresholdConfig     *models.ThresholdConfig  `json:"threshold_config"`
}

func (s *ChatbotService) UpdateGuardrails(ctx context.Context, bot *models.Chatbot, req GuardrailsRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		ConfidenceThreshold: req.ConfidenceThreshold,
		FallbackMessages:    req.FallbackMessages,
		TopicRestrictions:   req.TopicRestrictions,
		ThresholdConfig:     req.ThresholdConfig,
	})
}

type HandoffRequest struct {
	HandoffEnabled *bool                 `json:"handoff_enabled"`
	HandoffType    *string               `json:"handoff_type"`
	HandoffConfig  *models.HandoffConfig `json:"handoff_config"`
}

func (s *ChatbotService) UpdateHandoff(ctx context.Context, bot *models.Chatbot, req HandoffRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		HandoffEnabled: req.HandoffEnabled,
		HandoffType:    req.HandoffType,
		HandoffConfig:  req.HandoffConfig,
	})
}

type RefreshRequest struct {
	RefreshPolicy    *string `json:"refresh_policy"`
	RefreshFrequency *string `json:"refresh_frequency"`
}

func (s *ChatbotService) UpdateRefresh(ctx context.Context, bot *models.Chatbot, req RefreshRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		RefreshPolicy:    req.RefreshPolicy,
		RefreshFrequency: req.RefreshFrequency,
	})
}

type ScrapingConfigRequest struct {
	IncludePaths      *[]string `json:"include_paths"`
	ExcludePaths      *[]string `json:"exclude_paths"`
	SelectorWhitelist *[]string `json:"selector_whitelist"`
	DiscoveryMode     *string   `json:"discovery_mode"`
}

func (s *ChatbotService) UpdateScrapingConfig(ctx context.Context, bot *models.Chatbot, req ScrapingConfigRequest) (*models.Chatbot, *validation.FeatureError) {
	return s.Update(ctx, bot, ChatbotUpdateRequest{
		IncludePaths:      req.IncludePaths,
		ExcludePaths:      req.ExcludePaths,
		SelectorWhitelist: req.SelectorWhitelist,
		DiscoveryMode:     req.DiscoveryMode,
	})
}

// applyUpdates applies request fields to chatbot model.
func (s *ChatbotService) applyUpdates(c *models.Chatbot, req ChatbotUpdateRequest) {
	if req.Name != "" {
		c.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != nil {
		c.Description = req.Description
	}
	if req.CustomInstruction != nil {
		c.CustomInstruction = *req.CustomInstruction
	}
	if req.Language != nil {
		c.LanguageCode = normalizeLocale(*req.Language)
	}
	if req.Model != nil {
		c.Model = *req.Model
	}
	if req.Temperature != nil {
		c.Temperature = float32(*req.Temperature)
	}
	if req.MaxTokens != nil {
		c.MaxTokens = *req.MaxTokens
	}
	if req.ThemeColor != nil {
		c.ThemeColor = *req.ThemeColor
	}
	if req.WelcomeMessage != nil {
		c.WelcomeMessage = *req.WelcomeMessage
	}
	if req.Position != nil {
		c.Position = *req.Position
	}
	if req.BotMessageColor != nil {
		c.BotMessageColor = *req.BotMessageColor
	}
	if req.UserMessageColor != nil {
		c.UserMessageColor = *req.UserMessageColor
	}
	if req.BotMessageTextColor != nil {
		c.BotMessageTextColor = *req.BotMessageTextColor
	}
	if req.UserMessageTextColor != nil {
		c.UserMessageTextColor = *req.UserMessageTextColor
	}
	if req.ChatFontFamily != nil {
		c.ChatFontFamily = *req.ChatFontFamily
	}
	if req.ChatHeaderColor != nil {
		c.ChatHeaderColor = *req.ChatHeaderColor
	}
	if req.ChatHeaderTextColor != nil {
		c.ChatHeaderTextColor = *req.ChatHeaderTextColor
	}
	if req.ChatBackgroundColor != nil {
		c.ChatBackgroundColor = *req.ChatBackgroundColor
	}
	if req.BubbleRadius != nil {
		c.BubbleRadius = *req.BubbleRadius
	}
	if req.InputBackgroundColor != nil {
		c.InputBackgroundColor = *req.InputBackgroundColor
	}
	if req.InputTextColor != nil {
		c.InputTextColor = *req.InputTextColor
	}
	if req.SendButtonColor != nil {
		c.SendButtonColor = *req.SendButtonColor
	}
	if req.BotIcon != nil {
		c.BotIcon = req.BotIcon
	}
	if req.BotDisplayName != nil {
		c.BotDisplayName = req.BotDisplayName
	}
	if req.SecureEmbedEnabled != nil {
		c.SecureEmbedEnabled = *req.SecureEmbedEnabled
	}
	if req.AllowedDomains != nil {
		joined := strings.Join(req.AllowedDomains, ",")
		c.AllowedDomains = &joined
	}
	if req.EmbedSecret != nil {
		c.EmbedSecret = req.EmbedSecret
	}
	if req.SuggestedQuestions != nil {
		c.SuggestedQuestions = normalizeSuggestions(*req.SuggestedQuestions)
	}
	if req.SuggestionsEnabled != nil {
		c.SuggestionsEnabled = *req.SuggestionsEnabled
	}
	if req.IncludePaths != nil {
		c.IncludePaths = normalizePaths(*req.IncludePaths)
	}
	if req.ExcludePaths != nil {
		c.ExcludePaths = normalizePaths(*req.ExcludePaths)
	}
	if req.SelectorWhitelist != nil {
		c.SelectorWhitelist = normalizeSelectors(*req.SelectorWhitelist)
	}
	if req.DiscoveryMode != nil {
		c.DiscoveryMode = *req.DiscoveryMode
	}
	if req.RefreshPolicy != nil {
		c.RefreshPolicy = *req.RefreshPolicy
		if *req.RefreshPolicy == "auto" && req.RefreshFrequency != nil {
			c.RefreshFrequency = req.RefreshFrequency
			nextRefresh := calculateNextRefreshTime(*req.RefreshFrequency)
			c.NextRefreshAt = &nextRefresh
		} else if *req.RefreshPolicy == "manual" {
			c.NextRefreshAt = nil
		}
	}
	if req.RefreshFrequency != nil && c.RefreshPolicy == "auto" {
		c.RefreshFrequency = req.RefreshFrequency
		nextRefresh := calculateNextRefreshTime(*req.RefreshFrequency)
		c.NextRefreshAt = &nextRefresh
	}
	if req.HideBranding != nil {
		c.HideBranding = *req.HideBranding
		if !*req.HideBranding {
			c.CustomBranding = nil
		}
	}
	if req.CustomBranding != nil {
		c.CustomBranding = req.CustomBranding
	}
	if req.ConfidenceThreshold != nil {
		c.ConfidenceThreshold = *req.ConfidenceThreshold
	}
	if req.FallbackMessages != nil {
		c.FallbackMessages = req.FallbackMessages
	}
	if req.TopicRestrictions != nil {
		c.TopicRestrictions = req.TopicRestrictions
	}
	if req.ThresholdConfig != nil {
		c.ThresholdConfig = req.ThresholdConfig
	}
	if req.HandoffEnabled != nil {
		c.HandoffEnabled = *req.HandoffEnabled
	}
	if req.HandoffType != nil {
		c.HandoffType = *req.HandoffType
	}
	if req.HandoffConfig != nil {
		c.HandoffConfig = req.HandoffConfig
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func normalizeLocale(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "tr-TR"
	}
	return s
}

func normalizeSuggestions(items []string) []string {
	var result []string
	for _, item := range items {
		s := strings.TrimSpace(item)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func normalizePaths(items []string) []string {
	var result []string
	for _, item := range items {
		s := strings.TrimSpace(item)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func normalizeSelectors(items []string) []string {
	var result []string
	for _, item := range items {
		s := strings.TrimSpace(item)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func calculateNextRefreshTime(frequency string) time.Time {
	now := time.Now()
	switch frequency {
	case "daily":
		next := now.Add(24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, now.Location())
	case "weekly":
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		next := now.Add(time.Duration(daysUntilSunday) * 24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, now.Location())
	case "monthly":
		next := now.AddDate(0, 1, 0)
		return time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		next := now.Add(time.Duration(daysUntilSunday) * 24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, now.Location())
	}
}
