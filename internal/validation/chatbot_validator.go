package validation

import (
	"context"
	"strings"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

// =============================================================================
// CHATBOT VALIDATOR - Plan-based feature validation
// =============================================================================

// ChatbotValidator validates chatbot update requests against plan restrictions.
type ChatbotValidator struct {
	planRepo repository.PlanRepository
}

// NewChatbotValidator creates a new ChatbotValidator.
func NewChatbotValidator(planRepo repository.PlanRepository) *ChatbotValidator {
	return &ChatbotValidator{planRepo: planRepo}
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
	ManualQuestions     *[]string               `json:"manual_questions,omitempty"`
}

// ValidateUpdate checks all plan-based restrictions for a chatbot update.
// It fetches the plan ONCE and validates all fields against it.
func (v *ChatbotValidator) ValidateUpdate(ctx context.Context, req ChatbotUpdateRequest, userID string) *FeatureError {
	// Fetch plan once for all validations
	plan, err := v.planRepo.GetByUserID(ctx, userID)
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
	if err := v.validateManualQuestions(req, plan); err != nil {
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
	minLimit := plan.Limits.ChatMinResponseTokenLimit
	maxLimit := plan.Limits.ChatMaxResponseTokenLimit

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

func (v *ChatbotValidator) validateManualQuestions(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.ManualQuestions == nil {
		return nil
	}

	maxLimit := plan.Limits.ChatMaxManualQuestions
	// Default fallback if not set in plan (backward compatibility)
	if maxLimit <= 0 {
		maxLimit = 3 // Default to free plan limit
	}

	if len(*req.ManualQuestions) > maxLimit {
		return &FeatureError{
			Feature:         "manual_questions",
			Message:         "You have exceeded the maximum number of manual questions for your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateModel(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.Model == nil {
		return nil
	}
	if len(plan.Limits.ChatAllowedModels) > 0 {
		allowed := false
		for _, m := range plan.Limits.ChatAllowedModels {
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
	if req.HideBranding != nil && *req.HideBranding && !plan.Limits.BrandingCanHideBranding {
		return &FeatureError{
			Feature:         "hide_branding",
			Message:         "Your plan does not allow hiding branding",
			UpgradeRequired: true,
		}
	}
	if req.CustomBranding != nil && !plan.Limits.BrandingCanCustomBranding {
		return &FeatureError{
			Feature:         "custom_branding",
			Message:         "Custom branding requires Enterprise plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateRefresh(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.RefreshPolicy != nil && *req.RefreshPolicy == "auto" && !plan.Limits.RefreshEnabled {
		return &FeatureError{
			Feature:         "auto_refresh",
			Message:         "Auto refresh is not available on your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateDiscovery(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.DiscoveryMode != nil && *req.DiscoveryMode != "disabled" && plan.Limits.ScrapingMaxPagesPerCrawl <= 0 {
		return &FeatureError{
			Feature:         "discovery_mode",
			Message:         "Discovery mode is not available on your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateSecureEmbed(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.SecureEmbedEnabled != nil && *req.SecureEmbedEnabled && !plan.Limits.SecuritySecureEmbedEnabled {
		return &FeatureError{
			Feature:         "secure_embed",
			Message:         "Secure embed is not available on your plan",
			UpgradeRequired: true,
		}
	}
	if len(req.AllowedDomains) > 0 && !plan.Limits.SecuritySecureEmbedEnabled {
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
	limits := plan.Limits

	// Check threshold customization
	if (tc.HighThreshold != 0 || tc.MediumThreshold != 0 || tc.ShowConfidenceWarning) && !limits.GuardrailsCanCustomizeThresholds {
		return &FeatureError{
			Feature:         "thresholds",
			Message:         "Threshold customization is not available on your plan",
			UpgradeRequired: true,
		}
	}

	// Check smart fallback
	if tc.FallbackMode == "smart" && !limits.GuardrailsCanUseSmartFallback {
		return &FeatureError{
			Feature:         "smart_fallback",
			Message:         "Smart fallback is not available on your plan",
			UpgradeRequired: true,
		}
	}

	// Check escalate fallback
	if tc.FallbackMode == "escalate" && !limits.GuardrailsCanUseEscalateFallback {
		return &FeatureError{
			Feature:         "escalate_fallback",
			Message:         "Escalate fallback is not available on your plan",
			UpgradeRequired: true,
		}
	}

	return nil
}

func (v *ChatbotValidator) validateHandoff(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.HandoffEnabled != nil && *req.HandoffEnabled && !plan.Limits.GuardrailsCanUseEscalateFallback {
		return &FeatureError{
			Feature:         "escalate_fallback",
			Message:         "Human handoff is not available on your plan",
			UpgradeRequired: true,
		}
	}
	return nil
}

func (v *ChatbotValidator) validateTopicRestrictions(req ChatbotUpdateRequest, plan *models.Plan) *FeatureError {
	if req.TopicRestrictions != nil && !plan.Limits.GuardrailsCanManageTopics {
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
		if s != "" && !isValidColor(s) {
			return &FeatureError{
				Feature:         "color",
				Message:         "Invalid color format (must be HEX or RGBA)",
				UpgradeRequired: false,
			}
		}
	}
	return nil
}

// isValidColor checks if a string is a valid color (HEX or RGBA).
func isValidColor(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	// Check for HEX
	if strings.HasPrefix(s, "#") {
		if len(s) != 7 && len(s) != 4 && len(s) != 9 {
			return false
		}
		for i := 1; i < len(s); i++ {
			c := s[i]
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				return false
			}
		}
		return true
	}

	// Check for RGBA
	if strings.HasPrefix(strings.ToLower(s), "rgba(") && strings.HasSuffix(s, ")") {
		return true // Simplified check for now
	}

	return false
}
