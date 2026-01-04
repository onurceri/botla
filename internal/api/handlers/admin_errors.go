package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/repository"
)

type AdminErrorHandlers struct {
	AdminRepo repository.AdminRepository
}

func NewAdminErrorHandlers(adminRepo repository.AdminRepository) *AdminErrorHandlers {
	return &AdminErrorHandlers{AdminRepo: adminRepo}
}

// ListErrors returns a paginated list of system errors.
func (h *AdminErrorHandlers) ListErrors(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	severity := r.URL.Query().Get("severity")

	logs, total, err := h.AdminRepo.ListErrorLogs(r.Context(), severity, limit, offset)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"data":     logs,
		"total":    total,
		"page":     (offset / limit) + 1,
		"per_page": limit,
	})
}

// GetError returns full details for a single error entry.
func (h *AdminErrorHandlers) GetError(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	log, err := h.AdminRepo.GetErrorLogByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, log)
}

// GetErrorStats returns summary statistics for recent errors.
func (h *AdminErrorHandlers) GetErrorStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.AdminRepo.GetErrorStats(r.Context())
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, stats)
}
