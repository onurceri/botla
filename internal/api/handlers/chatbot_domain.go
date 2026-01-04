package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/services"
)

// UpdateBasicInfo handles PUT /chatbots/{id}/basic-info
func (h *ChatbotHandlers) UpdateBasicInfo(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.BasicInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateBasicInfo(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// UpdateAppearance handles PUT /chatbots/{id}/appearance
func (h *ChatbotHandlers) UpdateAppearance(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.AppearanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateAppearance(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// UpdateModelSettings handles PUT /chatbots/{id}/model
func (h *ChatbotHandlers) UpdateModelSettings(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.ModelSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateModelSettings(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// UpdateSecuritySettings handles PUT /chatbots/{id}/security
func (h *ChatbotHandlers) UpdateSecuritySettings(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.SecuritySettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateSecuritySettings(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// UpdateGuardrails handles PUT /chatbots/{id}/guardrails
func (h *ChatbotHandlers) UpdateGuardrails(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.GuardrailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateGuardrails(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// UpdateHandoff handles PUT /chatbots/{id}/handoff
func (h *ChatbotHandlers) UpdateHandoff(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.HandoffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateHandoff(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// UpdateRefresh handles PUT /chatbots/{id}/refresh
func (h *ChatbotHandlers) UpdateRefresh(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateRefresh(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// UpdateScrapingConfig handles PUT /chatbots/{id}/scraping
func (h *ChatbotHandlers) UpdateScrapingConfig(w http.ResponseWriter, r *http.Request) {
	c, _, ok := h.getChatbotFromContext(w, r)
	if !ok {
		return
	}

	var req services.ScrapingConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := h.ChatbotService.UpdateScrapingConfig(r.Context(), c, req)
	if err != nil {
		h.writeFeatureError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, updated)
}

// Helper to get chatbot from context and verify ownership
// Returns chatbot, botID, and bool indicating success (if false, response is already written)
func (h *ChatbotHandlers) getChatbotFromContext(w http.ResponseWriter, r *http.Request) (*models.Chatbot, string, bool) {
	return getChatbotContext(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
}
