package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// ByID handles GET/PUT/DELETE for a specific chatbot
func (h *ChatbotHandlers) ByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	botID, ok := parseBotIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if botID == "new" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil {
		log.Printf("[ERROR] ByID GetChatbotByID failed: botID=%s, err=%v", botID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if c.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getChatbot(w, c)
	case http.MethodPut:
		h.updateChatbot(w, r, c, botID)
	case http.MethodDelete:
		h.deleteChatbot(w, r, botID, userID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// getChatbot returns a single chatbot
func (h *ChatbotHandlers) getChatbot(w http.ResponseWriter, c *models.Chatbot) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(c); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// updateChatbot handles PUT request to update a chatbot
func (h *ChatbotHandlers) updateChatbot(w http.ResponseWriter, r *http.Request, c *models.Chatbot, botID string) {
	var req createChatbotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate hex color if provided
	if req.ChatBackgroundColor != nil {
		s := strings.TrimSpace(*req.ChatBackgroundColor)
		if s != "" && !isValidHexColor(s) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Validate branding permissions based on user's plan
	if req.HideBranding != nil && *req.HideBranding {
		plan, err := db.GetPlanByUserID(r.Context(), h.DB, c.UserID)
		if err != nil || plan == nil || !plan.Config.Branding.CanHideBranding {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":            "Your plan does not allow hiding branding",
				"upgrade_required": true,
				"feature":          "hide_branding",
			})
			return
		}
	}

	if req.CustomBranding != nil {
		plan, err := db.GetPlanByUserID(r.Context(), h.DB, c.UserID)
		if err != nil || plan == nil || !plan.Config.Branding.CanCustomBranding {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":            "Custom branding requires Enterprise plan",
				"upgrade_required": true,
				"feature":          "custom_branding",
			})
			return
		}
	}

	if req.RefreshPolicy != nil && *req.RefreshPolicy == "auto" {
		plan, err := db.GetPlanByUserID(r.Context(), h.DB, c.UserID)
		if err != nil || plan == nil || !plan.Config.Refresh.Enabled {
			base := api.BaseLang(c.LanguageCode)
			cfg := api.ConfigFromBase(base)
			api.WriteLocalizedError(w, http.StatusForbidden, api.ErrPlanRefreshUnavailable, cfg)
			return
		}
	}

	if req.DiscoveryMode != nil && *req.DiscoveryMode != "disabled" {
		plan, err := db.GetPlanByUserID(r.Context(), h.DB, c.UserID)
		if err != nil || plan == nil || plan.Config.Scraping.MaxPagesPerCrawl <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":            "Discovery mode is not available on your plan",
				"upgrade_required": true,
			})
			return
		}
	}

	if req.SecureEmbedEnabled != nil && *req.SecureEmbedEnabled {
		plan, err := db.GetPlanByUserID(r.Context(), h.DB, c.UserID)
		if err != nil || plan == nil || !plan.Config.Security.SecureEmbedEnabled {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":            "Secure embed is not available on your plan",
				"upgrade_required": true,
				"feature":          "secure_embed",
			})
			return
		}
	}

	// Apply updates
	// If branding is explicitly turned off, clear any custom branding
	if req.HideBranding != nil && !*req.HideBranding {
		req.CustomBranding = nil
	}
	applyChatbotUpdates(c, req)

	if err := db.UpdateChatbot(r.Context(), h.DB, c); err != nil {
		log.Printf("[ERROR] UpdateChatbot failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Re-read for updated_at change
	c2, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil || c2 == nil {
		log.Printf("[ERROR] GetChatbotByID after update failed: err=%v, c2=%v", err, c2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(c2); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// deleteChatbot handles DELETE request
func (h *ChatbotHandlers) deleteChatbot(w http.ResponseWriter, r *http.Request, botID, userID string) {
	sourceIDs, err := db.SoftDeleteChatbot(r.Context(), h.DB, botID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Clean up vectors if any sources were deleted
	if len(sourceIDs) > 0 && h.VectorStore != nil {
		// Best effort cleanup - run in background or parallel if many sources
		// For now, simple loop
		for _, sid := range sourceIDs {
			_ = h.VectorStore.DeleteBySourceID(r.Context(), sid)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// applyChatbotUpdates applies request fields to chatbot model
func applyChatbotUpdates(c *models.Chatbot, req createChatbotRequest) {
	if req.Name != "" {
		c.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != nil {
		c.Description = req.Description
	}
	if req.SystemPrompt != nil {
		c.SystemPrompt = *req.SystemPrompt
	}
	if req.Language != nil {
		c.LanguageCode = normalizeLocale(*req.Language)
	}
	if req.Model != nil {
		c.Model = *req.Model
	}
	if req.Temperature != nil {
		c.Temperature = *req.Temperature
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
		c.AllowedDomains = req.AllowedDomains
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
		// If switching to auto, calculate next refresh time
		if *req.RefreshPolicy == "auto" && req.RefreshFrequency != nil {
			c.RefreshFrequency = req.RefreshFrequency
			nextRefresh := calculateNextRefresh(*req.RefreshFrequency)
			c.NextRefreshAt = &nextRefresh
		} else if *req.RefreshPolicy == "manual" {
			// Clear next refresh when switching to manual
			c.NextRefreshAt = nil
		}
	}
	if req.RefreshFrequency != nil && c.RefreshPolicy == "auto" {
		c.RefreshFrequency = req.RefreshFrequency
		nextRefresh := calculateNextRefresh(*req.RefreshFrequency)
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

// calculateNextRefresh calculates the next refresh time based on frequency
func calculateNextRefresh(frequency string) time.Time {
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
		// Default to weekly
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		next := now.Add(time.Duration(daysUntilSunday) * 24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, now.Location())
	}
}
