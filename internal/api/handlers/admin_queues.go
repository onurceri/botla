package handlers

import (
	"net/http"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/middleware"
)

type AdminQueueHandlers struct {
	AdminService *services.AdminService
	QueueRepo    repository.QueueRepository
	SourceRepo   repository.SourceRepository
}

func NewAdminQueueHandlers(adminSvc *services.AdminService, queueRepo repository.QueueRepository, sourceRepo repository.SourceRepository) *AdminQueueHandlers {
	return &AdminQueueHandlers{
		AdminService: adminSvc,
		QueueRepo:    queueRepo,
		SourceRepo:   sourceRepo,
	}
}

// GetQueues returns statistics for processing queues.
func (h *AdminQueueHandlers) GetQueues(w http.ResponseWriter, r *http.Request) {
	stats, err := h.QueueRepo.GetQueueStats(r.Context())
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

	jobs, err := h.QueueRepo.GetStuckJobs(r.Context(), threshold)
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

	err := h.SourceRepo.UpdateForRefresh(r.Context(), id)
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

	err := h.SourceRepo.Delete(r.Context(), id)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Log the action
	adminID, _ := middleware.UserIDFromContext(r.Context())
	_ = h.AdminService.LogAction(r.Context(), adminID, "delete_job", "data_source", &id, nil, r)

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
