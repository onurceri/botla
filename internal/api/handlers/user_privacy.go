package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/httputil"
	"github.com/onurceri/botla-app/pkg/middleware"
)

type UserPrivacyHandlers struct {
	DB             *sql.DB
	PrivacyService *services.PrivacyService
	UserRepo       repository.UserRepository
	PrivacyRepo    repository.PrivacyRepository
}

// GetMyConsents returns user's current consent settings
func (h *UserPrivacyHandlers) GetMyConsents(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	consents, err := h.PrivacyRepo.GetUserConsents(r.Context(), userID)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
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

// ListMyPrivacyRequests returns user's own privacy requests with pagination and optional type filter
func (h *UserPrivacyHandlers) ListMyPrivacyRequests(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	requestType := r.URL.Query().Get("type")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	requests, total, err := h.PrivacyRepo.ListPrivacyRequestsByUserID(r.Context(), userID, requestType, limit, offset)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"data":  requests,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateMyConsents updates user's consent settings
func (h *UserPrivacyHandlers) UpdateMyConsents(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	var req struct {
		Marketing       *bool `json:"marketing"`
		Analytics       *bool `json:"analytics"`
		Personalization *bool `json:"personalization"`
		ThirdParty      *bool `json:"third_party"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	ip := httputil.ExtractIP(r)
	userAgent := r.UserAgent()

	// Update each consent if provided
	if req.Marketing != nil {
		if err := h.PrivacyRepo.UpsertConsent(r.Context(), userID, "marketing", *req.Marketing, ip, userAgent); err != nil {
			api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
			return
		}
	}
	if req.Analytics != nil {
		if err := h.PrivacyRepo.UpsertConsent(r.Context(), userID, "analytics", *req.Analytics, ip, userAgent); err != nil {
			api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
			return
		}
	}
	if req.Personalization != nil {
		if err := h.PrivacyRepo.UpsertConsent(r.Context(), userID, "personalization", *req.Personalization, ip, userAgent); err != nil {
			api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
			return
		}
	}
	if req.ThirdParty != nil {
		if err := h.PrivacyRepo.UpsertConsent(r.Context(), userID, "third_party", *req.ThirdParty, ip, userAgent); err != nil {
			api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
			return
		}
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// RequestMyDataExport creates a data export request
func (h *UserPrivacyHandlers) RequestMyDataExport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	user, err := h.UserRepo.GetByID(r.Context(), userID)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if user == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	privacyReq, err := h.PrivacyService.RequestExport(r.Context(), userID, user.Email, "")
	if err != nil {
		if err == services.ErrActiveRequestExists {
			api.WriteErrorCode(w, http.StatusConflict, api.ErrCodeConflict)
			return
		}
		if err == services.ErrRateLimitExceeded {
			api.WriteErrorCode(w, http.StatusTooManyRequests, api.ErrCodeTooManyRequests)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, privacyReq)
}

// RequestDataCorrection creates a data correction request
func (h *UserPrivacyHandlers) RequestDataCorrection(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	if req.Reason == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	user, err := h.UserRepo.GetByID(r.Context(), userID)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if user == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	privacyReq, err := h.PrivacyService.RequestCorrection(r.Context(), userID, user.Email, req.Reason)
	if err != nil {
		if errors.Is(err, services.ErrRateLimitExceeded) {
			api.WriteErrorCode(w, http.StatusTooManyRequests, api.ErrCodeTooManyRequests)
			return
		}
		if errors.Is(err, services.ErrActiveRequestExists) {
			api.WriteErrorCode(w, http.StatusConflict, api.ErrCodeConflict)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, privacyReq)
}

// RequestAccountDeletion creates a deletion request
func (h *UserPrivacyHandlers) RequestAccountDeletion(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	// We need user email
	user, err := h.UserRepo.GetByID(r.Context(), userID)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	if user == nil {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	privacyReq, err := h.PrivacyService.RequestDeletion(r.Context(), userID, user.Email, req.Reason)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, privacyReq)
}

func (h *UserPrivacyHandlers) GetMyPrivacyRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

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
	if req == nil || req.UserID == nil || *req.UserID != userID {
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	api.WriteJSON(w, http.StatusOK, req)
}

func (h *UserPrivacyHandlers) DownloadMyPrivacyExport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

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
	if req == nil || req.UserID == nil || *req.UserID != userID {
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

func (h *UserPrivacyHandlers) DownloadMyDataExport(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

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
	if exp == nil || exp.UserID == nil || *exp.UserID != userID {
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

// DeleteMyPrivacyRequest deletes a user's own privacy request
func (h *UserPrivacyHandlers) DeleteMyPrivacyRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrCodeUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	if err := h.PrivacyService.DeleteMyPrivacyRequest(r.Context(), id, userID); err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
