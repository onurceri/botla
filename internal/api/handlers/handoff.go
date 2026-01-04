package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/httputil"
	"github.com/onurceri/botla-app/pkg/logger"
)

// HandoffHandlers handles handoff-related HTTP endpoints
type HandoffHandlers struct {
	DB               *sql.DB
	Log              *logger.Logger
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
	HandoffService   *services.HandoffService
	ChatbotRepo      repository.ChatbotRepository
	ConversationRepo repository.ConversationRepository
	HandoffRepo      repository.HandoffRepository
}

// publicHandoffRequest represents a handoff request from the widget
type publicHandoffRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message,omitempty"`
}

// handoffStatusUpdate represents a status update request
type handoffStatusUpdate struct {
	Status     string  `json:"status"`
	AssignedTo *string `json:"assigned_to,omitempty"`
}

// PublicRequestHandoff handles POST /api/public/:botId/handoff
// This is called from the widget when a user requests human support
func (h *HandoffHandlers) PublicRequestHandoff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract bot ID from path
	const prefix = "/api/v1/public/chatbots/"
	botID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), "/handoff")
	if botID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !httputil.IsValidUUID(botID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	// Parse request
	var req publicHandoffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id is required"})
		return
	}

	// Get chatbot
	bot, err := h.ChatbotRepo.GetByID(r.Context(), botID)
	if err != nil || bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if handoff is enabled
	if !bot.HandoffEnabled {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrHandoffNotEnabled)
		return
	}

	// Get or create conversation
	conv, err := h.ConversationRepo.GetOrCreateBySessionID(r.Context(), botID, req.SessionID)
	if err != nil || conv == nil {
		if h.Log != nil {
			errText := ""
			if err != nil {
				errText = err.Error()
			}
			h.Log.Error("handoff_conversation_create_failed", map[string]any{"error": errText, "bot_id": botID})
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Create handoff request using injected service
	result, err := h.HandoffService.RequestHandoff(r.Context(), bot, conv.ID, req.Message)
	if err != nil {
		if status, code, ok := api.MapHandoffError(err); ok {
			api.WriteErrorCode(w, status, code)
			return
		}
		if h.Log != nil {
			h.Log.Error("handoff_request_failed", map[string]any{"error": err.Error(), "bot_id": botID, "conversation_id": conv.ID})
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, result)
}

// ListHandoffRequests handles GET /api/chatbots/:id/handoff-requests
func (h *HandoffHandlers) ListHandoffRequests(w http.ResponseWriter, r *http.Request) {
	_, botID, ok := getChatbotContext(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Get handoff requests using injected service
	requests, err := h.HandoffService.GetHandoffRequests(r.Context(), botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{"requests": requests})
}

// UpdateHandoffRequest handles PATCH /api/chatbots/:id/handoff-requests/:requestId
func (h *HandoffHandlers) UpdateHandoffRequest(w http.ResponseWriter, r *http.Request) {
	_, _, ok := getChatbotContext(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Extract IDs from path: /api/v1/chatbots/:id/handoff-requests/:requestId
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestID := parts[6]
	if !httputil.IsValidUUID(requestID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	// Parse request
	var req handoffStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidRequestBody)
		return
	}

	// Update status using injected service
	if err := h.HandoffService.UpdateHandoffStatus(r.Context(), requestID, req.Status, req.AssignedTo); err != nil {
		var invalid services.InvalidHandoffStatusError
		if errors.As(err, &invalid) {
			api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidStatus)
			return
		}
		if status, code, ok := api.MapHandoffError(err); ok {
			api.WriteErrorCode(w, status, code)
			return
		}
		if h.Log != nil {
			h.Log.Error("handoff_update_status_failed", map[string]any{"error": err.Error(), "request_id": requestID})
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// publicEmailSubmission represents an email submission from the widget
type publicEmailSubmission struct {
	Email string `json:"email"`
}

// PublicSubmitEmail handles POST /api/v1/public/chatbots/:botId/handoff/:requestId/contact
// This allows users to submit their email after handoff is triggered
func (h *HandoffHandlers) PublicSubmitEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract IDs from path: /api/v1/public/chatbots/:botId/handoff/:requestId/contact
	parts := strings.Split(r.URL.Path, "/")
	// Expected: ["", "api", "v1", "public", "chatbots", ":botId", "handoff", ":requestId", "contact"]
	if len(parts) < 9 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[5]
	requestID := parts[7]
	if !httputil.IsValidUUID(botID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	// Parse request
	var req publicEmailSubmission
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate email format using proper parsing
	if _, err := mail.ParseAddress(req.Email); err != nil {
		api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "valid email is required"})
		return
	}

	// Verify bot exists
	bot, err := h.ChatbotRepo.GetByID(r.Context(), botID)
	if err != nil || bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Update handoff request with email
	if err := h.HandoffRepo.UpdateHandoffRequestStatus(r.Context(), requestID, "pending", nil); err != nil {
		api.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save email"})
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Teşekkürler! E-posta adresiniz kaydedildi. En kısa sürede sizinle iletişime geçeceğiz.",
	})
}

// GetHandoffRequestDetail handles GET /api/v1/chatbots/:id/handoff-requests/:requestId
// Returns the handoff request with full conversation history
func (h *HandoffHandlers) GetHandoffRequestDetail(w http.ResponseWriter, r *http.Request) {
	_, _, ok := getChatbotContext(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Extract IDs from path: /api/v1/chatbots/:id/handoff-requests/:requestId
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestID := parts[6]
	if !httputil.IsValidUUID(requestID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	// Get request with messages
	handoffReq, err := h.HandoffRepo.GetHandoffRequestByID(r.Context(), requestID)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("get_handoff_detail_failed", map[string]any{"error": err.Error(), "request_id": requestID})
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if handoffReq == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrHandoffNotFound)
		return
	}

	// Get conversation messages
	messages, err := h.HandoffRepo.ListHandoffMessages(r.Context(), handoffReq.ConversationID, 100)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("get_handoff_messages_failed", map[string]any{"error": err.Error(), "request_id": requestID})
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Build combined response
	detail := struct {
		Request  *models.HandoffRequest `json:"request"`
		Messages []models.Message       `json:"messages"`
	}{
		Request:  handoffReq,
		Messages: messages,
	}

	api.WriteJSON(w, http.StatusOK, detail)
}
