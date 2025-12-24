package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
)

type AdminAuditHandlers struct {
	DB *sql.DB
}

func NewAdminAuditHandlers(database *sql.DB) *AdminAuditHandlers {
	return &AdminAuditHandlers{DB: database}
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

	filter := db.AuditFilter{}
	if adminID := r.URL.Query().Get("admin_user_id"); adminID != "" {
		filter.AdminUserID = &adminID
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = &action
	}
	if targetType := r.URL.Query().Get("target_type"); targetType != "" {
		filter.TargetType = &targetType
	}

	logs, total, err := db.ListAuditLogs(r.Context(), h.DB, filter, limit, offset)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to list audit logs", api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"data":     logs,
		"total":    total,
		"page":     (offset / limit) + 1,
		"per_page": limit,
	})
}
