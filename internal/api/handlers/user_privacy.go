package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/httputil"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type UserPrivacyHandlers struct {
	DB             *sql.DB
	PrivacyService *services.PrivacyService
}

// GetMyConsents returns user's current consent settings
func (h *UserPrivacyHandlers) GetMyConsents(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

	consents, err := db.GetUserConsents(r.Context(), h.DB, userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get consents", api.ErrCodeInternalError)
		return
	}

	// Transform to map for easier frontend consumption with defaults
	consentMap := map[string]bool{
		"marketing":       false,
		"analytics":       false,
		"personalization": false,
		"third_party":     false,
	}
	for _, c := range consents {
		consentMap[c.ConsentType] = c.Granted
	}

	api.WriteJSON(w, http.StatusOK, consentMap)
}

// UpdateMyConsents updates user's consent settings
func (h *UserPrivacyHandlers) UpdateMyConsents(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

	var req struct {
		Marketing       *bool `json:"marketing"`
		Analytics       *bool `json:"analytics"`
		Personalization *bool `json:"personalization"`
		ThirdParty      *bool `json:"third_party"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body", api.ErrCodeBadRequest)
		return
	}

	ip := httputil.ExtractIP(r)
	userAgent := r.UserAgent()

	// Update each consent if provided
	if req.Marketing != nil {
		if err := db.UpsertConsent(r.Context(), h.DB, userID, "marketing", *req.Marketing, ip, userAgent); err != nil {
			api.WriteError(w, http.StatusInternalServerError, "failed to update marketing consent", api.ErrCodeInternalError)
			return
		}
	}
	if req.Analytics != nil {
		if err := db.UpsertConsent(r.Context(), h.DB, userID, "analytics", *req.Analytics, ip, userAgent); err != nil {
			api.WriteError(w, http.StatusInternalServerError, "failed to update analytics consent", api.ErrCodeInternalError)
			return
		}
	}
	if req.Personalization != nil {
		if err := db.UpsertConsent(r.Context(), h.DB, userID, "personalization", *req.Personalization, ip, userAgent); err != nil {
			api.WriteError(w, http.StatusInternalServerError, "failed to update personalization consent", api.ErrCodeInternalError)
			return
		}
	}
	if req.ThirdParty != nil {
		if err := db.UpsertConsent(r.Context(), h.DB, userID, "third_party", *req.ThirdParty, ip, userAgent); err != nil {
			api.WriteError(w, http.StatusInternalServerError, "failed to update third party consent", api.ErrCodeInternalError)
			return
		}
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// RequestMyDataExport creates a data export request
func (h *UserPrivacyHandlers) RequestMyDataExport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

	user, err := db.GetUserByID(r.Context(), h.DB, userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get user", api.ErrCodeInternalError)
		return
	}
	if user == nil {
		api.WriteError(w, http.StatusNotFound, "user not found", api.ErrCodeNotFound)
		return
	}

	privacyReq, err := h.PrivacyService.RequestExport(r.Context(), userID, user.Email, "")
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to request export: "+err.Error(), api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, privacyReq)
}

// RequestDataCorrection creates a data correction request
func (h *UserPrivacyHandlers) RequestDataCorrection(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body", api.ErrCodeBadRequest)
		return
	}

	if req.Reason == "" {
		api.WriteError(w, http.StatusBadRequest, "missing correction details", api.ErrCodeBadRequest)
		return
	}

	user, err := db.GetUserByID(r.Context(), h.DB, userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get user", api.ErrCodeInternalError)
		return
	}
	if user == nil {
		api.WriteError(w, http.StatusNotFound, "user not found", api.ErrCodeNotFound)
		return
	}

	privacyReq, err := h.PrivacyService.RequestCorrection(r.Context(), userID, user.Email, req.Reason)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to request correction: "+err.Error(), api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, privacyReq)
}

// RequestAccountDeletion creates a deletion request
func (h *UserPrivacyHandlers) RequestAccountDeletion(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body", api.ErrCodeBadRequest)
		return
	}

	// We need user email
	user, err := db.GetUserByID(r.Context(), h.DB, userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get user", api.ErrCodeInternalError)
		return
	}
	if user == nil {
		api.WriteError(w, http.StatusNotFound, "user not found", api.ErrCodeNotFound)
		return
	}

	privacyReq, err := h.PrivacyService.RequestDeletion(r.Context(), userID, user.Email, req.Reason)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to request deletion: "+err.Error(), api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, privacyReq)
}

func (h *UserPrivacyHandlers) GetMyPrivacyRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

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
	if req == nil || req.UserID == nil || *req.UserID != userID {
		api.WriteError(w, http.StatusNotFound, "privacy request not found", api.ErrCodeNotFound)
		return
	}

	api.WriteJSON(w, http.StatusOK, req)
}

func (h *UserPrivacyHandlers) DownloadMyPrivacyExport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

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
	if req == nil || req.UserID == nil || *req.UserID != userID {
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

func (h *UserPrivacyHandlers) DownloadMyDataExport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", api.ErrCodeUnauthorized)
		return
	}

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
	if exp == nil || exp.UserID == nil || *exp.UserID != userID {
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
