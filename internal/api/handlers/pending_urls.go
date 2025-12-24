package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/logger"
)

// PendingURLsHandlers handles pending URL operations
type PendingURLsHandlers struct {
	DB               *sql.DB
	Queue            *processing.SourceQueue
	Log              *logger.Logger
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
}

// PendingURLResponse represents a pending URL in the response
type PendingURLResponse struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	DiscoveredAt string `json:"discovered_at"`
}

// ListPendingURLsResponse is the response for listing pending URLs
type ListPendingURLsResponse struct {
	URLs    []PendingURLResponse `json:"urls"`
	Total   int                  `json:"total"`
	Page    int                  `json:"page"`
	PerPage int                  `json:"per_page"`
}

// ApproveRejectRequest is the request body for approve/reject operations
type ApproveRejectRequest struct {
	URLIDs []string `json:"url_ids"`
}

// ApproveResponse is the response for approve operation
type ApproveResponse struct {
	ApprovedCount  int `json:"approved_count"`
	SourcesCreated int `json:"sources_created"`
}

// RejectResponse is the response for reject operation
type RejectResponse struct {
	RejectedCount int `json:"rejected_count"`
}

// ClearResponse is the response for clear operation
type ClearResponse struct {
	ClearedCount int `json:"cleared_count"`
}

// ListPendingURLs handles GET /api/v1/chatbots/:id/pending-urls
func (h *PendingURLsHandlers) ListPendingURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	_, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Parse pagination
	page := 1
	perPage := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err2 := strconv.Atoi(p); err2 == nil && parsed > 0 {
			page = parsed
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err2 := strconv.Atoi(pp); err2 == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}
	offset := (page - 1) * perPage

	// Get pending URLs
	urls, err := db.ListPendingURLs(r.Context(), h.DB, chatbotID, perPage, offset)
	if err != nil {
		h.logError("list_pending_urls_failed", err)
		api.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	// Get total count
	total, err := db.CountPendingURLs(r.Context(), h.DB, chatbotID)
	if err != nil {
		h.logError("count_pending_urls_failed", err)
		api.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	// Build response
	resp := ListPendingURLsResponse{
		URLs:    make([]PendingURLResponse, 0),
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}
	for _, u := range urls {
		resp.URLs = append(resp.URLs, PendingURLResponse{
			ID:           u.ID,
			URL:          u.URL,
			DiscoveredAt: u.DiscoveredAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	api.WriteJSON(w, http.StatusOK, resp)
}

// ApprovePendingURLs handles POST /api/v1/chatbots/:id/pending-urls/approve
func (h *PendingURLsHandlers) ApprovePendingURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	_, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Parse request body
	var req ApproveRejectRequest
	if err2 := json.NewDecoder(r.Body).Decode(&req); err2 != nil {
		api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	if len(req.URLIDs) == 0 {
		api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no_urls_provided"})
		return
	}

	// Get the pending URLs to create sources
	pendingURLs, err := db.GetPendingURLsByIDs(r.Context(), h.DB, chatbotID, req.URLIDs)
	if err != nil {
		h.logError("get_pending_urls_failed", err)
		api.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	// Create sources for each approved URL and enqueue them for processing
	// Use CreateDiscoveredSource so these URLs won't crawl further (1-level depth)
	sourcesCreated := 0
	for _, pu := range pendingURLs {
		newID, err2 := db.CreateDiscoveredSource(r.Context(), h.DB, chatbotID, pu.URL)
		if err2 == nil {
			sourcesCreated++
			// Enqueue for processing
			if h.Queue != nil {
				h.Queue.Enqueue(newID)
			}
		} else {
			h.logWarn("create_source_failed", map[string]any{"url": pu.URL, "error": err2.Error()})
		}
	}

	// Update status to selected
	approvedCount, err := db.UpdatePendingURLStatus(r.Context(), h.DB, chatbotID, req.URLIDs, "selected")
	if err != nil {
		h.logError("update_pending_urls_status_failed", err)
		api.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	api.WriteJSON(w, http.StatusOK, ApproveResponse{
		ApprovedCount:  approvedCount,
		SourcesCreated: sourcesCreated,
	})
}

// RejectPendingURLs handles POST /api/v1/chatbots/:id/pending-urls/reject
func (h *PendingURLsHandlers) RejectPendingURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	_, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Parse request body
	var req ApproveRejectRequest
	if err2 := json.NewDecoder(r.Body).Decode(&req); err2 != nil {
		api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request_body"})
		return
	}

	if len(req.URLIDs) == 0 {
		api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no_urls_provided"})
		return
	}

	// Update status to rejected
	rejectedCount, err := db.UpdatePendingURLStatus(r.Context(), h.DB, chatbotID, req.URLIDs, "rejected")
	if err != nil {
		h.logError("update_pending_urls_status_failed", err)
		api.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	api.WriteJSON(w, http.StatusOK, RejectResponse{
		RejectedCount: rejectedCount,
	})
}

// ClearPendingURLs handles POST /api/v1/chatbots/:id/pending-urls/clear
func (h *PendingURLsHandlers) ClearPendingURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	_, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Delete all pending URLs
	clearedCount, err := db.DeletePendingURLsByChatbot(r.Context(), h.DB, chatbotID)
	if err != nil {
		h.logError("clear_pending_urls_failed", err)
		api.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	api.WriteJSON(w, http.StatusOK, ClearResponse{
		ClearedCount: clearedCount,
	})
}

func (h *PendingURLsHandlers) logError(event string, err error) {
	if h.Log != nil {
		h.Log.Error(event, map[string]any{"error": err.Error()})
	}
}

func (h *PendingURLsHandlers) logWarn(event string, data map[string]any) {
	if h.Log != nil {
		h.Log.Warn(event, data)
	}
}
