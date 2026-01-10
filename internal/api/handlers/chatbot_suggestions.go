package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-app/internal/processing"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/workers"
	"github.com/onurceri/botla-app/pkg/logger"
)

type SuggestionsHandlers struct {
	Log               *logger.Logger
	WorkspaceService  *services.WorkspaceService
	OrgService        *services.OrganizationService
	SuggestionJobRepo repository.SuggestionJobRepository
	ChatbotRepo       repository.ChatbotRepository
	SourceRepo        repository.SourceRepository
	WorkerPool        *workers.WorkerPool
}

type RegenerateResponse struct {
	JobID string `json:"job_id"`
}

type SuggestionJobStatusResponse struct {
	JobID              string     `json:"job_id"`
	Status             string     `json:"status"`
	SuggestedQuestions [][]string `json:"suggested_questions,omitempty"`
	ErrorMessage       *string    `json:"error_message,omitempty"`
}

func (h *SuggestionsHandlers) RegenerateSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, chatbotID, ok := getChatbotContextWithRepo(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	ctx := context.Background()

	job, err := h.SuggestionJobRepo.Create(ctx, chatbotID)
	if err != nil {
		h.Log.Error("create_suggestion_job_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		http.Error(w, "Failed to create suggestion job", http.StatusInternalServerError)
		return
	}

	// Submit job to worker pool for async processing
	if h.WorkerPool != nil {
		jobID := job.ID
		cID := chatbotID
		h.WorkerPool.Submit(func(ctx context.Context) {
			processing.ReAggregateSuggestionsForChatbotWithJob(ctx, h.ChatbotRepo, h.SourceRepo, h.SuggestionJobRepo, cID, jobID, h.Log)
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(RegenerateResponse{JobID: job.ID}); err != nil {
		h.Log.Error("encode_response_failed", map[string]any{"error": err.Error()})
	}
}

func (h *SuggestionsHandlers) GetSuggestionJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, chatbotID, ok := getChatbotContextWithRepo(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	ctx := context.Background()

	job, err := h.SuggestionJobRepo.GetLatestForChatbot(ctx, chatbotID)
	if err != nil {
		h.Log.Error("get_suggestion_job_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		http.Error(w, "Failed to get suggestion job", http.StatusInternalServerError)
		return
	}

	if job == nil {
		http.Error(w, "No suggestion job found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := SuggestionJobStatusResponse{
		JobID:        job.ID,
		Status:       job.Status.String(),
		ErrorMessage: job.ErrorMessage,
	}
	// Only include suggested_questions if there are actual questions
	if len(job.SuggestedQuestions) > 0 {
		response.SuggestedQuestions = [][]string{job.SuggestedQuestions}
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Log.Error("encode_response_failed", map[string]any{"error": err.Error()})
	}
}
