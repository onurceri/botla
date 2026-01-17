package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)

type PrivacyHandlers struct {
	DB             *sql.DB
	PrivacyService *services.PrivacyService
	AdminService   *services.AdminService
	PrivacyRepo    repository.PrivacyRepository
	Log            *logger.Logger
}

// ListPrivacyRequests returns pending/processed KVKK requests
func (h *PrivacyHandlers) ListPrivacyRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status") // pending, processing, completed, denied
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// DEBUG: Log request details
	if h.Log != nil {
		userID, _ := middleware.UserIDFromContext(r.Context())
		h.Log.DebugCtx(r.Context(), "privacy_list_requests_start", map[string]any{
			"status":  status,
			"page":    page,
			"limit":   limit,
			"offset":  offset,
			"user_id": userID,
		})
	}

	requests, total, err := h.PrivacyRepo.ListPrivacyRequests(r.Context(), status, limit, offset)
	if err != nil {
		// DEBUG: Log the actual error
		if h.Log != nil {
			h.Log.ErrorCtx(r.Context(), "privacy_list_requests_error", map[string]any{
				"error":  err.Error(),
				"status": status,
				"page":   page,
				"limit":  limit,
				"offset": offset,
			})
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// DEBUG: Log success
	if h.Log != nil {
		h.Log.DebugCtx(r.Context(), "privacy_list_requests_success", map[string]any{
			"count":  len(requests),
			"total":  total,
			"status": status,
			"page":   page,
			"limit":  limit,
		})
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"data":     requests,
		"total":    total,
		"page":     page,
		"per_page": limit,
	})
}

// GetPrivacyRequest returns details of a specific request
func (h *PrivacyHandlers) GetPrivacyRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	req, err := h.PrivacyRepo.GetPrivacyRequest(r.Context(), id)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	api.WriteJSON(w, http.StatusOK, req)
}

func (h *PrivacyHandlers) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	req, err := h.PrivacyRepo.GetPrivacyRequest(r.Context(), id)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	if req.RequestType != "export" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}
	if req.Status != "completed" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}
	if req.ExportURL == nil || *req.ExportURL == "" {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if req.ExportExpiresAt != nil && time.Now().After(*req.ExportExpiresAt) {
		api.WriteErrorCode(w, http.StatusGone, api.ErrCodeBadRequest)
		return
	}
	if h.PrivacyService == nil || h.PrivacyService.Storage == nil {
		api.WriteErrorCode(w, http.StatusServiceUnavailable, api.ErrCodeServiceUnavailable)
		return
	}

	signedURL, err := h.PrivacyService.Storage.GetSignedURL(r.Context(), *req.ExportURL, 1*time.Hour)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{
		"url": signedURL,
	})
}

func (h *PrivacyHandlers) DownloadPrivacyExport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	req, err := h.PrivacyRepo.GetPrivacyRequest(r.Context(), id)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	if req.RequestType != "export" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}
	if req.Status != "completed" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}
	if req.ExportURL == nil || *req.ExportURL == "" {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if req.ExportExpiresAt != nil && time.Now().After(*req.ExportExpiresAt) {
		api.WriteErrorCode(w, http.StatusGone, api.ErrCodeBadRequest)
		return
	}
	if h.PrivacyService == nil || h.PrivacyService.Storage == nil {
		api.WriteErrorCode(w, http.StatusServiceUnavailable, api.ErrCodeServiceUnavailable)
		return
	}

	reader, err := h.PrivacyService.Storage.DownloadFile(r.Context(), *req.ExportURL)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	defer func() { _ = reader.Close() }()

	filename := path.Base(*req.ExportURL)
	if filename == "." || filename == "/" || filename == "" {
		filename = "export.json"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, reader)
}

func (h *PrivacyHandlers) DownloadDataExport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	exp, err := h.PrivacyRepo.GetDataExport(r.Context(), id)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if exp == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	if exp.Status != "completed" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}
	if exp.DownloadURL == nil || *exp.DownloadURL == "" {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if exp.ExpiresAt != nil && time.Now().After(*exp.ExpiresAt) {
		api.WriteErrorCode(w, http.StatusGone, api.ErrCodeBadRequest)
		return
	}
	if h.PrivacyService == nil || h.PrivacyService.Storage == nil {
		api.WriteErrorCode(w, http.StatusServiceUnavailable, api.ErrCodeServiceUnavailable)
		return
	}

	reader, err := h.PrivacyService.Storage.DownloadFile(r.Context(), *exp.DownloadURL)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	defer func() { _ = reader.Close() }()

	filename := path.Base(*exp.DownloadURL)
	if filename == "." || filename == "/" || filename == "" {
		filename = "export.json"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, reader)
}

// ProcessPrivacyRequest approves or denies a request
func (h *PrivacyHandlers) ProcessPrivacyRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	var payload struct {
		Action       string `json:"action"` // "approve" or "deny"
		DenialReason string `json:"denial_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	adminID, _ := middleware.UserIDFromContext(r.Context())
	if adminID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	var err error
	// Get request first
	req, errGet := h.PrivacyRepo.GetPrivacyRequest(r.Context(), id)
	if errGet != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	switch payload.Action {
	case "approve":
		switch req.RequestType {
		case "deletion":
			err = h.PrivacyService.ProcessDeletion(r.Context(), id, adminID)
		case "export":
			err = h.PrivacyService.ProcessExportRequest(r.Context(), id, adminID)
		default:
			// Just mark completed for other types
			err = h.PrivacyRepo.UpdatePrivacyRequestStatus(r.Context(), id, "completed", adminID, nil)
		}
	case "deny":
		err = h.PrivacyRepo.UpdatePrivacyRequestStatus(r.Context(), id, "denied", adminID, &payload.DenialReason)
	default:
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Log action
	details := map[string]any{
		"request_id":   id,
		"action":       payload.Action,
		"request_type": req.RequestType,
	}
	if payload.Action == "deny" {
		details["reason"] = payload.DenialReason
	}
	_ = h.AdminService.LogAction(r.Context(), adminID, "process_privacy_request", "privacy_request", &id, details, r)

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GenerateUserExport creates a data export for a user (admin-initiated)
func (h *PrivacyHandlers) GenerateUserExport(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	adminID, _ := middleware.UserIDFromContext(r.Context())
	export, err := h.PrivacyService.ExportUserData(r.Context(), userID, adminID)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Log action
	details := map[string]any{
		"user_id":   userID,
		"export_id": export.ID,
	}
	_ = h.AdminService.LogAction(r.Context(), adminID, "generate_user_export", "user", &userID, details, r)

	api.WriteJSON(w, http.StatusOK, export)
}
