package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type AdminQueueHandlers struct {
	DB           *sql.DB
	AdminService *services.AdminService
}

func NewAdminQueueHandlers(database *sql.DB, adminSvc *services.AdminService) *AdminQueueHandlers {
	return &AdminQueueHandlers{
		DB:           database,
		AdminService: adminSvc,
	}
}

// GetQueues returns statistics for processing queues.
func (h *AdminQueueHandlers) GetQueues(w http.ResponseWriter, r *http.Request) {
	stats, err := db.GetQueueStats(r.Context(), h.DB)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	api.WriteJSON(w, http.StatusOK, stats)
}

// GetStuckJobs returns jobs that have been processing for too long.
func (h *AdminQueueHandlers) GetStuckJobs(w http.ResponseWriter, r *http.Request) {
	threshold := 30 * time.Minute
	if t := r.URL.Query().Get("threshold"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			threshold = d
		}
	}

	jobs, err := db.GetStuckJobs(r.Context(), h.DB, threshold)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}
	api.WriteJSON(w, http.StatusOK, jobs)
}

// RetryJob resets a stuck or failed job to pending status.
func (h *AdminQueueHandlers) RetryJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	// Currently retrying means resetting data_source status to pending
	_, err := h.DB.ExecContext(r.Context(), `
		UPDATE data_sources 
		SET status = 'pending', error_message = NULL, last_refreshed_at = NOW()
		WHERE id = $1 AND status IN ('processing', 'error')
	`, id)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Log the action
	adminID, _ := middleware.UserIDFromContext(r.Context())
	_ = h.AdminService.LogAction(r.Context(), adminID, "retry_job", "data_source", &id, nil, r)

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteJob removes a job from the database.
func (h *AdminQueueHandlers) DeleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	_, err := h.DB.ExecContext(r.Context(), `DELETE FROM data_sources WHERE id = $1`, id)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Log the action
	adminID, _ := middleware.UserIDFromContext(r.Context())
	_ = h.AdminService.LogAction(r.Context(), adminID, "delete_job", "data_source", &id, nil, r)

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
