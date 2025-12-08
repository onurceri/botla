package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// HandoffHandlers handles handoff-related HTTP endpoints
type HandoffHandlers struct {
	DB  *sql.DB
	Log *logger.Logger
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
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[3]

	// Parse request
	var req publicHandoffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "session_id is required"})
		return
	}

	// Get chatbot
	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil || bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if handoff is enabled
	if !bot.HandoffEnabled {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "handoff is not enabled for this chatbot"})
		return
	}

	// Get or create conversation
	conv, err := db.GetOrCreateConversationBySessionID(r.Context(), h.DB, botID, req.SessionID)
	if err != nil || conv == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create handoff service and request
	svc := services.NewHandoffService(h.DB, h.Log)
	result, err := svc.RequestHandoff(r.Context(), bot, conv.ID, req.Message)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("handoff_request_failed", map[string]any{"error": err.Error()})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create handoff request"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// ListHandoffRequests handles GET /api/chatbots/:id/handoff-requests
func (h *HandoffHandlers) ListHandoffRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract chatbot ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[3]

	// Verify ownership
	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil || bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if bot.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Get handoff requests
	svc := services.NewHandoffService(h.DB, h.Log)
	requests, err := svc.GetHandoffRequests(r.Context(), botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"requests": requests})
}

// UpdateHandoffRequest handles PATCH /api/chatbots/:id/handoff-requests/:requestId
func (h *HandoffHandlers) UpdateHandoffRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract IDs from path: /api/chatbots/:id/handoff-requests/:requestId
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[3]
	requestID := parts[5]

	// Verify ownership
	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil || bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if bot.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Parse request
	var req handoffStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Update status
	svc := services.NewHandoffService(h.DB, h.Log)
	if err := svc.UpdateHandoffStatus(r.Context(), requestID, req.Status, req.AssignedTo); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
