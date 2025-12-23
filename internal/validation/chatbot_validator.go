package validation

import (
	"context"
	"database/sql"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
)

// =============================================================================
// CHATBOT VALIDATOR - Plan-based feature validation
// =============================================================================

// ChatbotValidator validates chatbot update requests against plan restrictions.
type ChatbotValidator struct {
	DB *sql.DB
}

// NewChatbotValidator creates a new ChatbotValidator.
func NewChatbotValidator(db *sql.DB) *ChatbotValidator {
	return &ChatbotValidator{DB: db}
}

// FeatureError represents a plan-based feature restriction error.
type FeatureError struct {
	Feature         string `json:"feature"`
	Message         string `json:"error"`
	UpgradeRequired bool   `json:"upgrade_required"`
}

func (e *FeatureError) Error() string {
	return e.Message
}

// ChatbotUpdateRequest represents fields that can be updated on a chatbot.
// This is a subset focused on fields that require plan validation.
type ChatbotUpdateRequest struct {
	Model               *string                 `json:"model,omitempty"`
	HideBranding        *bool                   `json:"hide_branding,omitempty"`
	CustomBranding      *models.CustomBranding  `json:"custom_branding,omitempty"`
	RefreshPolicy       *string                 `json:"refresh_policy,omitempty"`
	DiscoveryMode       *string                 `json:"discovery_mode,omitempty"`
	SecureEmbedEnabled  *bool                   `json:"secure_embed_enabled,omitempty"`
	AllowedDomains      []string                `json:"allowed_domains,omitempty"`
	ThresholdConfig     *models.ThresholdConfig `json:"threshold_config,omitempty"`
	HandoffEnabled      *bool                   `json:"handoff_enabled,omitempty"`
	TopicRestrictions   *models.TopicConfig     `json:"topic_restrictions,omitempty"`
	ChatBackgroundColor *string                 `json:"chat_background_color,omitempty"`
	MaxTokens           *int                    `json:"max_tokens,omitempty"`
}

// ValidateUpdate checks all plan-based restrictions for a chatbot update.
// It fetches the plan ONCE and validates all fields against it.
func (v *ChatbotValidator) ValidateUpdate(ctx context.Context, req ChatbotUpdateRequest, userID string) *FeatureError {
	// Fetch plan once for all validations
	plan, err := db.GetPlanByUserID(ctx, v.DB, userID)
	if err != nil || plan == nil {
		return &FeatureError{
			Feature:         "plan",
			Message:         "Could not verify plan",
			UpgradeRequired: false,
		}
	}

	// Validate each feature against plan restrictions
	if err := v.validateModel(req, plan); err != nil {
		return err
	}
	if err := v.validateBranding(req, plan); err != nil {
		return err
	}
	if err := v.validateRefresh(req, plan); err != nil {
		return err
	}
	if err := v.validateDiscovery(req, plan); err != nil {
		return err
	}
	if err := v.validateSecureEmbed(req, plan); err != nil {
		return err
	}
	if err := v.validateGuardrails(req, plan); err != nil {
		return err
	}
	if err := v.validateHandoff(req, plan); err != nil {
		return err
	}
	if err := v.validateTopicRestrictions(req, plan); err != nil {
		return err
	}
	if err := v.validateColors(req); err != nil {
		return err
	}
	if err := v.validateMaxTokens(req, plan); err != nil {
		return err
	}

	return nil
}

// =============================================================================
// INDIVIDUAL VALIDATORS
// =============================================================================

