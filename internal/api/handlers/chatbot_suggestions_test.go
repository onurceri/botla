package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler(t *testing.T) (*SuggestionsHandlers, *sql.DB) {
	// Use testdb.OpenParallelTestDB to properly run migrations and create isolated schema
	dbConn := testdb.OpenParallelTestDB(t)

	log := logger.New("info")

	h := &SuggestionsHandlers{
		Log:               log,
		WorkspaceService:  nil,
		OrgService:        nil,
		SuggestionJobRepo: repository.NewPostgresSuggestionJobRepo(dbConn),
		ChatbotRepo:       repository.NewPostgresChatbotRepo(dbConn),
	}

	return h, dbConn
}

func TestRegenerateSuggestions_Success(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID
	userID := result.User.ID

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	req.SetPathValue("id", chatbotID)
	ctx = context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusAccepted, rec.Code)

	var response RegenerateResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.NotEmpty(t, response.JobID)

	job, err := h.SuggestionJobRepo.GetByID(ctx, response.JobID)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, chatbotID, job.ChatbotID)
	// Job status may be pending or running due to background goroutine timing
	assert.True(t, job.Status == models.SuggestionJobStatusPending || job.Status == models.SuggestionJobStatusRunning,
		"expected status pending or running, got %s", job.Status)
}

func TestRegenerateSuggestions_MethodNotAllowed(t *testing.T) {
	h, _ := setupTestHandler(t)

	// Use a hardcoded UUID since we only test the method check (doesn't need real entity)
	chatbotID := uuid.NewString()

	// GET request should be rejected immediately
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestGetSuggestionJobStatus_Success(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID
	userID := result.User.ID

	job, err := h.SuggestionJobRepo.Create(ctx, chatbotID)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	req.SetPathValue("id", chatbotID)
	ctx = context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuggestionJobStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, job.ID, response.JobID)
	assert.Equal(t, models.SuggestionJobStatusPending.String(), response.Status)
}

func TestGetSuggestionJobStatus_NotFound(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	// Create a chatbot but don't create any suggestion jobs
	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID
	userID := result.User.ID

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	req.SetPathValue("id", chatbotID)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetSuggestionJobStatus_MethodNotAllowed(t *testing.T) {
	h, _ := setupTestHandler(t)

	// Use a hardcoded UUID since we only test the method check (doesn't need real entity)
	chatbotID := uuid.NewString()

	// POST request should be rejected immediately
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestGetSuggestionJobStatus_WithSuggestions(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID
	userID := result.User.ID

	job, err := h.SuggestionJobRepo.Create(ctx, chatbotID)
	assert.NoError(t, err)

	suggestions := []string{"Q1", "Q2", "Q3"}
	err = h.SuggestionJobRepo.Complete(ctx, job.ID, suggestions)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	req.SetPathValue("id", chatbotID)
	ctx = context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuggestionJobStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, models.SuggestionJobStatusCompleted.String(), response.Status)
	assert.Len(t, response.SuggestedQuestions, 1)
	assert.Len(t, response.SuggestedQuestions[0], 3)
}

func TestGetSuggestionJobStatus_WithError(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID
	userID := result.User.ID

	job, err := h.SuggestionJobRepo.Create(ctx, chatbotID)
	assert.NoError(t, err)

	errMsg := "test error"
	err = h.SuggestionJobRepo.Fail(ctx, job.ID, errMsg)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	req.SetPathValue("id", chatbotID)
	ctx = context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuggestionJobStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, models.SuggestionJobStatusFailed.String(), response.Status)
	assert.NotNil(t, response.ErrorMessage)
	assert.Equal(t, errMsg, *response.ErrorMessage)
}

func TestRegenerateSuggestions_InvalidChatbotID(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	// Need a valid user ID but an invalid chatbot ID
	result := testdb.CreateChatbot(t, dbConn)
	userID := result.User.ID

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/invalid-uuid/suggestions/regenerate", nil)
	req.SetPathValue("id", "invalid-uuid")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetSuggestionJobStatus_InvalidChatbotID(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	// Need a valid user ID but an invalid chatbot ID
	result := testdb.CreateChatbot(t, dbConn)
	userID := result.User.ID

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/invalid-uuid/suggestions/status", nil)
	req.SetPathValue("id", "invalid-uuid")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegenerateSuggestions_ReturnsJSON(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID
	userID := result.User.ID

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	req.SetPathValue("id", chatbotID)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req.WithContext(ctx))

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	_, ok := response["job_id"]
	assert.True(t, ok, "response should contain job_id")
}

func TestGetSuggestionJobStatus_ReturnsJSON(t *testing.T) {
	h, dbConn := setupTestHandler(t)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID
	userID := result.User.ID

	_, err := h.SuggestionJobRepo.Create(ctx, chatbotID)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	req.SetPathValue("id", chatbotID)
	ctx = context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req.WithContext(ctx))

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	_, ok := response["job_id"]
	assert.True(t, ok, "response should contain job_id")
	_, ok = response["status"]
	assert.True(t, ok, "response should contain status")
}

func TestRegenerateSuggestions_Unauthorized(t *testing.T) {
	h, _ := setupTestHandler(t)

	chatbotID := uuid.NewString()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	req.SetPathValue("id", chatbotID)
	rec := httptest.NewRecorder()

	// No user ID in context - should get 401
	h.RegenerateSuggestions(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetSuggestionJobStatus_Unauthorized(t *testing.T) {
	h, _ := setupTestHandler(t)

	chatbotID := uuid.NewString()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	req.SetPathValue("id", chatbotID)
	rec := httptest.NewRecorder()

	// No user ID in context - should get 401
	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
