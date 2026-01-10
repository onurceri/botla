package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/workers"
	"github.com/onurceri/botla-app/pkg/logger"
)

func TestFeedbackHandler_FirstThumbsUp(t *testing.T) {
	// Setup
	mockRepo := repository.NewMockAnalyticsRepo()
	log := logger.New("debug")
	pool := workers.NewWorkerPool(log, 1)
	defer pool.Shutdown(time.Second)

	h := &ChatHandlers{
		AnalyticsRepo: mockRepo,
		WorkerPool:    pool,
		Logger:        log,
	}

	chatbotID := "bot-123"
	msgID := "msg-456"

	// Mock UpdateMessageFeedback to return as if it's the first time (old value is nil)
	mockRepo.UpdateMessageFeedbackFunc = func(ctx context.Context, messageID string, thumbsUp bool) (string, *bool, error) {
		if messageID != msgID {
			t.Errorf("expected messageID %s, got %s", msgID, messageID)
		}
		if !thumbsUp {
			t.Error("expected thumbsUp to be true")
		}
		return chatbotID, nil, nil
	}

	// Mock IncrementFeedback to verify oldState is nil
	incrementCalled := make(chan bool)
	mockRepo.IncrementFeedbackFunc = func(ctx context.Context, cID string, oldState *bool, newState bool) error {
		defer close(incrementCalled)
		if cID != chatbotID {
			t.Errorf("expected chatbotID %s, got %s", chatbotID, cID)
		}
		if oldState != nil {
			t.Errorf("expected oldState to be nil, got %v", *oldState)
		}
		if !newState {
			t.Error("expected newState to be true")
		}
		return nil
	}

	// Request
	reqBody := map[string]any{"thumbs_up": true}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/"+msgID+"/feedback", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()

	h.FeedbackHandler(rr, req)

	// Assert Response
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Wait for async worker
	select {
	case <-incrementCalled:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for IncrementFeedback")
	}
}

func TestFeedbackHandler_ChangeThumbsDownToUp(t *testing.T) {
	// Setup
	mockRepo := repository.NewMockAnalyticsRepo()
	log := logger.New("debug")
	pool := workers.NewWorkerPool(log, 1)
	defer pool.Shutdown(time.Second)

	h := &ChatHandlers{
		AnalyticsRepo: mockRepo,
		WorkerPool:    pool,
		Logger:        log,
	}

	chatbotID := "bot-123"
	msgID := "msg-456"

	// Mock UpdateMessageFeedback to return as if it was previously thumbs down (false)
	mockRepo.UpdateMessageFeedbackFunc = func(ctx context.Context, messageID string, thumbsUp bool) (string, *bool, error) {
		oldVal := false
		return chatbotID, &oldVal, nil
	}

	// Mock IncrementFeedback
	incrementCalled := make(chan bool)
	mockRepo.IncrementFeedbackFunc = func(ctx context.Context, cID string, oldState *bool, newState bool) error {
		defer close(incrementCalled)
		if oldState == nil {
			t.Error("expected oldState to be non-nil")
		} else if *oldState != false {
			t.Errorf("expected oldState to be false, got %v", *oldState)
		}
		if !newState {
			t.Error("expected newState to be true")
		}
		return nil
	}

	// Request
	reqBody := map[string]any{"thumbs_up": true}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/"+msgID+"/feedback", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()

	h.FeedbackHandler(rr, req)

	// Assert Response
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Wait for async worker
	select {
	case <-incrementCalled:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for IncrementFeedback")
	}
}