func (v *ChatbotValidator) validateMaxTokens(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.MaxTokens == nil {
		return nil
	}

	val := *req.MaxTokens
	minLimit := plan.Config.Chat.MinResponseTokenLimit
	maxLimit := plan.Config.Chat.MaxResponseTokenLimit

	// Default fallbacks if not set in plan (backward compatibility)
	if minLimit <= 0 {
		minLimit = 20
	}
	if maxLimit <= 0 {
		maxLimit = 8192 // Default reasonable max
	}

	if val < minLimit {
		return &FeatureError{
			Feature:         "max_tokens",
			Message:         "Value is below the minimum allowed limit for your plan",
			UpgradeRequired: false,
		}
	}
	if val > maxLimit {
		return &FeatureError{
			Feature:         "max_tokens",
			Message:         "Value exceeds the maximum allowed limit for your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateModel(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.Model == nil {
		return nil
	}
	if len(plan.Config.Chat.AllowedModels) > 0 {
		allowed := false
		for _, m := range plan.Config.Chat.AllowedModels {
			if m == *req.Model || strings.HasSuffix(*req.Model, m) {
				allowed = true
				break
			}
		}
		if !allowed {
			return &FeatureError{
				Feature:         "model_selection",
				Message:         "This model is not available on your plan",
				UpgradeRequired: true,
			}
		}
	}
	return nil
}

func (v *ChatbotValidator) validateBranding(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.HideBranding != nil && *req.HideBranding && !plan.Config.Branding.CanHideBranding {
		return &FeatureError{
			Feature:         "hide_branding",
			Message:         "Your plan does not allow hiding branding",
			UpgradeRequired: true,
		}
	}
	if req.CustomBranding != nil && !plan.Config.Branding.CanCustomBranding {
		return &FeatureError{
			Feature:         "custom_branding",
			Message:         "Custom branding requires Enterprise plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateRefresh(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.RefreshPolicy != nil && *req.RefreshPolicy == "auto" && !plan.Config.Refresh.Enabled {
		return &FeatureError{
			Feature:         "auto_refresh",
			Message:         "Auto refresh is not available on your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateDiscovery(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.DiscoveryMode != nil && *req.DiscoveryMode != "disabled" && plan.Config.Scraping.MaxPagesPerCrawl <= 0 {
		return &FeatureError{
			Feature:         "discovery_mode",
			Message:         "Discovery mode is not available on your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateSecureEmbed(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.SecureEmbedEnabled != nil && *req.SecureEmbedEnabled && !plan.Config.Security.SecureEmbedEnabled {
		return &FeatureError{
			Feature:         "secure_embed",
			Message:         "Secure embed is not available on your plan",
			UpgradeRequired: true,
		}
	}
	if len(req.AllowedDomains) > 0 && !plan.Config.Security.SecureEmbedEnabled {
		return &FeatureError{
			Feature:         "secure_embed",
			Message:         "Domain restrictions require secure embed feature",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateGuardrails(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.ThresholdConfig == nil {
		return nil
	}

	tc := req.ThresholdConfig
	guardrails := plan.Config.Guardrails

	// Check threshold customization
	if (tc.HighThreshold != 0 || tc.MediumThreshold != 0 || tc.ShowConfidenceWarning) && !guardrails.CanCustomizeThresholds {
		return &FeatureError{
			Feature:         "thresholds",
			Message:         "Threshold customization is not available on your plan",
			UpgradeRequired: true,
		}
	}

	// Check smart fallback
	if tc.FallbackMode == "smart" && !guardrails.CanUseSmartFallback {
		return &FeatureError{
			Feature:         "smart_fallback",
			Message:         "Smart fallback is not available on your plan",
			UpgradeRequired: true,
		}
	}

	// Check escalate fallback
	if tc.FallbackMode == "escalate" && !guardrails.CanUseEscalateFallback {
		return &FeatureError{
			Feature:         "escalate_fallback",
			Message:         "Escalate fallback is not available on your plan",
			UpgradeRequired: true,
		}
	}

	return nil
}

func (v *ChatbotValidator) validateHandoff(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.HandoffEnabled != nil && *req.HandoffEnabled && !plan.Config.Guardrails.CanUseEscalateFallback {
		return &FeatureError{
			Feature:         "escalate_fallback",
			Message:         "Human handoff is not available on your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateTopicRestrictions(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.TopicRestrictions != nil && !plan.Config.Guardrails.CanManageTopics {
		return &FeatureError{
			Feature:         "topic_restrictions",
			Message:         "Topic restrictions are not available on your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateColors(req ChatbotUpdateRequest) *FeatureError {
	if req.ChatBackgroundColor != nil {
		s := strings.TrimSpace(*req.ChatBackgroundColor)
		if s != "" && !isValidHexColor(s) {
			return &FeatureError{
				Feature:         "color",
				Message:         "Invalid hex color format",
				UpgradeRequired: false,
			}
		}
	}
	return nil
}

// isValidHexColor checks if a string is a valid hex color.
func isValidHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
