package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/logger"
)

type SuggestionsHandlers struct {
	DB               *sql.DB
	Log              *logger.Logger
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
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

	_, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	ctx := context.Background()

	job, err := db.CreateSuggestionJob(ctx, h.DB, chatbotID)
	if err != nil {
		h.Log.Error("create_suggestion_job_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		http.Error(w, "Failed to create suggestion job", http.StatusInternalServerError)
		return
	}

	go processing.ReAggregateSuggestionsForChatbotWithJob(context.Background(), h.DB, chatbotID, job.ID, h.Log)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(RegenerateResponse{JobID: job.ID})
}

func (h *SuggestionsHandlers) GetSuggestionJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	ctx := context.Background()

	job, err := db.GetLatestSuggestionJobForChatbot(ctx, h.DB, chatbotID)
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
	json.NewEncoder(w).Encode(SuggestionJobStatusResponse{
		JobID:              job.ID,
		Status:             job.Status.String(),
		SuggestedQuestions: [][]string{job.SuggestedQuestions},
		ErrorMessage:       job.ErrorMessage,
	})
}
