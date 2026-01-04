package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/repository"
)

type AdminAuditHandlers struct {
	AdminRepo repository.AdminRepository
}

func NewAdminAuditHandlers(adminRepo repository.AdminRepository) *AdminAuditHandlers {
	return &AdminAuditHandlers{AdminRepo: adminRepo}
}

// ListAuditLogs returns a paginated list of admin audit logs.
func (h *AdminAuditHandlers) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filter := repository.AuditFilter{}
	if adminID := r.URL.Query().Get("admin_user_id"); adminID != "" {
		filter.AdminUserID = &adminID
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = &action
	}
	if targetType := r.URL.Query().Get("target_type"); targetType != "" {
		filter.TargetType = &targetType
	}
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &t
		}
	}

	logs, total, err := h.AdminRepo.ListAuditLogs(r.Context(), filter, limit, offset)
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
