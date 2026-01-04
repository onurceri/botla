package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/validation"
	"github.com/onurceri/botla-app/pkg/middleware"
)

// ByID handles GET/PUT/DELETE for a specific chatbot
func (h *ChatbotHandlers) ByID(w http.ResponseWriter, r *http.Request) {
	c, botID, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getChatbot(w, c)
	case http.MethodPut:
		h.updateChatbot(w, r, c, botID)
	case http.MethodDelete:
		h.deleteChatbot(w, r, botID, userIDFromContext(r))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func userIDFromContext(r *http.Request) string {
	userID, _ := middleware.UserIDFromContext(r.Context())
	return userID
}

// getChatbot returns a single chatbot
func (h *ChatbotHandlers) getChatbot(w http.ResponseWriter, c *models.Chatbot) {
	api.WriteJSON(w, http.StatusOK, c)
}

// updateChatbot handles PUT request to update a chatbot
func (h *ChatbotHandlers) updateChatbot(w http.ResponseWriter, r *http.Request, c *models.Chatbot, botID string) {
	var req createChatbotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// If ChatbotService is available, use it for validation and update
	if h.ChatbotService != nil {
		serviceReq := h.convertToServiceRequest(req)
		updated, featureErr := h.ChatbotService.Update(r.Context(), c, serviceReq)
		if featureErr != nil {
			h.writeFeatureError(w, featureErr)
			return
		}
		api.WriteJSON(w, http.StatusOK, updated)
		return
	}

	if err := h.ChatbotRepo.Update(r.Context(), c); err != nil {
		if h.Logger != nil {
			h.Logger.Error("UpdateChatbot failed", map[string]any{"error": err, "chatbot_id": botID})
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c2, err := h.ChatbotRepo.GetByID(r.Context(), botID)
	if err != nil || c2 == nil {
		if h.Logger != nil {
			h.Logger.Error("GetChatbotByID after update failed", map[string]any{"error": err, "chatbot_id": botID, "result_nil": c2 == nil})
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, c2)
}

// convertToServiceRequest converts handler request to service request
func (h *ChatbotHandlers) convertToServiceRequest(req createChatbotRequest) services.ChatbotUpdateRequest {
	result := services.ChatbotUpdateRequest{
		Name:                 req.Name,
		Description:          req.Description,
		CustomInstruction:    req.CustomInstruction,
		Language:             req.Language,
		Model:                req.Model,
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
		SecureEmbedEnabled:   req.SecureEmbedEnabled,
		EmbedSecret:          req.EmbedSecret,
		SuggestedQuestions:   req.SuggestedQuestions,
		SuggestionsEnabled:   req.SuggestionsEnabled,
		IncludePaths:         req.IncludePaths,
		ExcludePaths:         req.ExcludePaths,
		SelectorWhitelist:    req.SelectorWhitelist,
		DiscoveryMode:        req.DiscoveryMode,
		RefreshPolicy:        req.RefreshPolicy,
		RefreshFrequency:     req.RefreshFrequency,
		HideBranding:         req.HideBranding,
		CustomBranding:       req.CustomBranding,
		ConfidenceThreshold:  req.ConfidenceThreshold,
		FallbackMessages:     req.FallbackMessages,
		ThresholdConfig:      req.ThresholdConfig,
		HandoffEnabled:       req.HandoffEnabled,
		HandoffType:          req.HandoffType,
		HandoffConfig:        req.HandoffConfig,
	}

	// Handle temperature conversion (float32 -> float64)
	if req.Temperature != nil {
		temp := float64(*req.Temperature)
		result.Temperature = &temp
	}

	// Handle max_tokens
	if req.MaxTokens != nil {
		result.MaxTokens = req.MaxTokens
	}

	// Handle allowed_domains (string -> []string)
	if req.AllowedDomains != nil && *req.AllowedDomains != "" {
		domains := strings.Split(*req.AllowedDomains, ",")
		var cleaned []string
		for _, d := range domains {
			d = strings.TrimSpace(d)
			if d != "" {
				cleaned = append(cleaned, d)
			}
		}
		result.AllowedDomains = cleaned
	}

	// Handle topic_restrictions - no conversion needed as types match
	if req.TopicRestrictions != nil {
		result.TopicRestrictions = req.TopicRestrictions
	}

	return result
}

// writeFeatureError writes a feature error response
func (h *ChatbotHandlers) writeFeatureError(w http.ResponseWriter, err *validation.FeatureError) {
	api.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
		"error":            err.Message,
		"upgrade_required": err.UpgradeRequired,
		"feature":          err.Feature,
	})
}

// deleteChatbot handles DELETE request
func (h *ChatbotHandlers) deleteChatbot(w http.ResponseWriter, r *http.Request, botID, userID string) {
	sourceIDs, err := h.ChatbotRepo.SoftDelete(r.Context(), botID, userID)
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
