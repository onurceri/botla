# Task 003: Job Progress API Endpoint

**Priority:** 🔴 Critical (User Experience)  
**Phase:** 1 - Observability Foundation  
**Estimated Time:** 2-3 hours  
**Dependencies:** Task 002 (Job State Table)  

---

## Problem Statement

Frontend has no way to show training progress to users. Currently:
- User clicks "Add Source" and sees nothing until completion
- No feedback on long-running operations (20+ seconds for large files)
- No visibility into failed jobs or why they failed

**User Impact:**
- Users think the app is frozen
- No way to know if processing failed
- Poor user experience

---

## Objective

Create an API endpoint that:
1. Returns training job status and progress for a source
2. Includes current step, progress percentage, and error details
3. Supports polling from frontend

---

## Implementation Details

### Step 1: Create Handler

**File:** `internal/api/handlers/training_job.go` (NEW)

```go
package handlers

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/logger"
)

// TrainingJobHandlers handles training job related requests
type TrainingJobHandlers struct {
	DB  *sql.DB
	Log *logger.Logger
}

// JobStatusResponse is the response for job status endpoint
type JobStatusResponse struct {
	JobID           string                `json:"job_id"`
	SourceID        string                `json:"source_id"`
	Status          models.JobStatus      `json:"status"`
	CurrentStep     *models.TrainingStep  `json:"current_step,omitempty"`
	ProgressPercent int                   `json:"progress_percent"`
	ErrorCode       *string               `json:"error_code,omitempty"`
	ErrorMessage    *string               `json:"error_message,omitempty"`
	FailedStep      *models.TrainingStep  `json:"failed_step,omitempty"`
	StartedAt       *time.Time            `json:"started_at,omitempty"`
	CompletedAt     *time.Time            `json:"completed_at,omitempty"`
}

// GetJobStatus handles GET /api/v1/sources/{id}/job
func (h *TrainingJobHandlers) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	if sourceID == "" {
		api.WriteError(w, http.StatusBadRequest, api.ErrMissingID)
		return
	}

	// Validate source exists and user has access
	source, err := db.GetSourceByID(r.Context(), h.DB, sourceID)
	if err != nil || source == nil {
		api.WriteError(w, http.StatusNotFound, api.ErrNotFound)
		return
	}

	// Check user has access to this source's chatbot
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		api.WriteError(w, http.StatusUnauthorized, api.ErrUnauthorized)
		return
	}

	chatbot, err := db.GetChatbotByID(r.Context(), h.DB, source.ChatbotID)
	if err != nil || chatbot == nil || chatbot.UserID != userID {
		api.WriteError(w, http.StatusForbidden, api.ErrForbidden)
		return
	}

	// Get latest job for this source
	job, err := db.GetJobBySourceID(r.Context(), h.DB, sourceID)
	if err != nil {
		h.Log.Error("get_job_by_source_failed", map[string]any{"error": err.Error()})
		api.WriteError(w, http.StatusInternalServerError, api.ErrInternalServer)
		return
	}

	// If no job exists, return source status as-is
	if job == nil {
		resp := JobStatusResponse{
			SourceID:        sourceID,
			Status:          models.JobStatus(source.Status),
			ProgressPercent: getProgressFromStatus(source.Status),
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

func getProgressFromStatus(status string) int {
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
```

### Step 2: Register Route

**File:** `internal/api/router/routes_sources.go` (MODIFY)

Add the new endpoint:

```go
// In the sources routes section
mux.HandleFunc("GET /api/v1/sources/{id}/job", h.TrainingJobHandlers.GetJobStatus)
```

### Step 3: Create Handler Instance

**File:** `internal/api/router/router.go` (MODIFY)

Initialize the handler:

```go
// In router setup
trainingJobHandlers := &handlers.TrainingJobHandlers{
    DB:  deps.DB,
    Log: log,
}
```

---

## Frontend Integration (Optional for this task)

Frontend can poll this endpoint to show progress:

```typescript
// Example frontend usage
const pollJobStatus = async (sourceId: string) => {
  const response = await fetch(`/api/v1/sources/${sourceId}/job`);
  const job = await response.json();
  
  if (job.status === 'running') {
    // Update progress bar
    setProgress(job.progress_percent);
    setCurrentStep(job.current_step);
    // Poll again in 2 seconds
    setTimeout(() => pollJobStatus(sourceId), 2000);
  } else if (job.status === 'completed') {
    // Refresh source list
    refetchSources();
  } else if (job.status === 'failed') {
    // Show error
    showError(job.error_message);
  }
};
```

