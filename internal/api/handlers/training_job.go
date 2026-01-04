package handlers

import (
	"net/http"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/logger"
)

// TrainingJobHandlers handles training job related requests
type TrainingJobHandlers struct {
	Log              *logger.Logger
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
	Queue            interface {
		Enqueue(jobID string)
	}
	TrainingJobRepo repository.TrainingJobRepository
	SourceRepo      repository.SourceRepository
	ChatbotRepo     repository.ChatbotRepository
}

// JobStatusResponse is the response for job status endpoint
type JobStatusResponse struct {
	JobID           string               `json:"job_id,omitempty"`
	SourceID        string               `json:"source_id"`
	Status          models.JobStatus     `json:"status"`
	CurrentStep     *models.TrainingStep `json:"current_step,omitempty"`
	ProgressPercent int                  `json:"progress_percent"`
	ErrorCode       *string              `json:"error_code,omitempty"`
	ErrorMessage    *string              `json:"error_message,omitempty"`
	FailedStep      *models.TrainingStep `json:"failed_step,omitempty"`
	StartedAt       *time.Time           `json:"started_at,omitempty"`
	CompletedAt     *time.Time           `json:"completed_at,omitempty"`
}

// GetJobStatus handles GET /api/v1/sources/{id}/job
func (h *TrainingJobHandlers) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	// Use the shared getSourceContext to validate access
	source, _, sourceID, ok := getSourceContextWithRepos(w, r, h.SourceRepo, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Get latest job for this source
	job, err := h.TrainingJobRepo.GetBySourceID(r.Context(), sourceID)
	if err != nil {
		h.logError("get_job_by_source_failed", map[string]any{"error": err.Error(), "source_id": sourceID})
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// If no job exists, return source status as-is
	if job == nil {
		resp := JobStatusResponse{
			SourceID:        sourceID,
			Status:          mapSourceStatusToJobStatus(source.Status),
			ProgressPercent: getProgressFromSourceStatus(source.Status),
		}
		api.WriteJSON(w, http.StatusOK, resp)
		return
	}

	resp := JobStatusResponse{
		JobID:           job.ID,
		SourceID:        job.SourceID,
		Status:          job.Status,
		CurrentStep:     job.CurrentStep,
		ProgressPercent: job.ProgressPercent,
		ErrorCode:       job.ErrorCode,
		ErrorMessage:    job.ErrorMessage,
		FailedStep:      job.FailedStep,
		StartedAt:       job.StartedAt,
		CompletedAt:     job.CompletedAt,
	}

	api.WriteJSON(w, http.StatusOK, resp)
}

// RetryJob handles POST /api/v1/sources/{id}/job/retry
func (h *TrainingJobHandlers) RetryJob(w http.ResponseWriter, r *http.Request) {
	// Use the shared getSourceContext to validate access
	_, _, sourceID, ok := getSourceContextWithRepos(w, r, h.SourceRepo, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Get latest job for this source
	job, err := h.TrainingJobRepo.GetBySourceID(r.Context(), sourceID)
	if err != nil || job == nil {
		h.logError("get_job_by_source_failed", map[string]any{"error": err, "source_id": sourceID})
		api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
		return
	}

	// Only allow retry on failed jobs
	if job.Status != models.JobStatusFailed {
		api.WriteErrorCode(w, http.StatusBadRequest, "ERR_JOB_NOT_FAILED")
		return
	}

	// Reset retry count for manual retry
	err = h.TrainingJobRepo.ResetForRetry(r.Context(), job.ID)
	if err != nil {
		h.logError("reset_job_failed", map[string]any{"error": err.Error(), "job_id": job.ID})
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Enqueue for processing
	if h.Queue != nil {
		h.Queue.Enqueue(job.ID)
	}

	api.WriteJSON(w, http.StatusAccepted, map[string]string{
		"job_id":  job.ID,
		"message": "Job queued for retry",
	})
}

// mapSourceStatusToJobStatus converts source status to job status for backward compatibility
func mapSourceStatusToJobStatus(status string) models.JobStatus {
	switch status {
	case "pending":
		return models.JobStatusPending
	case "processing":
		return models.JobStatusRunning
	case "completed":
		return models.JobStatusCompleted
	case "failed":
		return models.JobStatusFailed
	default:
		return models.JobStatusPending
	}
}

// getProgressFromSourceStatus returns progress percentage based on source status
func getProgressFromSourceStatus(status string) int {
	switch status {
	case "pending":
		return 0
	case "processing":
		return 50
	case "completed":
		return 100
	case "failed":
		return 0
	default:
		return 0
	}
}

// logError logs an error if logger is available
func (h *TrainingJobHandlers) logError(event string, data map[string]any) {
	if h.Log != nil {
		h.Log.Error(event, data)
	}
}
