package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type PrivacyHandlers struct {
	DB             *sql.DB
	PrivacyService *services.PrivacyService
	AdminService   *services.AdminService
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

	requests, total, err := db.ListPrivacyRequests(r.Context(), h.DB, status, limit, offset)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to list privacy requests", api.ErrCodeInternalError)
		return
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
		api.WriteError(w, http.StatusBadRequest, "missing request id", api.ErrCodeBadRequest)
		return
	}

	req, err := db.GetPrivacyRequest(r.Context(), h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get privacy request", api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteError(w, http.StatusNotFound, "privacy request not found", api.ErrCodeNotFound)
		return
	}

	api.WriteJSON(w, http.StatusOK, req)
}

func (h *PrivacyHandlers) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "missing request id", api.ErrCodeBadRequest)
		return
	}

	req, err := db.GetPrivacyRequest(r.Context(), h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get privacy request", api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteError(w, http.StatusNotFound, "privacy request not found", api.ErrCodeNotFound)
		return
	}

	if req.RequestType != "export" {
		api.WriteError(w, http.StatusBadRequest, "request is not an export request", api.ErrCodeBadRequest)
		return
	}
	if req.Status != "completed" {
		api.WriteError(w, http.StatusBadRequest, "export is not ready", api.ErrCodeBadRequest)
		return
	}
	if req.ExportURL == nil || *req.ExportURL == "" {
		api.WriteError(w, http.StatusInternalServerError, "export url missing", api.ErrCodeInternalError)
		return
	}
	if req.ExportExpiresAt != nil && time.Now().After(*req.ExportExpiresAt) {
		api.WriteError(w, http.StatusGone, "export expired", api.ErrCodeBadRequest)
		return
	}
	if h.PrivacyService == nil || h.PrivacyService.Storage == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "storage not configured", api.ErrCodeServiceUnavailable)
		return
	}

	signedURL, err := h.PrivacyService.Storage.GetSignedURL(r.Context(), *req.ExportURL, 1*time.Hour)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to generate signed url", api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{
		"url": signedURL,
	})
}

func (h *PrivacyHandlers) DownloadPrivacyExport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "missing request id", api.ErrCodeBadRequest)
		return
	}

	req, err := db.GetPrivacyRequest(r.Context(), h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get privacy request", api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteError(w, http.StatusNotFound, "privacy request not found", api.ErrCodeNotFound)
		return
	}

	if req.RequestType != "export" {
		api.WriteError(w, http.StatusBadRequest, "request is not an export request", api.ErrCodeBadRequest)
		return
	}
	if req.Status != "completed" {
		api.WriteError(w, http.StatusBadRequest, "export is not ready", api.ErrCodeBadRequest)
		return
	}
	if req.ExportURL == nil || *req.ExportURL == "" {
		api.WriteError(w, http.StatusInternalServerError, "export url missing", api.ErrCodeInternalError)
		return
	}
	if req.ExportExpiresAt != nil && time.Now().After(*req.ExportExpiresAt) {
		api.WriteError(w, http.StatusGone, "export expired", api.ErrCodeBadRequest)
		return
	}
	if h.PrivacyService == nil || h.PrivacyService.Storage == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "storage not configured", api.ErrCodeServiceUnavailable)
		return
	}

	reader, err := h.PrivacyService.Storage.DownloadFile(r.Context(), *req.ExportURL)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to download export", api.ErrCodeInternalError)
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
		api.WriteError(w, http.StatusBadRequest, "missing export id", api.ErrCodeBadRequest)
		return
	}

	exp, err := db.GetDataExport(r.Context(), h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get export", api.ErrCodeInternalError)
		return
	}
	if exp == nil {
		api.WriteError(w, http.StatusNotFound, "export not found", api.ErrCodeNotFound)
		return
	}

	if exp.Status != "completed" {
		api.WriteError(w, http.StatusBadRequest, "export is not ready", api.ErrCodeBadRequest)
		return
	}
	if exp.DownloadURL == nil || *exp.DownloadURL == "" {
		api.WriteError(w, http.StatusInternalServerError, "download url missing", api.ErrCodeInternalError)
		return
	}
	if exp.ExpiresAt != nil && time.Now().After(*exp.ExpiresAt) {
		api.WriteError(w, http.StatusGone, "export expired", api.ErrCodeBadRequest)
		return
	}
	if h.PrivacyService == nil || h.PrivacyService.Storage == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "storage not configured", api.ErrCodeServiceUnavailable)
		return
	}

	reader, err := h.PrivacyService.Storage.DownloadFile(r.Context(), *exp.DownloadURL)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to download export", api.ErrCodeInternalError)
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
		api.WriteError(w, http.StatusBadRequest, "missing request id", api.ErrCodeBadRequest)
		return
	}

	var payload struct {
		Action       string `json:"action"` // "approve" or "deny"
		DenialReason string `json:"denial_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body", api.ErrCodeBadRequest)
		return
	}

	adminID, _ := middleware.UserIDFromContext(r.Context())
	if adminID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

	var err error
	// Get request first
	req, errGet := db.GetPrivacyRequest(r.Context(), h.DB, id)
	if errGet != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get request", api.ErrCodeInternalError)
		return
	}
	if req == nil {
		api.WriteError(w, http.StatusNotFound, "request not found", api.ErrCodeNotFound)
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
			err = db.UpdatePrivacyRequestStatus(r.Context(), h.DB, id, "completed", adminID, nil)
		}
	case "deny":
		err = db.UpdatePrivacyRequestStatus(r.Context(), h.DB, id, "denied", adminID, &payload.DenialReason)
	default:
		api.WriteError(w, http.StatusBadRequest, "invalid action", api.ErrCodeBadRequest)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to process request: "+err.Error(), api.ErrCodeInternalError)
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
		api.WriteError(w, http.StatusBadRequest, "missing user id", api.ErrCodeBadRequest)
		return
	}

	adminID, _ := middleware.UserIDFromContext(r.Context())
	export, err := h.PrivacyService.ExportUserData(r.Context(), userID, adminID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to generate export: "+err.Error(), api.ErrCodeInternalError)
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