---

## Tests to Write

### Unit Tests

**File:** `internal/api/handlers/training_job_test.go` (NEW)

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetJobStatus_NoJob(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	handler := &TrainingJobHandlers{DB: db}
	
	sourceID := createTestSource(t, db)
	
	req := httptest.NewRequest("GET", "/api/v1/sources/"+sourceID+"/job", nil)
	req.SetPathValue("id", sourceID)
	// Add auth context
	req = req.WithContext(withUserID(req.Context(), testUserID))
	
	rec := httptest.NewRecorder()
	handler.GetJobStatus(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	
	var resp JobStatusResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	
	if resp.Status != "pending" {
		t.Errorf("expected pending status, got %s", resp.Status)
	}
}

func TestGetJobStatus_WithJob(t *testing.T) {
	db := setupTestDB(t)
	handler := &TrainingJobHandlers{DB: db}
	
	sourceID := createTestSource(t, db)
	chatbotID := getTestSourceChatbotID(t, db, sourceID)
	
	// Create a job
	job, _ := db.CreateTrainingJob(context.Background(), db, sourceID, chatbotID)
	step := models.StepChunkText
	db.UpdateJobStatus(context.Background(), db, job.ID, models.JobStatusRunning, &step)
	
	req := httptest.NewRequest("GET", "/api/v1/sources/"+sourceID+"/job", nil)
	req.SetPathValue("id", sourceID)
	req = req.WithContext(withUserID(req.Context(), testUserID))
	
	rec := httptest.NewRecorder()
	handler.GetJobStatus(rec, req)
	
	var resp JobStatusResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	
	if resp.Status != models.JobStatusRunning {
		t.Errorf("expected running, got %s", resp.Status)
	}
	if resp.ProgressPercent != 50 { // chunk_text = 50%
		t.Errorf("expected 50%%, got %d", resp.ProgressPercent)
	}
}

func TestGetJobStatus_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	handler := &TrainingJobHandlers{DB: db}
	
	// Create source owned by different user
	sourceID := createTestSourceForUser(t, db, "other-user")
	
	req := httptest.NewRequest("GET", "/api/v1/sources/"+sourceID+"/job", nil)
	req.SetPathValue("id", sourceID)
	req = req.WithContext(withUserID(req.Context(), "current-user"))
	
	rec := httptest.NewRecorder()
	handler.GetJobStatus(rec, req)
	
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
```

### Integration Test

**File:** `internal/integration/training_job_test.go` (NEW)

```go
package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestJobStatusEndpoint_Integration(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	// Create user and chatbot
	token := authToken(t, te.Server.URL, "jobtest@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "Job Test Bot")
	
	// Create a source
	sourceID := createURLSource(t, te.Server.URL, token, chatbotID, "https://example.com")
	
	// Get job status
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	
	var job map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&job)
	
	if job["source_id"] != sourceID {
		t.Errorf("expected source_id %s, got %s", sourceID, job["source_id"])
	}
}
```

---

## Verification Steps

1. **Run unit tests:**
   ```bash
   go test ./internal/api/handlers/... -v -run TestGetJobStatus
   ```

2. **Run integration tests:**
   ```bash
   go test ./internal/integration/... -v -run TestJobStatus
   ```

3. **Manual verification:**
   ```bash
   # Start server
   make be-run
   
   # Login and get token
   TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password"}' | jq -r '.access_token')
   
   # Create source
   SOURCE_ID=$(curl -s -X POST http://localhost:8080/api/v1/chatbots/{bot_id}/sources \
     -H "Authorization: Bearer $TOKEN" \
     -F "source_type=url" \
     -F "source_url=https://example.com" | jq -r '.id')
   
   # Check job status
   curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/sources/$SOURCE_ID/job
   ```

---

## Acceptance Criteria

- [ ] Endpoint returns 200 with job status
- [ ] Returns correct progress percentage based on current step
- [ ] Returns error details for failed jobs
- [ ] Returns 403 for sources user doesn't own
- [ ] Returns 404 for non-existent sources
- [ ] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `internal/api/handlers/training_job.go` | CREATE |
| `internal/api/handlers/training_job_test.go` | CREATE |
| `internal/api/router/routes_sources.go` | MODIFY |
| `internal/api/router/router.go` | MODIFY |
| `internal/integration/training_job_test.go` | CREATE |
